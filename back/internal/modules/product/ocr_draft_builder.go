package product

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidOCRRequest = errors.New("invalid ocr request")

type OCRClient interface {
	ExtractText(ctx context.Context, imageBase64, lang, region string) (string, error)
}

// OCRDraftBuilder builds a structured, editable OCR draft suitable for UI validation.
//
// Pipeline:
// OCR -> text cleaning -> block detection -> parsing -> fuzzy matching (DB) -> normalization
// -> confidence scoring -> editable output DTO
type OCRDraftBuilder struct {
	ocr          OCRClient
	baseProducts BaseProductRepository
}

func NewOCRDraftBuilder(ocr OCRClient, baseProducts BaseProductRepository) *OCRDraftBuilder {
	if baseProducts == nil {
		baseProducts = NewNoopBaseProductRepository()
	}
	return &OCRDraftBuilder{ocr: ocr, baseProducts: baseProducts}
}

func (b *OCRDraftBuilder) Build(ctx context.Context, req OCRDraftRequest, ocrMode string) (OCRDraftDTO, error) {
	images := make([]string, 0, 4)
	for _, img := range req.Images {
		img = strings.TrimSpace(img)
		if img != "" {
			images = append(images, img)
		}
	}
	if strings.TrimSpace(req.Image) != "" {
		images = append(images, strings.TrimSpace(req.Image))
	}
	if len(images) == 0 {
		return OCRDraftDTO{}, fmt.Errorf("%w: ocr image required", ErrInvalidOCRRequest)
	}
	if b == nil || b.ocr == nil {
		return OCRDraftDTO{}, fmt.Errorf("%w: ocr engine not configured", ErrInvalidOCRRequest)
	}

	perPhoto := make([]OCRDraft, 0, len(images))
	for _, img := range images {
		text, err := b.ocr.ExtractText(ctx, img, req.Lang, req.Region)
		if err != nil {
			return OCRDraftDTO{}, err
		}
		cleaned := cleanOCRText(text)
		blocks := detectOCRBlocks(cleaned)
		ocrQuality := estimateOCRQuality(cleaned)

		ings := parseIngredients(blocks.Ingredients)
		for i := range ings {
			ings[i].ClientID = stableIngredientClientID(ings[i].Name)
			b.matchIngredient(ctx, &ings[i])
			ings[i].Confidence = ingredientConfidence(ocrQuality, ings[i])
		}

		nut := parseNutrition(blocks.Nutrition)
		scoreNutritionFields(ocrQuality, &nut)

		d := OCRDraft{
			OCRQuality:    ocrQuality,
			Ingredients:   ings,
			Nutrition:     nut,
			MissingFields: missingFields(ings, nut),
		}
		d.OverallConfidence = overallConfidence(d)
		perPhoto = append(perPhoto, d)
	}

	merged := mergeDrafts(perPhoto)
	merged.MissingFields = missingFields(merged.Ingredients, merged.Nutrition)
	merged.OverallConfidence = overallConfidence(merged)

	return merged.toDTO(ocrMode), nil
}

func (d OCRDraft) toDTO(ocrMode string) OCRDraftDTO {
	dto := OCRDraftDTO{
		OCRMode:           ocrMode,
		OCRQuality:        clamp01(d.OCRQuality),
		OverallConfidence: clamp01(d.OverallConfidence),
		MissingFields:     append([]string(nil), d.MissingFields...),
	}
	dto.Ingredients = make([]OCRIngredientDTO, 0, len(d.Ingredients))
	for _, ing := range d.Ingredients {
		dto.Ingredients = append(dto.Ingredients, OCRIngredientDTO{
			ClientID:         ing.ClientID,
			Raw:              ing.Raw,
			Name:             ing.Name,
			MatchedProductID: ing.MatchedProductID,
			MatchName:        strings.TrimSpace(ing.MatchName),
			MatchScore:       clamp01(ing.MatchScore),
			Confidence:       clamp01(ing.Confidence),
			IsVerified:       ing.IsVerified,
		})
	}
	dto.Nutrition = OCRNutritionDTO{
		EnergyUnit: d.Nutrition.EnergyUnit,
		MassUnit:   d.Nutrition.MassUnit,
		Calories:   nutritionFieldToDTO(d.Nutrition.Calories),
		Protein:    nutritionFieldToDTO(d.Nutrition.Protein),
		Fat:        nutritionFieldToDTO(d.Nutrition.Fat),
		Carbs:      nutritionFieldToDTO(d.Nutrition.Carbs),
	}
	if len(d.Conflicts) > 0 {
		dto.Conflicts = make([]OCRConflictDTO, 0, len(d.Conflicts))
		for _, c := range d.Conflicts {
			dto.Conflicts = append(dto.Conflicts, OCRConflictDTO{Field: c.Field, Note: c.Note})
		}
	}
	return dto
}

func nutritionFieldToDTO(f OCRNutritionField) OCRNutritionFieldDTO {
	return OCRNutritionFieldDTO{
		Value:       f.Value,
		Confidence:  clamp01(f.Confidence),
		IsEstimated: f.IsEstimated,
		IsVerified:  f.IsVerified,
	}
}

func stableIngredientClientID(name string) string {
	h := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(name))))
	return "ing_" + hex.EncodeToString(h[:6])
}

func (b *OCRDraftBuilder) matchIngredient(ctx context.Context, ing *OCRIngredient) {
	if ing == nil || b == nil || b.baseProducts == nil {
		return
	}
	q := normalizeForMatch(ing.Name)
	if q == "" {
		return
	}
	candidates, err := b.baseProducts.Search(ctx, q, 12)
	if err != nil || len(candidates) == 0 {
		return
	}

	best, score := bestBaseProductMatch(q, candidates)
	if best == nil {
		return
	}
	if score < 0.65 {
		ing.MatchScore = score
		return
	}
	ing.MatchedProductID = &best.ID
	ing.MatchName = best.Name
	ing.MatchScore = score
}

func bestBaseProductMatch(query string, candidates []BaseProduct) (*BaseProduct, float64) {
	query = normalizeForMatch(query)
	if query == "" || len(candidates) == 0 {
		return nil, 0
	}
	type scored struct {
		item  BaseProduct
		score float64
	}
	scoredItems := make([]scored, 0, len(candidates))
	for _, c := range candidates {
		name := normalizeForMatch(c.Name)
		if name == "" {
			continue
		}
		score := trigramJaccard(query, name)
		if strings.HasPrefix(name, query) || strings.HasPrefix(query, name) {
			score = math.Max(score, 0.85)
		}
		scoredItems = append(scoredItems, scored{item: c, score: score})
	}
	if len(scoredItems) == 0 {
		return nil, 0
	}
	sort.Slice(scoredItems, func(i, j int) bool {
		if scoredItems[i].score == scoredItems[j].score {
			return len(scoredItems[i].item.Name) < len(scoredItems[j].item.Name)
		}
		return scoredItems[i].score > scoredItems[j].score
	})
	best := scoredItems[0]
	return &best.item, clamp01(best.score)
}

func trigramJaccard(a, b string) float64 {
	a = normalizeForMatch(a)
	b = normalizeForMatch(b)
	if a == "" || b == "" {
		return 0
	}
	ta := trigrams(a)
	tb := trigrams(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	inter := 0
	union := map[string]struct{}{}
	for t := range ta {
		union[t] = struct{}{}
	}
	for t := range tb {
		if _, ok := ta[t]; ok {
			inter++
		}
		union[t] = struct{}{}
	}
	return float64(inter) / float64(len(union))
}

func trigrams(s string) map[string]struct{} {
	r := []rune(s)
	if len(r) < 3 {
		return map[string]struct{}{s: {}}
	}
	out := make(map[string]struct{}, len(r))
	for i := 0; i+3 <= len(r); i++ {
		out[string(r[i:i+3])] = struct{}{}
	}
	return out
}

func cleanOCRText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.TrimSpace(s)
	s = strings.Map(func(r rune) rune {
		switch r {
		case '•', '·', '◦', '▪', '◊':
			return '\n'
		default:
			return r
		}
	}, s)
	return s
}

type ocrBlocks struct {
	Ingredients string
	Nutrition   string
}

func detectOCRBlocks(text string) ocrBlocks {
	lines := strings.Split(text, "\n")

	ingIdx := -1
	nutIdx := -1
	for i, line := range lines {
		if ingIdx == -1 && isIngredientsHeader(line) {
			ingIdx = i
		}
		if nutIdx == -1 && isNutritionHeader(line) {
			nutIdx = i
		}
	}

	var ingredients string
	var nutrition string

	if ingIdx >= 0 {
		ingredients = strings.TrimSpace(afterColon(lines[ingIdx]))
		if ingredients == "" && reNum.MatchString(lines[ingIdx]) {
			ingredients = strings.TrimSpace(lines[ingIdx])
		}
		start := ingIdx + 1
		end := len(lines)
		if nutIdx >= 0 && nutIdx > ingIdx {
			end = nutIdx
			if nutrition == "" {
				nutrition = strings.TrimSpace(afterColon(lines[nutIdx]))
				if nutrition == "" && reNum.MatchString(lines[nutIdx]) {
					nutrition = strings.TrimSpace(lines[nutIdx])
				}
				if nutIdx+1 < len(lines) {
					nutrition = strings.TrimSpace(strings.TrimSpace(nutrition + "\n" + strings.Join(lines[nutIdx+1:], "\n")))
				}
			}
		}
		if start < end {
			body := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
			if ingredients == "" {
				ingredients = body
			} else if body != "" {
				ingredients = strings.TrimSpace(ingredients + "\n" + body)
			}
		}
	} else {
		ingredients = text
	}

	if nutrition == "" && nutIdx >= 0 {
		nutrition = strings.TrimSpace(afterColon(lines[nutIdx]))
		if nutrition == "" && reNum.MatchString(lines[nutIdx]) {
			nutrition = strings.TrimSpace(lines[nutIdx])
		}
		if nutIdx+1 < len(lines) {
			body := strings.TrimSpace(strings.Join(lines[nutIdx+1:], "\n"))
			if nutrition == "" {
				nutrition = body
			} else if body != "" {
				nutrition = strings.TrimSpace(nutrition + "\n" + body)
			}
		}
	}

	return ocrBlocks{Ingredients: ingredients, Nutrition: nutrition}
}

func afterColon(line string) string {
	line = strings.TrimSpace(line)
	if idx := strings.IndexRune(line, ':'); idx >= 0 && idx+1 < len(line) {
		return strings.TrimSpace(line[idx+1:])
	}
	return ""
}

func isIngredientsHeader(line string) bool {
	s := strings.ToLower(strings.TrimSpace(line))
	if s == "" {
		return false
	}
	return strings.Contains(s, "ingredients") ||
		strings.Contains(s, "ingredient") ||
		strings.Contains(s, "ингредиент") ||
		strings.Contains(s, "состав") ||
		strings.Contains(s, "coctab") // Cyrillic->Latin OCR confusion for "СОСТАВ"
}

func isNutritionHeader(line string) bool {
	s := strings.ToLower(strings.TrimSpace(line))
	if s == "" {
		return false
	}
	return strings.Contains(s, "nutrition") ||
		strings.Contains(s, "nutritional") ||
		strings.Contains(s, "пищев") ||
		strings.Contains(s, "питательн") ||
		strings.Contains(s, "энергетическ") ||
		strings.Contains(s, "energy") ||
		strings.Contains(s, "calories")
}

func parseIngredients(block string) []OCRIngredient {
	block = strings.TrimSpace(block)
	if block == "" {
		return nil
	}
	block = trimIngredientsBlock(block)
	if strings.TrimSpace(block) == "" {
		return nil
	}
	block = normalizeIngredientsBlock(block)

	seen := map[string]struct{}{}
	out := make([]OCRIngredient, 0, 24)
	for _, part := range splitIngredientsBlock(block) {
		raw := strings.TrimSpace(part)
		raw = strings.TrimFunc(raw, func(r rune) bool {
			return unicode.IsSpace(r) || r == '-' || r == '—' || r == '–' || r == ':' || r == '.' || r == '•'
		})
		if raw == "" {
			continue
		}
		raw = collapseSpaces(raw)
		name := normalizeIngredientName(raw)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, OCRIngredient{Raw: raw, Name: name})
	}
	return out
}

var ingredientListPrefix = regexp.MustCompile(`^\s*(?:[-*•]+|\d+[.)])\s*`)

func normalizeIngredientsBlock(block string) string {
	block = strings.ReplaceAll(block, "\t", " ")
	rawLines := strings.Split(block, "\n")
	lines := make([]string, 0, len(rawLines))
	hasExplicitSeparators := countTopLevelSeparators(block) > 0

	for _, rawLine := range rawLines {
		line := collapseSpaces(strings.TrimSpace(rawLine))
		line = ingredientListPrefix.ReplaceAllString(line, "")
		line = collapseSpaces(strings.TrimSpace(line))
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return ""
	}

	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			prev := lines[i-1]
			if shouldJoinIngredientLines(prev, line, hasExplicitSeparators) {
				b.WriteByte(' ')
			} else {
				b.WriteByte('\n')
			}
		}
		b.WriteString(line)
	}
	return b.String()
}

func splitIngredientsBlock(block string) []string {
	block = strings.TrimSpace(block)
	if block == "" {
		return nil
	}

	out := make([]string, 0, 24)
	var buf strings.Builder
	parenDepth := 0
	bracketDepth := 0

	flush := func() {
		item := collapseSpaces(strings.TrimSpace(buf.String()))
		if item != "" {
			out = append(out, item)
		}
		buf.Reset()
	}

	for _, r := range block {
		switch r {
		case '(':
			parenDepth++
			buf.WriteRune(r)
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
			buf.WriteRune(r)
		case '[':
			bracketDepth++
			buf.WriteRune(r)
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			buf.WriteRune(r)
		case ',', ';', '|':
			if parenDepth == 0 && bracketDepth == 0 {
				flush()
				continue
			}
			buf.WriteRune(r)
		case '\n':
			if parenDepth == 0 && bracketDepth == 0 {
				flush()
				continue
			}
			buf.WriteByte(' ')
		default:
			buf.WriteRune(r)
		}
	}
	flush()
	return out
}

func countTopLevelSeparators(block string) int {
	count := 0
	parenDepth := 0
	bracketDepth := 0
	for _, r := range block {
		switch r {
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case ',', ';', '|':
			if parenDepth == 0 && bracketDepth == 0 {
				count++
			}
		}
	}
	return count
}

func shouldJoinIngredientLines(prev string, next string, hasExplicitSeparators bool) bool {
	prev = strings.TrimSpace(prev)
	next = strings.TrimSpace(next)
	if prev == "" || next == "" {
		return false
	}
	if endsWithTopLevelSeparator(prev) {
		return false
	}
	if hasUnclosedParenthesis(prev) {
		return true
	}
	return hasExplicitSeparators && endsWithTopLevelSeparator(next)
}

func endsWithTopLevelSeparator(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	parenDepth := 0
	bracketDepth := 0
	lastTopLevel := rune(0)
	for _, r := range s {
		switch r {
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		default:
			if parenDepth == 0 && bracketDepth == 0 && !unicode.IsSpace(r) {
				lastTopLevel = r
			}
		}
	}
	return lastTopLevel == ',' || lastTopLevel == ';' || lastTopLevel == '|'
}

func hasUnclosedParenthesis(s string) bool {
	parenDepth := 0
	bracketDepth := 0
	for _, r := range s {
		switch r {
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		}
	}
	return parenDepth > 0 || bracketDepth > 0
}

func trimIngredientsBlock(block string) string {
	block = strings.TrimSpace(block)
	if block == "" {
		return ""
	}
	lower := strings.ToLower(block)
	markers := []string{
		"contains",
		"may contain",
		"allergen",
		"storage",
		"manufacturer",
		"address",
		"importer",
		"содержит",
		"может содерж",
		"аллерген",
		"хранить",
		"изготовитель",
		"адрес",
		"импортер",
		"производ",
	}
	cut := -1
	for _, m := range markers {
		idx := strings.Index(lower, m)
		if idx < 0 {
			continue
		}
		if idx <= 15 {
			continue
		}
		if cut == -1 || idx < cut {
			cut = idx
		}
	}
	if cut >= 0 {
		return strings.TrimSpace(block[:cut])
	}
	return block
}

var parenContent = regexp.MustCompile(`\([^)]*\)`)

func normalizeIngredientName(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	raw = strings.NewReplacer("—", "-", "–", "-", "/", " ").Replace(raw)
	raw = collapseSpaces(raw)

	raw = strings.TrimSpace(strings.TrimRightFunc(raw, func(r rune) bool {
		return unicode.IsDigit(r) || r == '%' || r == '.' || r == ',' || unicode.IsSpace(r)
	}))

	noParen := strings.TrimSpace(flattenUsefulParentheticalContent(raw))
	if noParen == "" {
		noParen = strings.TrimSpace(parenContent.ReplaceAllString(raw, " "))
	}
	noParen = collapseSpaces(noParen)
	if noParen != "" {
		raw = noParen
	}

	raw = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || r == '-' {
			return r
		}
		return ' '
	}, raw)
	raw = strings.TrimFunc(raw, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-'
	})
	raw = collapseSpaces(raw)
	if len([]rune(raw)) < 2 || !containsLetter(raw) {
		return ""
	}
	return raw
}

func flattenUsefulParentheticalContent(raw string) string {
	var out strings.Builder
	var inner strings.Builder
	parenDepth := 0

	appendSpace := func() {
		if out.Len() == 0 {
			return
		}
		s := out.String()
		if s[len(s)-1] != ' ' {
			out.WriteByte(' ')
		}
	}

	for _, r := range raw {
		switch r {
		case '(':
			if parenDepth == 0 {
				appendSpace()
			} else {
				inner.WriteRune(r)
			}
			parenDepth++
		case ')':
			if parenDepth == 0 {
				out.WriteByte(' ')
				continue
			}
			parenDepth--
			if parenDepth == 0 {
				content := normalizeForMatch(inner.String())
				inner.Reset()
				if shouldKeepParentheticalContent(content) {
					out.WriteString(content)
					out.WriteByte(' ')
				}
				continue
			}
			inner.WriteRune(r)
		default:
			if parenDepth > 0 {
				inner.WriteRune(r)
			} else {
				out.WriteRune(r)
			}
		}
	}

	return collapseSpaces(out.String())
}

func shouldKeepParentheticalContent(content string) bool {
	content = collapseSpaces(strings.TrimSpace(content))
	if content == "" {
		return false
	}
	words := strings.Fields(content)
	if len(words) > 5 {
		return false
	}
	return len([]rune(content)) <= 40
}

func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func normalizeForMatch(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, s)
	return collapseSpaces(s)
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// Nutrition parsing

var (
	reNum = regexp.MustCompile(`([0-9]+([\\.,][0-9]+)?)`)

	reValueKcal = regexp.MustCompile(`(?i)([0-9]+([\\.,][0-9]+)?)\\s*(kcal|ккал)`)
	reValueKj   = regexp.MustCompile(`(?i)([0-9]+([\\.,][0-9]+)?)\\s*(kj|кдж)`)
	reValueGram = regexp.MustCompile(`(?i)([0-9]+([\\.,][0-9]+)?)\\s*(g|гр|г)`)
)

func isEnergyLine(line string) bool {
	s := strings.ToLower(line)
	return strings.Contains(s, "energy") ||
		strings.Contains(s, "calor") ||
		strings.Contains(s, "энерг") ||
		strings.Contains(s, "ккал") ||
		strings.Contains(s, "kj") ||
		strings.Contains(s, "кдж")
}

func isProteinLine(line string) bool {
	s := strings.ToLower(line)
	return strings.Contains(s, "protein") || strings.Contains(s, "белк")
}

func isFatLine(line string) bool {
	s := strings.ToLower(line)
	return strings.Contains(s, "fat") || strings.Contains(s, "жир")
}

func isCarbLine(line string) bool {
	s := strings.ToLower(line)
	return strings.Contains(s, "carbohydrate") || strings.Contains(s, "carb") || strings.Contains(s, "углевод")
}

func parseNutrition(block string) OCRNutrition {
	block = strings.TrimSpace(block)
	if block == "" {
		return OCRNutrition{}
	}
	lines := strings.Split(block, "\n")
	n := OCRNutrition{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch {
		case isEnergyLine(line):
			val, unit, ok, estimated := parseNumberWithEnergyUnit(line)
			if ok {
				n.Calories.Value = &val
				n.EnergyUnit = unit
				n.Calories.IsEstimated = n.Calories.IsEstimated || estimated
			}
		case isProteinLine(line):
			val, unit, ok := parseNumberWithMassUnit(line)
			if ok {
				n.Protein.Value = &val
				n.MassUnit = preferUnit(n.MassUnit, unit)
				n.Protein.IsEstimated = n.Protein.IsEstimated || unit == ""
			}
		case isFatLine(line):
			val, unit, ok := parseNumberWithMassUnit(line)
			if ok {
				n.Fat.Value = &val
				n.MassUnit = preferUnit(n.MassUnit, unit)
				n.Fat.IsEstimated = n.Fat.IsEstimated || unit == ""
			}
		case isCarbLine(line):
			val, unit, ok := parseNumberWithMassUnit(line)
			if ok {
				n.Carbs.Value = &val
				n.MassUnit = preferUnit(n.MassUnit, unit)
				n.Carbs.IsEstimated = n.Carbs.IsEstimated || unit == ""
			}
		}
	}

	if n.Calories.Value != nil && n.EnergyUnit == "" {
		n.Calories.IsEstimated = true
	}
	if (n.Protein.Value != nil || n.Fat.Value != nil || n.Carbs.Value != nil) && n.MassUnit == "" {
		n.Protein.IsEstimated = n.Protein.Value != nil
		n.Fat.IsEstimated = n.Fat.Value != nil
		n.Carbs.IsEstimated = n.Carbs.Value != nil
	}
	return n
}

func preferUnit(existing, candidate string) string {
	if existing != "" {
		return existing
	}
	return candidate
}

func parseNumberWithEnergyUnit(line string) (float64, string, bool, bool) {
	if m := reValueKcal.FindStringSubmatch(line); len(m) > 0 {
		val, ok := parseFloat(m[1])
		if !ok {
			return 0, "", false, false
		}
		return val, "kcal", true, false
	}
	if m := reValueKj.FindStringSubmatch(line); len(m) > 0 {
		valKj, ok := parseFloat(m[1])
		if !ok {
			return 0, "", false, false
		}
		kcal := valKj / 4.184
		kcal = math.Round(kcal*10) / 10
		return kcal, "kcal", true, true
	}
	if m := reNum.FindStringSubmatch(line); len(m) > 0 {
		val, ok := parseFloat(m[1])
		if !ok {
			return 0, "", false, false
		}
		return val, "", true, true
	}
	return 0, "", false, false
}

func parseNumberWithMassUnit(line string) (float64, string, bool) {
	if m := reValueGram.FindStringSubmatch(line); len(m) > 0 {
		val, ok := parseFloat(m[1])
		if !ok {
			return 0, "", false
		}
		return val, "g", true
	}
	if m := reNum.FindStringSubmatch(line); len(m) > 0 {
		val, ok := parseFloat(m[1])
		if !ok {
			return 0, "", false
		}
		return val, "", true
	}
	return 0, "", false
}

func parseFloat(s string) (float64, bool) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) || f < 0 {
		return 0, false
	}
	return f, true
}

// Confidence scoring

func estimateOCRQuality(text string) float64 {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return 0
	}
	printable := 0
	weird := 0
	for _, r := range runes {
		if r == '\uFFFD' {
			weird++
		}
		if r == '\n' || r == '\t' || r == ' ' {
			printable++
			continue
		}
		if r >= 0x20 && r != 0x7F {
			printable++
		}
	}
	base := float64(printable) / float64(len(runes))
	penalty := float64(weird) / float64(len(runes))
	return clamp01(base - penalty*2)
}

func ingredientConfidence(ocrQuality float64, ing OCRIngredient) float64 {
	if strings.TrimSpace(ing.Name) == "" {
		return 0
	}
	parseQ := 0.7
	if len([]rune(ing.Name)) >= 4 {
		parseQ = 0.85
	}
	matchQ := clamp01(ing.MatchScore)
	return clamp01(0.15 + 0.35*clamp01(ocrQuality) + 0.25*parseQ + 0.25*matchQ)
}

func scoreNutritionFields(ocrQuality float64, n *OCRNutrition) {
	if n == nil {
		return
	}
	n.Calories.Confidence = nutritionFieldConfidence(ocrQuality, n.Calories, "kcal", n.EnergyUnit)
	n.Protein.Confidence = nutritionFieldConfidence(ocrQuality, n.Protein, "g", n.MassUnit)
	n.Fat.Confidence = nutritionFieldConfidence(ocrQuality, n.Fat, "g", n.MassUnit)
	n.Carbs.Confidence = nutritionFieldConfidence(ocrQuality, n.Carbs, "g", n.MassUnit)
}

func nutritionFieldConfidence(ocrQuality float64, f OCRNutritionField, expectedUnit string, detectedUnit string) float64 {
	if f.Value == nil {
		return 0
	}
	unitQ := 0.5
	if detectedUnit == expectedUnit {
		unitQ = 1
	} else if detectedUnit != "" && detectedUnit != expectedUnit {
		unitQ = 0.6
	}
	estPenalty := 0.0
	if f.IsEstimated {
		estPenalty = 0.2
	}
	return clamp01(0.2 + 0.5*clamp01(ocrQuality) + 0.3*unitQ - estPenalty)
}

func overallConfidence(d OCRDraft) float64 {
	ingAvg := 0.0
	for _, ing := range d.Ingredients {
		ingAvg += clamp01(ing.Confidence)
	}
	if len(d.Ingredients) > 0 {
		ingAvg /= float64(len(d.Ingredients))
	}
	nutFields := []OCRNutritionField{d.Nutrition.Calories, d.Nutrition.Protein, d.Nutrition.Fat, d.Nutrition.Carbs}
	nutAvg := 0.0
	nutCount := 0
	for _, f := range nutFields {
		if f.Value == nil {
			continue
		}
		nutAvg += clamp01(f.Confidence)
		nutCount++
	}
	if nutCount > 0 {
		nutAvg /= float64(nutCount)
	}

	completeness := 0.0
	if len(d.Ingredients) > 0 {
		completeness += 0.5
	}
	if nutCount > 0 {
		completeness += 0.5
	}

	return clamp01(0.15 + 0.5*math.Max(ingAvg, 0) + 0.2*nutAvg + 0.15*completeness - 0.1*float64(len(d.Conflicts)))
}

func missingFields(ings []OCRIngredient, n OCRNutrition) []string {
	var missing []string
	if len(ings) == 0 {
		missing = append(missing, "ingredients")
	}
	if n.Calories.Value == nil {
		missing = append(missing, "nutrition.calories")
	}
	if n.Protein.Value == nil {
		missing = append(missing, "nutrition.protein")
	}
	if n.Fat.Value == nil {
		missing = append(missing, "nutrition.fat")
	}
	if n.Carbs.Value == nil {
		missing = append(missing, "nutrition.carbs")
	}
	if n.Calories.Value != nil && n.EnergyUnit == "" {
		missing = append(missing, "nutrition.energyUnit")
	}
	if (n.Protein.Value != nil || n.Fat.Value != nil || n.Carbs.Value != nil) && n.MassUnit == "" {
		missing = append(missing, "nutrition.massUnit")
	}
	return missing
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

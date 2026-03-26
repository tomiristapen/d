package product

import (
	"math"
	"sort"
	"strings"
)

func mergeDrafts(drafts []OCRDraft) OCRDraft {
	if len(drafts) == 0 {
		return OCRDraft{}
	}
	if len(drafts) == 1 {
		return drafts[0]
	}

	out := OCRDraft{}

	maxQ := 0.0
	for _, d := range drafts {
		if d.OCRQuality > maxQ {
			maxQ = d.OCRQuality
		}
	}
	out.OCRQuality = maxQ

	// Ingredients: union by normalized name.
	byKey := map[string]OCRIngredient{}
	for _, d := range drafts {
		for _, ing := range d.Ingredients {
			key := strings.ToLower(strings.TrimSpace(ing.Name))
			if key == "" {
				continue
			}
			existing, ok := byKey[key]
			if !ok {
				byKey[key] = ing
				continue
			}
			// Prefer verified, then higher confidence.
			if ing.IsVerified && !existing.IsVerified {
				byKey[key] = ing
				continue
			}
			if ing.Confidence > existing.Confidence {
				// Keep verification if any was verified.
				ing.IsVerified = ing.IsVerified || existing.IsVerified
				byKey[key] = ing
				continue
			}
			// Merge verification flag.
			existing.IsVerified = existing.IsVerified || ing.IsVerified
			byKey[key] = existing
		}
	}
	out.Ingredients = make([]OCRIngredient, 0, len(byKey))
	for _, ing := range byKey {
		ing.ClientID = stableIngredientClientID(ing.Name)
		out.Ingredients = append(out.Ingredients, ing)
	}
	sort.Slice(out.Ingredients, func(i, j int) bool { return out.Ingredients[i].Name < out.Ingredients[j].Name })

	// Nutrition: choose highest confidence per field, detect conflicts.
	out.Nutrition = OCRNutrition{}
	out.Conflicts = nil
	for _, d := range drafts {
		mergeNutritionUnit(&out.Nutrition, d.Nutrition)
	}
	out.Nutrition.Calories, out.Conflicts = mergeNutritionField("nutrition.calories", out.Nutrition.Calories, drafts, func(d OCRDraft) OCRNutritionField {
		return d.Nutrition.Calories
	}, out.Conflicts)
	out.Nutrition.Protein, out.Conflicts = mergeNutritionField("nutrition.protein", out.Nutrition.Protein, drafts, func(d OCRDraft) OCRNutritionField {
		return d.Nutrition.Protein
	}, out.Conflicts)
	out.Nutrition.Fat, out.Conflicts = mergeNutritionField("nutrition.fat", out.Nutrition.Fat, drafts, func(d OCRDraft) OCRNutritionField {
		return d.Nutrition.Fat
	}, out.Conflicts)
	out.Nutrition.Carbs, out.Conflicts = mergeNutritionField("nutrition.carbs", out.Nutrition.Carbs, drafts, func(d OCRDraft) OCRNutritionField {
		return d.Nutrition.Carbs
	}, out.Conflicts)

	return out
}

func mergeNutritionUnit(out *OCRNutrition, candidate OCRNutrition) {
	if out == nil {
		return
	}
	if out.EnergyUnit == "" && candidate.EnergyUnit != "" {
		out.EnergyUnit = candidate.EnergyUnit
	}
	if out.MassUnit == "" && candidate.MassUnit != "" {
		out.MassUnit = candidate.MassUnit
	}
}

func mergeNutritionField(
	field string,
	current OCRNutritionField,
	drafts []OCRDraft,
	get func(OCRDraft) OCRNutritionField,
	conflicts []OCRConflict,
) (OCRNutritionField, []OCRConflict) {
	best := current
	var seenValues []float64
	for _, d := range drafts {
		f := get(d)
		if f.Value == nil {
			continue
		}
		seenValues = append(seenValues, *f.Value)
		if best.Value == nil || f.Confidence > best.Confidence {
			best = f
		}
	}

	// Conflict detection: two materially different values across photos.
	if len(seenValues) >= 2 {
		minV, maxV := seenValues[0], seenValues[0]
		for _, v := range seenValues[1:] {
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
		if minV > 0 {
			ratio := maxV / minV
			if ratio >= 1.12 { // >12% difference
				conflicts = append(conflicts, OCRConflict{
					Field: field,
					Note:  "values differ between photos",
				})
			}
		} else if math.Abs(maxV-minV) > 0.001 {
			conflicts = append(conflicts, OCRConflict{
				Field: field,
				Note:  "values differ between photos",
			})
		}
	}

	return best, conflicts
}

package product

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ExternalProvider interface {
	FetchByBarcode(ctx context.Context, barcode string) (Product, error)
}

type OpenFoodFactsClient struct {
	BaseURL string
	HTTP    *http.Client
}

func NewOpenFoodFactsClient(baseURL string) *OpenFoodFactsClient {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://world.openfoodfacts.org"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenFoodFactsClient{
		BaseURL: baseURL,
		HTTP: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (c *OpenFoodFactsClient) FetchByBarcode(ctx context.Context, barcode string) (Product, error) {
	url := fmt.Sprintf("%s/api/v2/product/%s.json", c.BaseURL, barcode)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Product{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nutri-ai-app/0.1")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Product{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Product{}, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8_192))
		return Product{}, fmt.Errorf("openfoodfacts http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw offResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return Product{}, err
	}
	if raw.Status == 0 {
		return Product{}, ErrNotFound
	}
	return normalizeOpenFoodFacts(barcode, raw), nil
}

type offResponse struct {
	Status  int        `json:"status"`
	Code    string     `json:"code"`
	Product offProduct `json:"product"`
}

type offProduct struct {
	ProductName   string          `json:"product_name"`
	ProductNameEN string          `json:"product_name_en"`
	Brands        string          `json:"brands"`
	Ingredients   []offIngredient `json:"ingredients"`
	IngredientsTX string          `json:"ingredients_text"`
	IngredientsEN string          `json:"ingredients_text_en"`
	Nutriments    map[string]any  `json:"nutriments"`
}

type offIngredient struct {
	Text string `json:"text"`
	ID   string `json:"id"`
}

func normalizeOpenFoodFacts(barcode string, raw offResponse) Product {
	p := raw.Product

	name := strings.TrimSpace(p.ProductName)
	if name == "" {
		name = strings.TrimSpace(p.ProductNameEN)
	}
	if name == "" {
		name = "Unknown product"
	}

	brand := ""
	if strings.TrimSpace(p.Brands) != "" {
		parts := strings.Split(p.Brands, ",")
		brand = strings.TrimSpace(parts[0])
	}

	ingredients := extractIngredients(p.Ingredients, p.IngredientsTX, p.IngredientsEN)

	calories := readNutriment(p.Nutriments, "energy-kcal_100g", "energy-kcal")
	protein := readNutriment(p.Nutriments, "proteins_100g", "proteins")
	fat := readNutriment(p.Nutriments, "fat_100g", "fat")
	carbs := readNutriment(p.Nutriments, "carbohydrates_100g", "carbohydrates")

	confidence := confidenceScore(name, brand, ingredients, calories, protein, fat, carbs)

	return Product{
		Barcode:         canonicalBarcode(raw.Code, barcode),
		Name:            name,
		Brand:           brand,
		Ingredients:     ingredients,
		Calories:        calories,
		Protein:         protein,
		Fat:             fat,
		Carbohydrates:   carbs,
		ConfidenceScore: confidence,
		Source:          "openfoodfacts",
	}
}

func extractIngredients(items []offIngredient, ingredientsText string, ingredientsTextEN string) []string {
	var out []string
	seen := map[string]struct{}{}

	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}

	for _, ing := range items {
		add(ing.Text)
	}

	if len(out) == 0 {
		text := strings.TrimSpace(ingredientsText)
		if text == "" {
			text = strings.TrimSpace(ingredientsTextEN)
		}
		if text != "" {
			parts := strings.FieldsFunc(text, func(r rune) bool { return r == ',' || r == ';' || r == '\n' })
			for _, part := range parts {
				add(part)
			}
		}
	}

	const max = 100
	if len(out) > max {
		out = out[:max]
	}
	return out
}

func readNutriment(n map[string]any, keys ...string) float64 {
	for _, key := range keys {
		value, ok := n[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case json.Number:
			f, err := v.Float64()
			if err == nil {
				return f
			}
		case string:
			f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err == nil {
				return f
			}
		}
	}
	return 0
}

func confidenceScore(name string, brand string, ingredients []string, calories float64, protein float64, fat float64, carbs float64) float64 {
	total := 7.0
	have := 0.0
	if strings.TrimSpace(name) != "" && name != "Unknown product" {
		have++
	}
	if strings.TrimSpace(brand) != "" {
		have++
	}
	if len(ingredients) > 0 {
		have++
	}
	if calories > 0 {
		have++
	}
	if protein > 0 {
		have++
	}
	if fat > 0 {
		have++
	}
	if carbs > 0 {
		have++
	}
	if total == 0 {
		return 0
	}
	score := have / total
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

var _ ExternalProvider = (*OpenFoodFactsClient)(nil)

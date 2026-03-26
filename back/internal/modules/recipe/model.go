package recipe

import "back/internal/modules/input"

type IngredientInput struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type ResolvedIngredient struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type AnalyzeRequest struct {
	Name        string            `json:"name,omitempty"`
	Ingredients []IngredientInput `json:"ingredients"`
}

type AnalyzeResponse struct {
	Product     input.ProductResult  `json:"product"`
	Per100G     input.ProductResult  `json:"per_100g"`
	Ingredients []ResolvedIngredient `json:"ingredients"`
	Confidence  float64              `json:"confidence"`
	AmountG     float64              `json:"amount_g"`
}

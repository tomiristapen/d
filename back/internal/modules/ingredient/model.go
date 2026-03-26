package ingredient

import "back/internal/modules/nutrition"

type Nutrients = nutrition.Nutrients

type Ingredient struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Per100g Nutrients `json:"per_100g"`
}

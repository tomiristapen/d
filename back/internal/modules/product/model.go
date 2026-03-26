package product

import "time"

type Product struct {
	ID              int64     `json:"id"`
	Barcode         string    `json:"barcode"`
	Name            string    `json:"name"`
	Brand           string    `json:"brand"`
	Ingredients     []string  `json:"ingredients"`
	Calories        float64   `json:"calories"`
	Protein         float64   `json:"protein"`
	Fat             float64   `json:"fat"`
	Carbohydrates   float64   `json:"carbohydrates"`
	ConfidenceScore float64   `json:"confidenceScore"`
	Source          string    `json:"source"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

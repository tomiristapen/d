package input

type ProductResult struct {
	Name     string  `json:"name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type ManualAnalyzeRequest struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type ManualAnalyzeResponse struct {
	Product     *ProductResult `json:"product"`
	Per100G     *ProductResult `json:"per_100g,omitempty"`
	Suggestions []string       `json:"suggestions"`
	Confidence  *float64       `json:"confidence"`
	AmountG     float64        `json:"amount_g"`
}

type CustomProductRequest struct {
	Name     string  `json:"name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type CustomProductResponse struct {
	Status string `json:"status"`
}

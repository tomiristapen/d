package product

// OCRDraftRequest is the input for the structured, user-assisted OCR flow.
// It supports multi-photo uploads by allowing multiple base64-encoded images.
type OCRDraftRequest struct {
	// Images are base64-encoded image bytes (preferred) or base64-encoded UTF-8 text (dev/legacy).
	Images []string `json:"images,omitempty"`
	// Image is a single-image alias for backwards compatibility with earlier clients.
	Image string `json:"image,omitempty"`

	Lang   string `json:"lang,omitempty"`
	Region string `json:"region,omitempty"`
}

// Domain models (internal).

type OCRDraft struct {
	OCRQuality        float64
	OverallConfidence float64

	Ingredients []OCRIngredient
	Nutrition   OCRNutrition

	MissingFields []string
	Conflicts     []OCRConflict
}

type OCRIngredient struct {
	ClientID string

	Raw  string
	Name string // normalized, user-editable canonical string

	MatchedProductID *int64
	MatchName        string
	MatchScore       float64

	Confidence float64
	IsVerified bool
}

type OCRNutrition struct {
	EnergyUnit string // "kcal" or "kJ" (empty when unknown)
	MassUnit   string // "g" (empty when unknown)

	Calories OCRNutritionField
	Protein  OCRNutritionField
	Fat      OCRNutritionField
	Carbs    OCRNutritionField
}

type OCRNutritionField struct {
	Value *float64

	Confidence  float64
	IsEstimated bool
	IsVerified  bool
}

type OCRConflict struct {
	Field string
	Note  string
}

// DTOs (API).

type OCRDraftDTO struct {
	OCRMode           string  `json:"ocrMode"`
	OCRQuality        float64 `json:"ocrQuality"`
	OverallConfidence float64 `json:"overallConfidence"`

	Ingredients []OCRIngredientDTO `json:"ingredients"`
	Nutrition   OCRNutritionDTO    `json:"nutrition"`

	MissingFields []string         `json:"missingFields"`
	Conflicts     []OCRConflictDTO `json:"conflicts,omitempty"`
}

type OCRIngredientDTO struct {
	ClientID string `json:"clientId"`

	Raw  string `json:"raw"`
	Name string `json:"name"`

	MatchedProductID *int64  `json:"matchedProductId,omitempty"`
	MatchName        string  `json:"matchName,omitempty"`
	MatchScore       float64 `json:"matchScore"`

	Confidence float64 `json:"confidence"`
	IsVerified bool    `json:"isVerified"`
}

type OCRNutritionDTO struct {
	EnergyUnit string `json:"energyUnit"` // "kcal" | "kJ" | ""
	MassUnit   string `json:"massUnit"`   // "g" | ""

	Calories OCRNutritionFieldDTO `json:"calories"`
	Protein  OCRNutritionFieldDTO `json:"protein"`
	Fat      OCRNutritionFieldDTO `json:"fat"`
	Carbs    OCRNutritionFieldDTO `json:"carbs"`
}

type OCRNutritionFieldDTO struct {
	Value *float64 `json:"value,omitempty"`

	Confidence  float64 `json:"confidence"`
	IsEstimated bool    `json:"isEstimated"`
	IsVerified  bool    `json:"isVerified"`
}

type OCRConflictDTO struct {
	Field string `json:"field"`
	Note  string `json:"note"`
}

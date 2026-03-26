package profile

import "time"

type UpdateRequest struct {
	Age                  *int      `json:"age"`
	Gender               *string   `json:"gender"`
	HeightCM             *float64  `json:"height_cm"`
	WeightKG             *float64  `json:"weight_kg"`
	ActivityLevel        *string   `json:"activity_level"`
	Goal                 *string   `json:"goal"`
	NutritionGoal        *string   `json:"nutrition_goal"`
	Allergies            *[]string `json:"allergies"`
	CustomAllergies      *[]string `json:"custom_allergies"`
	Intolerances         *[]string `json:"intolerances"`
	DietType             *string   `json:"diet_type"`
	ReligiousRestriction *string   `json:"religious_restriction"`
}

type ProfileResponse struct {
	UserID               string    `json:"user_id"`
	Age                  int       `json:"age"`
	Gender               string    `json:"gender"`
	HeightCM             float64   `json:"height_cm"`
	WeightKG             float64   `json:"weight_kg"`
	ActivityLevel        string    `json:"activity_level"`
	Goal                 string    `json:"goal"`
	Allergies            []string  `json:"allergies"`
	CustomAllergies      []string  `json:"custom_allergies"`
	Intolerances         []string  `json:"intolerances"`
	DietType             string    `json:"diet_type"`
	ReligiousRestriction string    `json:"religious_restriction"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ResetResponse struct {
	Status string `json:"status"`
}

package profile

import (
	"time"

	"back/internal/modules/nutrition"
)

type Profile struct {
	UserID               string              `json:"user_id"`
	Age                  int                 `json:"age"`
	Gender               string              `json:"gender"`
	HeightCM             float64             `json:"height_cm"`
	WeightKG             float64             `json:"weight_kg"`
	ActivityLevel        string              `json:"activity_level"`
	Goal                 string              `json:"goal"`
	Allergies            []string            `json:"allergies"`
	CustomAllergies      []string            `json:"custom_allergies"`
	Intolerances         []string            `json:"intolerances"`
	DietType             string              `json:"diet_type"`
	ReligiousRestriction string              `json:"religious_restriction"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
	Targets              nutrition.Nutrients `json:"-"`
}

func NewProfile(userID string, age int, gender string, heightCM float64, weightKG float64, activityLevel string, goal string, allergies []string, customAllergies []string, intolerances []string, dietType string, religiousRestriction string) Profile {
	now := time.Now().UTC()
	return Profile{
		UserID:               userID,
		Age:                  age,
		Gender:               gender,
		HeightCM:             heightCM,
		WeightKG:             weightKG,
		ActivityLevel:        activityLevel,
		Goal:                 goal,
		Allergies:            allergies,
		CustomAllergies:      customAllergies,
		Intolerances:         intolerances,
		DietType:             dietType,
		ReligiousRestriction: religiousRestriction,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

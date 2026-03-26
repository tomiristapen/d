package onboarding

import "back/internal/modules/profilemeta"

type StatusResponse struct {
	ProfileCompleted bool `json:"profile_completed"`
}

type OptionsResponse struct {
	Genders               []profilemeta.Option `json:"genders"`
	ActivityLevels        []profilemeta.Option `json:"activity_levels"`
	Goals                 []profilemeta.Option `json:"goals"`
	NutritionGoals        []profilemeta.Option `json:"nutrition_goals"`
	Allergies             []profilemeta.Option `json:"allergies"`
	Intolerances          []profilemeta.Option `json:"intolerances"`
	DietTypes             []profilemeta.Option `json:"diet_types"`
	ReligiousRestrictions []profilemeta.Option `json:"religious_restrictions"`
}

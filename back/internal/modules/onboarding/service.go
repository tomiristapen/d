package onboarding

import (
	"context"

	"back/internal/modules/profilemeta"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Status(ctx context.Context, userID string) (Status, error) {
	return s.repo.GetStatus(ctx, userID)
}

func (s *Service) Options() OptionsResponse {
	catalog := profilemeta.Options()
	return OptionsResponse{
		Genders:               catalog.Genders,
		ActivityLevels:        catalog.ActivityLevels,
		Goals:                 catalog.Goals,
		NutritionGoals:        catalog.NutritionGoals,
		Allergies:             catalog.Allergies,
		Intolerances:          catalog.Intolerances,
		DietTypes:             catalog.DietTypes,
		ReligiousRestrictions: catalog.ReligiousRestrictions,
	}
}

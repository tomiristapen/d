package profile

import (
	"context"
	"fmt"
	"strings"
	"time"

	"back/internal/modules/nutrition"
	"back/internal/modules/profilemeta"
)

type Service struct {
	repo Repository
}

const (
	minAge      = 1
	maxAge      = 120
	minHeightCM = 30
	maxHeightCM = 300
	minWeightKG = 2
	maxWeightKG = 500
)

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Update(ctx context.Context, userID string, req UpdateRequest) (Profile, error) {
	if userID == "" {
		return Profile{}, fmt.Errorf("user id is required")
	}

	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil && err != ErrNotFound {
		return Profile{}, err
	}
	hasExisting := err == nil

	merged, err := mergeProfile(existing, hasExisting, userID, req)
	if err != nil {
		return Profile{}, err
	}

	if err := s.repo.Save(ctx, merged); err != nil {
		return Profile{}, err
	}
	return s.repo.GetByUserID(ctx, userID)
}

func mergeProfile(existing Profile, hasExisting bool, userID string, req UpdateRequest) (Profile, error) {
	goal := firstNonEmptyPtr(req.Goal, req.NutritionGoal)
	if !hasExisting {
		// Creating: require all fields present (same behavior as the legacy onboarding complete).
		if req.Age == nil || req.Gender == nil || req.HeightCM == nil || req.WeightKG == nil ||
			req.ActivityLevel == nil || goal == nil || req.Allergies == nil || req.CustomAllergies == nil ||
			req.Intolerances == nil || req.DietType == nil || req.ReligiousRestriction == nil {
			return Profile{}, fmt.Errorf("missing required fields")
		}
		existing = NewProfile(userID, 0, "", 0, 0, "", "", nil, nil, nil, "", "")
	}

	merged := existing
	merged.UserID = userID

	if req.Age != nil {
		merged.Age = *req.Age
	}
	if req.Gender != nil {
		merged.Gender = normalizeValue(*req.Gender)
	}
	if req.HeightCM != nil {
		merged.HeightCM = *req.HeightCM
	}
	if req.WeightKG != nil {
		merged.WeightKG = *req.WeightKG
	}
	if req.ActivityLevel != nil {
		merged.ActivityLevel = normalizeActivityLevel(*req.ActivityLevel)
	}
	if goal != nil {
		merged.Goal = normalizeGoal(*goal)
	}
	if req.DietType != nil {
		merged.DietType = normalizeValue(*req.DietType)
	}
	if req.ReligiousRestriction != nil {
		merged.ReligiousRestriction = normalizeValue(*req.ReligiousRestriction)
	}
	if req.Allergies != nil {
		merged.Allergies = normalizeList(*req.Allergies)
	}
	if req.CustomAllergies != nil {
		merged.CustomAllergies = normalizeList(*req.CustomAllergies)
	}
	if req.Intolerances != nil {
		merged.Intolerances = normalizeList(*req.Intolerances)
	}

	if merged.Age < minAge || merged.Age > maxAge {
		return Profile{}, fmt.Errorf("age must be between %d and %d", minAge, maxAge)
	}
	if merged.HeightCM < minHeightCM || merged.HeightCM > maxHeightCM {
		return Profile{}, fmt.Errorf("height_cm must be between %d and %d", minHeightCM, maxHeightCM)
	}
	if merged.WeightKG < minWeightKG || merged.WeightKG > maxWeightKG {
		return Profile{}, fmt.Errorf("weight_kg must be between %d and %d", minWeightKG, maxWeightKG)
	}
	if !profilemeta.IsValidGender(merged.Gender) {
		return Profile{}, fmt.Errorf("invalid gender")
	}
	if !profilemeta.IsValidActivityLevel(merged.ActivityLevel) {
		return Profile{}, fmt.Errorf("invalid activity_level")
	}
	if !profilemeta.IsValidGoal(merged.Goal) {
		return Profile{}, fmt.Errorf("invalid goal")
	}
	if !profilemeta.IsValidDietType(merged.DietType) {
		return Profile{}, fmt.Errorf("invalid diet_type")
	}
	if !profilemeta.IsValidReligiousRestriction(merged.ReligiousRestriction) {
		return Profile{}, fmt.Errorf("invalid religious_restriction")
	}
	for _, allergy := range merged.Allergies {
		if !profilemeta.IsValidAllergy(allergy) {
			return Profile{}, fmt.Errorf("invalid allergy")
		}
	}
	for _, intolerance := range merged.Intolerances {
		if !profilemeta.IsValidIntolerance(intolerance) {
			return Profile{}, fmt.Errorf("invalid intolerance")
		}
	}

	merged.UpdatedAt = time.Now().UTC()
	if merged.CreatedAt.IsZero() {
		merged.CreatedAt = merged.UpdatedAt
	}
	targets, err := nutrition.CalculateDailyTargets(nutrition.TargetProfile{
		WeightKG:      merged.WeightKG,
		HeightCM:      merged.HeightCM,
		Age:           merged.Age,
		Gender:        merged.Gender,
		ActivityLevel: merged.ActivityLevel,
		Goal:          merged.Goal,
	})
	if err != nil {
		return Profile{}, err
	}
	merged.Targets = targets
	return merged, nil
}

func (s *Service) Get(ctx context.Context, userID string) (Profile, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) Reset(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}

func normalizeValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeList(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = normalizeValue(value)
		if value != "" {
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func firstNonEmptyPtr(values ...*string) *string {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func normalizeActivityLevel(value string) string {
	switch normalizeValue(value) {
	case "low":
		return "sedentary"
	case "high":
		return "active"
	default:
		return normalizeValue(value)
	}
}

func normalizeGoal(value string) string {
	switch normalizeValue(value) {
	case "lose_weight":
		return "lose"
	case "maintain_weight", "healthy_eating":
		return "maintain"
	case "gain_weight":
		return "gain"
	default:
		return normalizeValue(value)
	}
}

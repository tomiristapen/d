package profile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type memoryRepo struct {
	called  bool
	profile Profile
}

func (r *memoryRepo) Save(_ context.Context, profile Profile) error {
	r.called = true
	r.profile = profile
	return nil
}

func (r *memoryRepo) GetByUserID(_ context.Context, _ string) (Profile, error) {
	if !r.called {
		return Profile{}, ErrNotFound
	}
	return r.profile, nil
}

func (r *memoryRepo) Delete(_ context.Context, _ string) error { return nil }

func TestUpsert_NormalizesLists(t *testing.T) {
	repo := &memoryRepo{}
	service := NewService(repo)

	_, err := service.Update(context.Background(), "user-1", UpdateRequest{
		Age:                  ptrInt(28),
		Gender:               ptrString("Female"),
		HeightCM:             ptrFloat(168),
		WeightKG:             ptrFloat(60),
		ActivityLevel:        ptrString("moderate"),
		Goal:                 ptrString("maintain"),
		Allergies:            ptrStrings([]string{"milk", " Milk "}),
		CustomAllergies:      ptrStrings([]string{" Honey ", "honey"}),
		Intolerances:         ptrStrings([]string{"gluten"}),
		DietType:             ptrString("vegetarian"),
		ReligiousRestriction: ptrString("halal"),
	})
	require.NoError(t, err)
	require.True(t, repo.called)
	require.Equal(t, []string{"milk"}, repo.profile.Allergies)
	require.Equal(t, []string{"honey"}, repo.profile.CustomAllergies)
}

func TestUpsert_NormalizesIntolerancesBeforeValidation(t *testing.T) {
	repo := &memoryRepo{}
	service := NewService(repo)

	_, err := service.Update(context.Background(), "user-1", UpdateRequest{
		Age:                  ptrInt(28),
		Gender:               ptrString("Female"),
		HeightCM:             ptrFloat(168),
		WeightKG:             ptrFloat(60),
		ActivityLevel:        ptrString("light"),
		Goal:                 ptrString("maintain"),
		Allergies:            ptrStrings([]string{}),
		CustomAllergies:      ptrStrings([]string{}),
		Intolerances:         ptrStrings([]string{" Lactose "}),
		DietType:             ptrString("none"),
		ReligiousRestriction: ptrString("none"),
	})
	require.NoError(t, err)
	require.Equal(t, []string{"lactose"}, repo.profile.Intolerances)
}

func TestUpsert_RejectsInvalidActivityLevel(t *testing.T) {
	repo := &memoryRepo{}
	service := NewService(repo)

	_, err := service.Update(context.Background(), "user-1", UpdateRequest{
		Age:                  ptrInt(28),
		Gender:               ptrString("Female"),
		HeightCM:             ptrFloat(168),
		WeightKG:             ptrFloat(60),
		ActivityLevel:        ptrString("super_active"),
		Goal:                 ptrString("maintain"),
		Allergies:            ptrStrings([]string{}),
		CustomAllergies:      ptrStrings([]string{}),
		Intolerances:         ptrStrings([]string{"lactose"}),
		DietType:             ptrString("none"),
		ReligiousRestriction: ptrString("none"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid activity_level")
}

func TestUpsert_RejectsInvalidAllergy(t *testing.T) {
	repo := &memoryRepo{}
	service := NewService(repo)

	_, err := service.Update(context.Background(), "user-1", UpdateRequest{
		Age:                  ptrInt(28),
		Gender:               ptrString("Female"),
		HeightCM:             ptrFloat(168),
		WeightKG:             ptrFloat(60),
		ActivityLevel:        ptrString("light"),
		Goal:                 ptrString("maintain"),
		Allergies:            ptrStrings([]string{"dragonfruit"}),
		CustomAllergies:      ptrStrings([]string{}),
		Intolerances:         ptrStrings([]string{"lactose"}),
		DietType:             ptrString("none"),
		ReligiousRestriction: ptrString("none"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid allergy")
}

func ptrInt(v int) *int { return &v }

func ptrFloat(v float64) *float64 { return &v }

func ptrString(v string) *string { return &v }

func ptrStrings(v []string) *[]string { return &v }

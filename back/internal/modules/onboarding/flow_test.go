package onboarding

import (
	"context"
	"testing"

	"back/internal/modules/profile"

	"github.com/stretchr/testify/require"
)

type sharedFlowRepo struct {
	profile         profile.Profile
	hasProfile      bool
	profileComplete bool
}

func (r *sharedFlowRepo) Save(_ context.Context, p profile.Profile) error {
	r.profile = p
	r.hasProfile = true
	r.profileComplete = true
	return nil
}

func (r *sharedFlowRepo) GetByUserID(_ context.Context, _ string) (profile.Profile, error) {
	if !r.hasProfile {
		return profile.Profile{}, profile.ErrNotFound
	}
	return r.profile, nil
}

func (r *sharedFlowRepo) Delete(_ context.Context, _ string) error {
	r.hasProfile = false
	r.profileComplete = false
	return nil
}

func (r *sharedFlowRepo) GetStatus(_ context.Context, _ string) (Status, error) {
	return Status{ProfileCompleted: r.profileComplete}, nil
}

func TestProfileUpdateSyncsOnboardingStatus(t *testing.T) {
	repo := &sharedFlowRepo{}
	profileService := profile.NewService(repo)
	onboardingService := NewService(repo)

	status, err := onboardingService.Status(context.Background(), "user-1")
	require.NoError(t, err)
	require.False(t, status.ProfileCompleted)

	_, err = profileService.Update(context.Background(), "user-1", profile.UpdateRequest{
		Age:                  intPtr(28),
		Gender:               stringPtr("female"),
		HeightCM:             floatPtr(168),
		WeightKG:             floatPtr(60),
		ActivityLevel:        stringPtr("moderate"),
		Goal:                 stringPtr("maintain"),
		Allergies:            stringSlicePtr([]string{"milk"}),
		CustomAllergies:      stringSlicePtr([]string{}),
		Intolerances:         stringSlicePtr([]string{"lactose"}),
		DietType:             stringPtr("vegetarian"),
		ReligiousRestriction: stringPtr("none"),
	})
	require.NoError(t, err)

	status, err = onboardingService.Status(context.Background(), "user-1")
	require.NoError(t, err)
	require.True(t, status.ProfileCompleted)

	require.NoError(t, profileService.Reset(context.Background(), "user-1"))

	status, err = onboardingService.Status(context.Background(), "user-1")
	require.NoError(t, err)
	require.False(t, status.ProfileCompleted)
}

func intPtr(v int) *int { return &v }

func floatPtr(v float64) *float64 { return &v }

func stringPtr(v string) *string { return &v }

func stringSlicePtr(v []string) *[]string { return &v }

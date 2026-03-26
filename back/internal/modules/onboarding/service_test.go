package onboarding

import (
	"testing"

	"back/internal/modules/profilemeta"

	"github.com/stretchr/testify/require"
)

func TestOptionsExposeBackendCatalog(t *testing.T) {
	service := NewService(nil)

	resp := service.Options()

	require.NotEmpty(t, resp.ActivityLevels)
	require.NotEmpty(t, resp.Allergies)
	require.Contains(t, resp.Allergies, profilemeta.Option{Key: "milk", Label: "Milk"})
	require.Contains(t, resp.Genders, profilemeta.Option{Key: "female", Label: "Female"})
	require.Contains(t, resp.Goals, profilemeta.Option{Key: "maintain", Label: "Maintain"})
}

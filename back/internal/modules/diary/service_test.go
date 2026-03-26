package diary

import (
	"context"
	"testing"
	"time"

	"back/internal/modules/nutrition"

	"github.com/stretchr/testify/require"
)

type memoryRepo struct {
	saved Entry
}

func (r *memoryRepo) Create(_ context.Context, e Entry) (Entry, error) {
	e.ID = 1
	r.saved = e
	return e, nil
}

func (r *memoryRepo) GetDailyTotals(_ context.Context, userID string, day time.Time) (DailyTotals, error) {
	return DailyTotals{UserID: userID, Date: day.Format("2006-01-02")}, nil
}

func (r *memoryRepo) GetTargets(_ context.Context, _ string) (nutrition.Nutrients, error) {
	return nutrition.Nutrients{Calories: 2000, Protein: 100, Fat: 70, Carbs: 250}, nil
}

func (r *memoryRepo) DeleteEntry(_ context.Context, _ string, _ int64) error { return nil }

func (r *memoryRepo) UpdateEntry(_ context.Context, _ string, _ UpdateEntryInput) (Entry, error) {
	return Entry{}, nil
}

func TestService_AddEntry_ValidatesUserID(t *testing.T) {
	svc := NewService(&memoryRepo{})
	_, err := svc.AddEntry(context.Background(), Entry{Name: "x", Source: "manual", AmountG: 100})
	require.Error(t, err)
}

func TestService_AddEntry_Saves(t *testing.T) {
	repo := &memoryRepo{}
	svc := NewService(repo)

	got, err := svc.AddEntry(context.Background(), Entry{
		UserID:   "user-1",
		Source:   "manual",
		Name:     "Milk",
		AmountG:  100,
		Calories: 50,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), got.ID)
	require.Equal(t, "Milk", repo.saved.Name)
	require.NotEmpty(t, repo.saved.EntryDate)
}

func TestService_AddToDiary_ScalesPer100G(t *testing.T) {
	repo := &memoryRepo{}
	svc := NewService(repo)

	got, err := svc.AddToDiary(context.Background(), AddToDiaryInput{
		UserID:  "user-1",
		Source:  "barcode",
		Name:    "Yogurt",
		AmountG: 150,
		Per100G: nutrition.Nutrients{
			Calories: 100,
			Protein:  10,
			Fat:      4,
			Carbs:    12,
		},
		TimezoneOffsetMinutes: 300,
	})
	require.NoError(t, err)
	require.Equal(t, 150.0, got.Calories)
	require.Equal(t, 15.0, got.Protein)
	require.Equal(t, 6.0, got.Fat)
	require.Equal(t, 18.0, got.Carbs)
}

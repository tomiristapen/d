package ingredient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type memoryRepo struct {
	lastQuery string
	item      Ingredient
	items     []string
	err       error
}

func (r *memoryRepo) Search(_ context.Context, query string) ([]string, error) {
	r.lastQuery = query
	return append([]string(nil), r.items...), nil
}

func (r *memoryRepo) FindBestMatch(_ context.Context, query string) (Ingredient, error) {
	r.lastQuery = query
	return r.item, r.err
}

func TestService_LookupByName_NormalizesQuery(t *testing.T) {
	repo := &memoryRepo{item: Ingredient{ID: 1, Name: "Milk"}}
	svc := NewService(repo)

	_, err := svc.LookupByName(context.Background(), "  MILK  ")
	require.NoError(t, err)
	require.Equal(t, "milk", repo.lastQuery)
}

func TestService_LookupByName_Empty(t *testing.T) {
	repo := &memoryRepo{}
	svc := NewService(repo)

	_, err := svc.LookupByName(context.Background(), "   ")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestService_Autocomplete_DedupesSuggestions(t *testing.T) {
	repo := &memoryRepo{
		items: []string{"Honey", " honey ", "Mango", "mango"},
	}
	svc := NewService(repo)

	got, err := svc.Autocomplete(context.Background(), "  HO ")
	require.NoError(t, err)
	require.Equal(t, "ho", repo.lastQuery)
	require.Equal(t, []string{"Honey", "Mango"}, got)
}

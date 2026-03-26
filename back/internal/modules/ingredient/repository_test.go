package ingredient

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepositorySearch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)
	rows := sqlmock.NewRows([]string{"name"}).AddRow("Мёд").AddRow("Манго")

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH searchable_products AS (
			SELECT id, name, NULL::TEXT AS name_en, calories, protein, fat, carbs, 0 AS source_priority
			FROM base_products
			UNION ALL
			SELECT id, name, name_en, calories, protein, fat, carbs, 1 AS source_priority
			FROM food_reference_products_csv
		)
		SELECT name
		FROM searchable_products
		WHERE name ILIKE $1 OR COALESCE(name_en, '') ILIKE $1
		ORDER BY
			CASE
				WHEN name ILIKE $1 THEN 0
				WHEN COALESCE(name_en, '') ILIKE $1 THEN 1
				ELSE 2
			END,
			source_priority,
			LENGTH(name),
			name
		LIMIT 10
	`)).WithArgs("ho%").WillReturnRows(rows)

	items, err := repo.Search(context.Background(), "ho")
	require.NoError(t, err)
	require.Equal(t, []string{"Мёд", "Манго"}, items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryFindBestMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)
	rows := sqlmock.NewRows([]string{"id", "name", "calories", "protein", "fat", "carbs"}).
		AddRow(int64(131), "Укроп свежий", 43.0, 3.5, 1.1, 7.0)

	mock.ExpectQuery(regexp.QuoteMeta(`
		WITH searchable_products AS (
			SELECT id, name, NULL::TEXT AS name_en, calories, protein, fat, carbs, 0 AS source_priority
			FROM base_products
			UNION ALL
			SELECT id, name, name_en, calories, protein, fat, carbs, 1 AS source_priority
			FROM food_reference_products_csv
		)
		SELECT id, name, calories, protein, fat, carbs
		FROM searchable_products
		WHERE name ILIKE $1 OR COALESCE(name_en, '') ILIKE $1
		ORDER BY
			CASE
				WHEN name ILIKE $2 THEN 0
				WHEN COALESCE(name_en, '') ILIKE $2 THEN 1
				ELSE 2
			END,
			source_priority,
			LENGTH(name),
			name
		LIMIT 1
	`)).
		WithArgs("%dill%", "dill%").
		WillReturnRows(rows)

	item, err := repo.FindBestMatch(context.Background(), "dill")
	require.NoError(t, err)
	require.Equal(t, int64(131), item.ID)
	require.Equal(t, "Укроп свежий", item.Name)
	require.InDelta(t, 43, item.Per100g.Calories, 0.000001)
	require.InDelta(t, 3.5, item.Per100g.Protein, 0.000001)
	require.InDelta(t, 1.1, item.Per100g.Fat, 0.000001)
	require.InDelta(t, 7, item.Per100g.Carbs, 0.000001)
	require.NoError(t, mock.ExpectationsWereMet())
}

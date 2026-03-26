package product

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"back/internal/platform/errcode"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresBaseProductRepository(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, repo *PostgresBaseProductRepository, mock sqlmock.Sqlmock)
	}{
		{
			name: "find exact match returns base product",
			run: func(t *testing.T, repo *PostgresBaseProductRepository, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "calories", "protein", "fat", "carbs"}).
					AddRow(int64(1), "chicken", 165.0, 31.0, 3.6, 0.0)

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
		WHERE LOWER(name) = LOWER($1) OR LOWER(COALESCE(name_en, '')) = LOWER($1)
		ORDER BY source_priority, id
		LIMIT 1
	`)).
					WithArgs("chicken").
					WillReturnRows(rows)

				got, err := repo.FindExactByName(context.Background(), "chicken")
				require.NoError(t, err)
				require.Equal(t, int64(1), got.ID)
				require.Equal(t, "chicken", got.Name)
				require.InDelta(t, 165, got.Calories, 0.000001)
				require.InDelta(t, 31, got.Protein, 0.000001)
				require.InDelta(t, 3.6, got.Fat, 0.000001)
				require.InDelta(t, 0, got.Carbs, 0.000001)
			},
		},
		{
			name: "find exact match returns NOT_FOUND when missing",
			run: func(t *testing.T, repo *PostgresBaseProductRepository, mock sqlmock.Sqlmock) {
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
		WHERE LOWER(name) = LOWER($1) OR LOWER(COALESCE(name_en, '')) = LOWER($1)
		ORDER BY source_priority, id
		LIMIT 1
	`)).
					WithArgs("missing").
					WillReturnError(sql.ErrNoRows)

				_, err := repo.FindExactByName(context.Background(), "missing")
				require.ErrorIs(t, err, errcode.NotFound)
			},
		},
		{
			name: "find fuzzy match returns up to matching products",
			run: func(t *testing.T, repo *PostgresBaseProductRepository, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "calories", "protein", "fat", "carbs"}).
					AddRow(int64(1), "chicken breast", 165.0, 31.0, 3.6, 0.0).
					AddRow(int64(2), "chicken thigh", 177.0, 24.0, 8.0, 0.0)

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
		LIMIT $3
	`)).
					WithArgs("%chicken%", "chicken%", 10).
					WillReturnRows(rows)

				got, err := repo.FindFuzzyByName(context.Background(), "chicken")
				require.NoError(t, err)
				require.Len(t, got, 2)
				require.Equal(t, "chicken breast", got[0].Name)
				require.Equal(t, "chicken thigh", got[1].Name)
			},
		},
		{
			name: "find exact match resolves english reference alias",
			run: func(t *testing.T, repo *PostgresBaseProductRepository, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "calories", "protein", "fat", "carbs"}).
					AddRow(int64(133), "Мёд", 304.0, 0.3, 0.0, 82.4)

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
		WHERE LOWER(name) = LOWER($1) OR LOWER(COALESCE(name_en, '')) = LOWER($1)
		ORDER BY source_priority, id
		LIMIT 1
	`)).
					WithArgs("honey").
					WillReturnRows(rows)

				got, err := repo.FindExactByName(context.Background(), "honey")
				require.NoError(t, err)
				require.Equal(t, int64(133), got.ID)
				require.Equal(t, "Мёд", got.Name)
				require.InDelta(t, 304, got.Calories, 0.000001)
				require.InDelta(t, 0.3, got.Protein, 0.000001)
				require.InDelta(t, 0, got.Fat, 0.000001)
				require.InDelta(t, 82.4, got.Carbs, 0.000001)
			},
		},
		{
			name: "create saves custom product",
			run: func(t *testing.T, repo *PostgresBaseProductRepository, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "calories", "protein", "fat", "carbs"}).
					AddRow(int64(7), "custom oats", 400.0, 10.0, 7.0, 70.0)

				mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO base_products(name, calories, protein, fat, carbs)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, calories, protein, fat, carbs
	`)).
					WithArgs("custom oats", 400.0, 10.0, 7.0, 70.0).
					WillReturnRows(rows)

				got, err := repo.Create(context.Background(), BaseProduct{
					Name:     "custom oats",
					Calories: 400,
					Protein:  10,
					Fat:      7,
					Carbs:    70,
				})
				require.NoError(t, err)
				require.Equal(t, int64(7), got.ID)
				require.Equal(t, "custom oats", got.Name)
				require.InDelta(t, 400, got.Calories, 0.000001)
				require.InDelta(t, 10, got.Protein, 0.000001)
				require.InDelta(t, 7, got.Fat, 0.000001)
				require.InDelta(t, 70, got.Carbs, 0.000001)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewPostgresBaseProductRepository(db)
			tt.run(t, repo, mock)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

package ingredient

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Repository interface {
	Search(ctx context.Context, query string) ([]string, error)
	FindBestMatch(ctx context.Context, name string) (Ingredient, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Search(ctx context.Context, query string) ([]string, error) {
	prefix := query + "%"
	rows, err := r.db.QueryContext(ctx, `
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
	`, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) FindBestMatch(ctx context.Context, name string) (Ingredient, error) {
	if r == nil || r.db == nil {
		return Ingredient{}, ErrNotFound
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Ingredient{}, ErrNotFound
	}
	like := "%" + strings.ReplaceAll(name, " ", "%") + "%"
	prefix := name + "%"

	var item Ingredient
	err := r.db.QueryRowContext(ctx, `
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
	`, like, prefix).Scan(
		&item.ID,
		&item.Name,
		&item.Per100g.Calories,
		&item.Per100g.Protein,
		&item.Per100g.Fat,
		&item.Per100g.Carbs,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Ingredient{}, ErrNotFound
	}
	if err != nil {
		return Ingredient{}, err
	}
	return item, nil
}

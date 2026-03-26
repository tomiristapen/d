package product

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"back/internal/platform/errcode"
)

type BaseProduct struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Calories float64 `json:"calories"`
	Protein  float64 `json:"protein"`
	Fat      float64 `json:"fat"`
	Carbs    float64 `json:"carbs"`
}

type BaseProductRepository interface {
	Search(ctx context.Context, query string, limit int) ([]BaseProduct, error)
	FindExactByName(ctx context.Context, name string) (BaseProduct, error)
	FindFuzzyByName(ctx context.Context, name string) ([]BaseProduct, error)
	Create(ctx context.Context, p BaseProduct) (BaseProduct, error)
}

type NoopBaseProductRepository struct{}

func NewNoopBaseProductRepository() *NoopBaseProductRepository { return &NoopBaseProductRepository{} }

func (r *NoopBaseProductRepository) Search(_ context.Context, _ string, _ int) ([]BaseProduct, error) {
	return nil, nil
}

func (r *NoopBaseProductRepository) FindExactByName(_ context.Context, _ string) (BaseProduct, error) {
	return BaseProduct{}, errcode.NotFound
}

func (r *NoopBaseProductRepository) FindFuzzyByName(_ context.Context, _ string) ([]BaseProduct, error) {
	return nil, nil
}

func (r *NoopBaseProductRepository) Create(_ context.Context, p BaseProduct) (BaseProduct, error) {
	return p, nil
}

type PostgresBaseProductRepository struct {
	db *sql.DB
}

func NewPostgresBaseProductRepository(db *sql.DB) *PostgresBaseProductRepository {
	return &PostgresBaseProductRepository{db: db}
}

func (r *PostgresBaseProductRepository) Search(ctx context.Context, query string, limit int) ([]BaseProduct, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 12
	}
	like := "%" + strings.ReplaceAll(query, " ", "%") + "%"
	prefix := query + "%"

	rows, err := r.db.QueryContext(ctx, `
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
	`, like, prefix, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]BaseProduct, 0, limit)
	for rows.Next() {
		var p BaseProduct
		if err := rows.Scan(&p.ID, &p.Name, &p.Calories, &p.Protein, &p.Fat, &p.Carbs); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PostgresBaseProductRepository) FindExactByName(ctx context.Context, name string) (BaseProduct, error) {
	if r == nil || r.db == nil {
		return BaseProduct{}, errcode.NotFound
	}

	var item BaseProduct
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
		WHERE LOWER(name) = LOWER($1) OR LOWER(COALESCE(name_en, '')) = LOWER($1)
		ORDER BY source_priority, id
		LIMIT 1
	`, name).Scan(
		&item.ID,
		&item.Name,
		&item.Calories,
		&item.Protein,
		&item.Fat,
		&item.Carbs,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return BaseProduct{}, errcode.NotFound
	}
	if err != nil {
		return BaseProduct{}, err
	}
	return item, nil
}

func (r *PostgresBaseProductRepository) FindFuzzyByName(ctx context.Context, name string) ([]BaseProduct, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	return r.Search(ctx, name, 10)
}

func (r *PostgresBaseProductRepository) Create(ctx context.Context, p BaseProduct) (BaseProduct, error) {
	if r == nil || r.db == nil {
		return BaseProduct{}, errcode.NotFound
	}

	var saved BaseProduct
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO base_products(name, calories, protein, fat, carbs)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, calories, protein, fat, carbs
	`, p.Name, p.Calories, p.Protein, p.Fat, p.Carbs).Scan(
		&saved.ID,
		&saved.Name,
		&saved.Calories,
		&saved.Protein,
		&saved.Fat,
		&saved.Carbs,
	)
	if err != nil {
		return BaseProduct{}, err
	}
	return saved, nil
}

package product

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("product not found")

type Repository interface {
	GetByBarcode(ctx context.Context, barcode string) (Product, error)
	Upsert(ctx context.Context, p Product) (Product, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetByBarcode(ctx context.Context, barcode string) (Product, error) {
	var p Product
	var ingredientsJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT id, barcode, name, brand, ingredients, calories, protein, fat, carbohydrates, confidence_score, source, created_at, updated_at
		FROM products
		WHERE barcode = $1
	`, barcode).Scan(
		&p.ID,
		&p.Barcode,
		&p.Name,
		&p.Brand,
		&ingredientsJSON,
		&p.Calories,
		&p.Protein,
		&p.Fat,
		&p.Carbohydrates,
		&p.ConfidenceScore,
		&p.Source,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, err
	}
	if len(ingredientsJSON) > 0 {
		if err := json.Unmarshal(ingredientsJSON, &p.Ingredients); err != nil {
			return Product{}, fmt.Errorf("decode ingredients: %w", err)
		}
	}
	return p, nil
}

func (r *PostgresRepository) Upsert(ctx context.Context, p Product) (Product, error) {
	ingredientsJSON, err := json.Marshal(p.Ingredients)
	if err != nil {
		return Product{}, fmt.Errorf("encode ingredients: %w", err)
	}

	var saved Product
	var savedIngredientsJSON []byte
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO products(barcode, name, brand, ingredients, calories, protein, fat, carbohydrates, confidence_score, source, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW())
		ON CONFLICT (barcode) DO UPDATE SET
			name = EXCLUDED.name,
			brand = EXCLUDED.brand,
			ingredients = EXCLUDED.ingredients,
			calories = EXCLUDED.calories,
			protein = EXCLUDED.protein,
			fat = EXCLUDED.fat,
			carbohydrates = EXCLUDED.carbohydrates,
			confidence_score = EXCLUDED.confidence_score,
			source = EXCLUDED.source,
			updated_at = NOW()
		RETURNING id, barcode, name, brand, ingredients, calories, protein, fat, carbohydrates, confidence_score, source, created_at, updated_at
	`, p.Barcode, p.Name, p.Brand, ingredientsJSON, p.Calories, p.Protein, p.Fat, p.Carbohydrates, p.ConfidenceScore, p.Source).Scan(
		&saved.ID,
		&saved.Barcode,
		&saved.Name,
		&saved.Brand,
		&savedIngredientsJSON,
		&saved.Calories,
		&saved.Protein,
		&saved.Fat,
		&saved.Carbohydrates,
		&saved.ConfidenceScore,
		&saved.Source,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	if err != nil {
		return Product{}, err
	}
	if len(savedIngredientsJSON) > 0 {
		if err := json.Unmarshal(savedIngredientsJSON, &saved.Ingredients); err != nil {
			return Product{}, fmt.Errorf("decode ingredients: %w", err)
		}
	}
	return saved, nil
}

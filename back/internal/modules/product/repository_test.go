package product

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepositoryGetByBarcode_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, barcode, name, brand, ingredients, calories, protein, fat, carbohydrates, confidence_score, source, created_at, updated_at
		FROM products
		WHERE barcode = $1
	`)).WithArgs("123").WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByBarcode(context.Background(), "123")
	require.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepositoryUpsert_EncodesIngredients(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	ingredients := []string{"water", "sugar"}
	ingredientsJSON, err := json.Marshal(ingredients)
	require.NoError(t, err)

	now := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "barcode", "name", "brand", "ingredients", "calories", "protein", "fat", "carbohydrates", "confidence_score", "source", "created_at", "updated_at",
	}).AddRow(
		int64(1),
		"123",
		"Name",
		"Brand",
		ingredientsJSON,
		10.0,
		1.0,
		2.0,
		3.0,
		0.8,
		"openfoodfacts",
		now,
		now,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs("123", "Name", "Brand", ingredientsJSON, 10.0, 1.0, 2.0, 3.0, 0.8, "openfoodfacts").
		WillReturnRows(rows)

	saved, err := repo.Upsert(context.Background(), Product{
		Barcode:         "123",
		Name:            "Name",
		Brand:           "Brand",
		Ingredients:     ingredients,
		Calories:        10,
		Protein:         1,
		Fat:             2,
		Carbohydrates:   3,
		ConfidenceScore: 0.8,
		Source:          "openfoodfacts",
	})
	require.NoError(t, err)
	require.Equal(t, ingredients, saved.Ingredients)
	require.NoError(t, mock.ExpectationsWereMet())
}

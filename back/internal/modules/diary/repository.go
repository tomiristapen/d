package diary

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"back/internal/modules/nutrition"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	Create(ctx context.Context, e Entry) (Entry, error)
	GetDailyTotals(ctx context.Context, userID string, day time.Time) (DailyTotals, error)
	GetTargets(ctx context.Context, userID string) (nutrition.Nutrients, error)
	DeleteEntry(ctx context.Context, userID string, entryID int64) error
	UpdateEntry(ctx context.Context, userID string, input UpdateEntryInput) (Entry, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, e Entry) (Entry, error) {
	if r == nil || r.db == nil {
		return Entry{}, fmt.Errorf("db not configured")
	}

	ingredientsJSON, err := json.Marshal(e.Ingredients)
	if err != nil {
		return Entry{}, fmt.Errorf("encode ingredients: %w", err)
	}

	day, err := time.Parse("2006-01-02", e.EntryDate)
	if err != nil {
		return Entry{}, fmt.Errorf("invalid entry_date: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Entry{}, err
	}

	saved, err := insertEntry(ctx, tx, e, ingredientsJSON)
	if err != nil {
		if isUniqueViolation(err) && e.IdempotencyKey != "" {
			_ = tx.Rollback()
			return r.findByIdempotencyKey(ctx, e.UserID, e.IdempotencyKey)
		}
		_ = tx.Rollback()
		return Entry{}, err
	}

	if err := upsertDailyTotals(ctx, tx, e.UserID, day, nutrition.Nutrients{
		Calories: e.Calories,
		Protein:  e.Protein,
		Fat:      e.Fat,
		Carbs:    e.Carbs,
	}); err != nil {
		_ = tx.Rollback()
		return Entry{}, err
	}

	if err := tx.Commit(); err != nil {
		return Entry{}, err
	}
	return saved, nil
}

func (r *PostgresRepository) GetDailyTotals(ctx context.Context, userID string, day time.Time) (DailyTotals, error) {
	if r == nil || r.db == nil {
		return DailyTotals{}, fmt.Errorf("db not configured")
	}

	var totals DailyTotals
	var dbDay time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, date, calories, protein, fat, carbs, updated_at
		FROM daily_totals
		WHERE user_id = $1 AND date = $2
	`, userID, day.Format("2006-01-02")).Scan(
		&totals.UserID,
		&dbDay,
		&totals.Calories,
		&totals.Protein,
		&totals.Fat,
		&totals.Carbs,
		&totals.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return DailyTotals{
			UserID: userID,
			Date:   day.Format("2006-01-02"),
		}, nil
	}
	if err != nil {
		return DailyTotals{}, err
	}
	totals.Date = dbDay.Format("2006-01-02")
	return totals, nil
}

func (r *PostgresRepository) GetTargets(ctx context.Context, userID string) (nutrition.Nutrients, error) {
	if r == nil || r.db == nil {
		return nutrition.Nutrients{}, fmt.Errorf("db not configured")
	}

	var targets nutrition.Nutrients
	err := r.db.QueryRowContext(ctx, `
		SELECT calories, protein, fat, carbs
		FROM user_targets
		WHERE user_id = $1
	`, userID).Scan(
		&targets.Calories,
		&targets.Protein,
		&targets.Fat,
		&targets.Carbs,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nutrition.Nutrients{}, ErrTargetsNotFound
	}
	if err != nil {
		return nutrition.Nutrients{}, err
	}
	return targets, nil
}

func (r *PostgresRepository) DeleteEntry(ctx context.Context, userID string, entryID int64) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("db not configured")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	entry, day, err := findEntryForMutation(ctx, tx, userID, entryID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM diary_entries WHERE id = $1 AND user_id = $2`, entryID, userID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := addToDailyTotals(ctx, tx, userID, day, nutrition.Nutrients{
		Calories: -entry.Calories,
		Protein:  -entry.Protein,
		Fat:      -entry.Fat,
		Carbs:    -entry.Carbs,
	}); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) UpdateEntry(ctx context.Context, userID string, input UpdateEntryInput) (Entry, error) {
	if r == nil || r.db == nil {
		return Entry{}, fmt.Errorf("db not configured")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Entry{}, err
	}

	oldEntry, day, err := findEntryForMutation(ctx, tx, userID, input.EntryID)
	if err != nil {
		_ = tx.Rollback()
		return Entry{}, err
	}

	nextNutrients, err := nutrition.ScalePer100g(input.Per100G, input.AmountG)
	if err != nil {
		_ = tx.Rollback()
		return Entry{}, err
	}

	ingredientsJSON, err := json.Marshal(input.Ingredients)
	if err != nil {
		_ = tx.Rollback()
		return Entry{}, fmt.Errorf("encode ingredients: %w", err)
	}

	var saved Entry
	var savedIngredientsJSON []byte
	var entryDate time.Time
	err = tx.QueryRowContext(ctx, `
		UPDATE diary_entries
		SET
			source = $3,
			name = $4,
			amount_g = $5,
			calories = $6,
			protein = $7,
			fat = $8,
			carbs = $9,
			ingredients = $10
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, source, name, amount_g, calories, protein, fat, carbs, ingredients, entry_date, created_at
	`, input.EntryID, userID, input.Source, input.Name, input.AmountG, nextNutrients.Calories, nextNutrients.Protein, nextNutrients.Fat, nextNutrients.Carbs, ingredientsJSON).Scan(
		&saved.ID,
		&saved.UserID,
		&saved.Source,
		&saved.Name,
		&saved.AmountG,
		&saved.Calories,
		&saved.Protein,
		&saved.Fat,
		&saved.Carbs,
		&savedIngredientsJSON,
		&entryDate,
		&saved.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return Entry{}, ErrEntryNotFound
	}
	if err != nil {
		_ = tx.Rollback()
		return Entry{}, err
	}
	if err := json.Unmarshal(savedIngredientsJSON, &saved.Ingredients); err != nil {
		_ = tx.Rollback()
		return Entry{}, fmt.Errorf("decode ingredients: %w", err)
	}
	saved.EntryDate = entryDate.Format("2006-01-02")

	delta := nutrition.Subtract(
		nutrition.Nutrients{
			Calories: saved.Calories,
			Protein:  saved.Protein,
			Fat:      saved.Fat,
			Carbs:    saved.Carbs,
		},
		nutrition.Nutrients{
			Calories: oldEntry.Calories,
			Protein:  oldEntry.Protein,
			Fat:      oldEntry.Fat,
			Carbs:    oldEntry.Carbs,
		},
	)
	if err := addToDailyTotals(ctx, tx, userID, day, delta); err != nil {
		_ = tx.Rollback()
		return Entry{}, err
	}

	if err := tx.Commit(); err != nil {
		return Entry{}, err
	}
	return saved, nil
}

func insertEntry(ctx context.Context, tx *sql.Tx, e Entry, ingredientsJSON []byte) (Entry, error) {
	var saved Entry
	var savedIngredientsJSON []byte
	var entryDate time.Time
	err := tx.QueryRowContext(ctx, `
		INSERT INTO diary_entries(user_id, source, name, amount_g, calories, protein, fat, carbs, ingredients, entry_date, idempotency_key, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())
		RETURNING id, user_id, source, name, amount_g, calories, protein, fat, carbs, ingredients, entry_date, created_at
	`, e.UserID, e.Source, e.Name, e.AmountG, e.Calories, e.Protein, e.Fat, e.Carbs, ingredientsJSON, e.EntryDate, nullableString(e.IdempotencyKey)).Scan(
		&saved.ID,
		&saved.UserID,
		&saved.Source,
		&saved.Name,
		&saved.AmountG,
		&saved.Calories,
		&saved.Protein,
		&saved.Fat,
		&saved.Carbs,
		&savedIngredientsJSON,
		&entryDate,
		&saved.CreatedAt,
	)
	if err != nil {
		return Entry{}, err
	}
	if len(savedIngredientsJSON) > 0 {
		if err := json.Unmarshal(savedIngredientsJSON, &saved.Ingredients); err != nil {
			return Entry{}, fmt.Errorf("decode ingredients: %w", err)
		}
	}
	saved.EntryDate = entryDate.Format("2006-01-02")
	return saved, nil
}

func upsertDailyTotals(ctx context.Context, tx *sql.Tx, userID string, day time.Time, nutrients nutrition.Nutrients) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO daily_totals(user_id, date, calories, protein, fat, carbs, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, date) DO UPDATE SET
			calories = daily_totals.calories + EXCLUDED.calories,
			protein = daily_totals.protein + EXCLUDED.protein,
			fat = daily_totals.fat + EXCLUDED.fat,
			carbs = daily_totals.carbs + EXCLUDED.carbs,
			updated_at = NOW()
	`, userID, day.Format("2006-01-02"), nutrients.Calories, nutrients.Protein, nutrients.Fat, nutrients.Carbs)
	return err
}

func addToDailyTotals(ctx context.Context, tx *sql.Tx, userID string, day time.Time, delta nutrition.Nutrients) error {
	res, err := tx.ExecContext(ctx, `
		UPDATE daily_totals
		SET
			calories = GREATEST(0, daily_totals.calories + $3),
			protein = GREATEST(0, daily_totals.protein + $4),
			fat = GREATEST(0, daily_totals.fat + $5),
			carbs = GREATEST(0, daily_totals.carbs + $6),
			updated_at = NOW()
		WHERE user_id = $1 AND date = $2
	`, userID, day.Format("2006-01-02"), delta.Calories, delta.Protein, delta.Fat, delta.Carbs)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("daily_totals row not found")
	}
	return nil
}

func findEntryForMutation(ctx context.Context, tx *sql.Tx, userID string, entryID int64) (Entry, time.Time, error) {
	var entry Entry
	var day time.Time
	var ingredientsJSON []byte
	err := tx.QueryRowContext(ctx, `
		SELECT id, user_id, source, name, amount_g, calories, protein, fat, carbs, ingredients, entry_date, created_at
		FROM diary_entries
		WHERE id = $1 AND user_id = $2
	`, entryID, userID).Scan(
		&entry.ID,
		&entry.UserID,
		&entry.Source,
		&entry.Name,
		&entry.AmountG,
		&entry.Calories,
		&entry.Protein,
		&entry.Fat,
		&entry.Carbs,
		&ingredientsJSON,
		&day,
		&entry.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Entry{}, time.Time{}, ErrEntryNotFound
	}
	if err != nil {
		return Entry{}, time.Time{}, err
	}
	if err := json.Unmarshal(ingredientsJSON, &entry.Ingredients); err != nil {
		return Entry{}, time.Time{}, fmt.Errorf("decode ingredients: %w", err)
	}
	entry.EntryDate = day.Format("2006-01-02")
	return entry, day, nil
}

func (r *PostgresRepository) findByIdempotencyKey(ctx context.Context, userID string, key string) (Entry, error) {
	var saved Entry
	var ingredientsJSON []byte
	var entryDate time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, source, name, amount_g, calories, protein, fat, carbs, ingredients, entry_date, created_at
		FROM diary_entries
		WHERE user_id = $1 AND idempotency_key = $2
	`, userID, key).Scan(
		&saved.ID,
		&saved.UserID,
		&saved.Source,
		&saved.Name,
		&saved.AmountG,
		&saved.Calories,
		&saved.Protein,
		&saved.Fat,
		&saved.Carbs,
		&ingredientsJSON,
		&entryDate,
		&saved.CreatedAt,
	)
	if err != nil {
		return Entry{}, err
	}
	if err := json.Unmarshal(ingredientsJSON, &saved.Ingredients); err != nil {
		return Entry{}, fmt.Errorf("decode ingredients: %w", err)
	}
	saved.EntryDate = entryDate.Format("2006-01-02")
	return saved, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

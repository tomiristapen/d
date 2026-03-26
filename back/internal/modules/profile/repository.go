package profile

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
)

type Repository interface {
	Save(ctx context.Context, profile Profile) error
	GetByUserID(ctx context.Context, userID string) (Profile, error)
	Delete(ctx context.Context, userID string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, profile Profile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO profiles(user_id, age, gender, height_cm, weight_kg, activity_level, goal, diet_type, religious_restriction, intolerances, custom_allergies, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (user_id) DO UPDATE SET
			age = EXCLUDED.age,
			gender = EXCLUDED.gender,
			height_cm = EXCLUDED.height_cm,
			weight_kg = EXCLUDED.weight_kg,
			activity_level = EXCLUDED.activity_level,
			goal = EXCLUDED.goal,
			diet_type = EXCLUDED.diet_type,
			religious_restriction = EXCLUDED.religious_restriction,
			intolerances = EXCLUDED.intolerances,
			custom_allergies = EXCLUDED.custom_allergies,
			updated_at = EXCLUDED.updated_at
	`, profile.UserID, profile.Age, profile.Gender, profile.HeightCM, profile.WeightKG, profile.ActivityLevel, profile.Goal, profile.DietType, profile.ReligiousRestriction, stringArray(profile.Intolerances), stringArray(profile.CustomAllergies), profile.CreatedAt, profile.UpdatedAt); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM profile_allergies WHERE user_id = $1`, profile.UserID); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, allergy := range profile.Allergies {
		if _, err := tx.ExecContext(ctx, `INSERT INTO profile_allergies(user_id, allergy) VALUES ($1, $2)`, profile.UserID, allergy); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO user_targets(user_id, calories, protein, fat, carbs, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			calories = EXCLUDED.calories,
			protein = EXCLUDED.protein,
			fat = EXCLUDED.fat,
			carbs = EXCLUDED.carbs,
			updated_at = NOW()
	`, profile.UserID, profile.Targets.Calories, profile.Targets.Protein, profile.Targets.Fat, profile.Targets.Carbs); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE users SET profile_completed = TRUE, updated_at = NOW() WHERE id = $1`, profile.UserID); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetByUserID(ctx context.Context, userID string) (Profile, error) {
	var profile Profile
	var intolerancesJSON []byte
	var customAllergiesJSON []byte
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			user_id,
			age,
			gender,
			height_cm,
			weight_kg,
			activity_level,
			goal,
			diet_type,
			religious_restriction,
			to_json(intolerances) AS intolerances_json,
			to_json(custom_allergies) AS custom_allergies_json,
			created_at,
			updated_at
		FROM profiles
		WHERE user_id = $1
	`, userID).Scan(
		&profile.UserID,
		&profile.Age,
		&profile.Gender,
		&profile.HeightCM,
		&profile.WeightKG,
		&profile.ActivityLevel,
		&profile.Goal,
		&profile.DietType,
		&profile.ReligiousRestriction,
		&intolerancesJSON,
		&customAllergiesJSON,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Profile{}, ErrNotFound
		}
		return Profile{}, err
	}
	if err := json.Unmarshal(intolerancesJSON, &profile.Intolerances); err != nil {
		return Profile{}, err
	}
	if err := json.Unmarshal(customAllergiesJSON, &profile.CustomAllergies); err != nil {
		return Profile{}, err
	}

	rows, err := r.db.QueryContext(ctx, `SELECT allergy FROM profile_allergies WHERE user_id = $1 ORDER BY allergy`, userID)
	if err != nil {
		return Profile{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var allergy string
		if err := rows.Scan(&allergy); err != nil {
			return Profile{}, err
		}
		profile.Allergies = append(profile.Allergies, allergy)
	}
	if err := rows.Err(); err != nil {
		return Profile{}, err
	}

	return profile, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM profiles WHERE user_id = $1`, userID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if affected == 0 {
		_ = tx.Rollback()
		return ErrNotFound
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM profile_allergies WHERE user_id = $1`, userID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_targets WHERE user_id = $1`, userID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET profile_completed = FALSE, updated_at = NOW() WHERE id = $1`, userID); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type stringArray []string

func (a stringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	escaped := make([]string, 0, len(a))
	for _, item := range a {
		item = strings.ReplaceAll(item, `"`, `\"`)
		escaped = append(escaped, `"`+item+`"`)
	}
	return "{" + strings.Join(escaped, ",") + "}", nil
}

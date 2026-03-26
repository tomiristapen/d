ALTER TABLE profiles
    ALTER COLUMN height_cm TYPE DOUBLE PRECISION USING height_cm::DOUBLE PRECISION,
    ALTER COLUMN weight_kg TYPE DOUBLE PRECISION USING weight_kg::DOUBLE PRECISION;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'profiles' AND column_name = 'nutrition_goal'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'profiles' AND column_name = 'goal'
    ) THEN
        ALTER TABLE profiles RENAME COLUMN nutrition_goal TO goal;
    END IF;
END $$;

UPDATE profiles
SET activity_level = CASE activity_level
    WHEN 'low' THEN 'sedentary'
    WHEN 'moderate' THEN 'moderate'
    WHEN 'high' THEN 'active'
    ELSE activity_level
END
WHERE activity_level IN ('low', 'moderate', 'high');

UPDATE profiles
SET goal = CASE goal
    WHEN 'lose_weight' THEN 'lose'
    WHEN 'maintain_weight' THEN 'maintain'
    WHEN 'gain_weight' THEN 'gain'
    WHEN 'healthy_eating' THEN 'maintain'
    ELSE goal
END
WHERE goal IN ('lose_weight', 'maintain_weight', 'gain_weight', 'healthy_eating');

CREATE TABLE IF NOT EXISTS user_targets (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    calories DOUBLE PRECISION NOT NULL DEFAULT 0,
    protein DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS daily_totals (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    calories DOUBLE PRECISION NOT NULL DEFAULT 0,
    protein DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, date)
);

ALTER TABLE diary_entries
    ADD COLUMN IF NOT EXISTS entry_date DATE;

UPDATE diary_entries
SET entry_date = DATE(created_at)
WHERE entry_date IS NULL;

ALTER TABLE diary_entries
    ALTER COLUMN entry_date SET NOT NULL;

ALTER TABLE diary_entries
    ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_diary_entries_user_idempotency_key
    ON diary_entries(user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_daily_totals_user_date
    ON daily_totals(user_id, date DESC);

INSERT INTO user_targets (user_id, calories, protein, fat, carbs, created_at, updated_at)
SELECT
    p.user_id,
    ROUND((
        (
            CASE
                WHEN p.gender = 'male' THEN (10 * p.weight_kg) + (6.25 * p.height_cm) - (5 * p.age) + 5
                WHEN p.gender = 'female' THEN (10 * p.weight_kg) + (6.25 * p.height_cm) - (5 * p.age) - 161
                ELSE 0
            END
        ) * (
            CASE
                WHEN p.activity_level = 'sedentary' THEN 1.2
                WHEN p.activity_level = 'light' THEN 1.375
                WHEN p.activity_level = 'moderate' THEN 1.55
                WHEN p.activity_level = 'active' THEN 1.725
                ELSE 1.55
            END
        ) + (
            CASE
                WHEN p.goal = 'lose' THEN -500
                WHEN p.goal = 'gain' THEN 500
                ELSE 0
            END
        )
    )),
    ROUND(p.weight_kg * 1.8, 2),
    ROUND(p.weight_kg * 0.9, 2),
    GREATEST(
        0,
        ROUND((
            (
                (
                    CASE
                        WHEN p.gender = 'male' THEN (10 * p.weight_kg) + (6.25 * p.height_cm) - (5 * p.age) + 5
                        WHEN p.gender = 'female' THEN (10 * p.weight_kg) + (6.25 * p.height_cm) - (5 * p.age) - 161
                        ELSE 0
                    END
                ) * (
                    CASE
                        WHEN p.activity_level = 'sedentary' THEN 1.2
                        WHEN p.activity_level = 'light' THEN 1.375
                        WHEN p.activity_level = 'moderate' THEN 1.55
                        WHEN p.activity_level = 'active' THEN 1.725
                        ELSE 1.55
                    END
                ) + (
                    CASE
                        WHEN p.goal = 'lose' THEN -500
                        WHEN p.goal = 'gain' THEN 500
                        ELSE 0
                    END
                )
            ) - ((p.weight_kg * 1.8) * 4) - ((p.weight_kg * 0.9) * 9)
        ) / 4, 2)
    ),
    NOW(),
    NOW()
FROM profiles p
WHERE p.gender IN ('male', 'female')
  AND p.activity_level IN ('sedentary', 'light', 'moderate', 'active')
  AND p.goal IN ('lose', 'maintain', 'gain')
ON CONFLICT (user_id) DO UPDATE SET
    calories = EXCLUDED.calories,
    protein = EXCLUDED.protein,
    fat = EXCLUDED.fat,
    carbs = EXCLUDED.carbs,
    updated_at = NOW();

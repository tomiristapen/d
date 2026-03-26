ALTER TABLE base_products
    ADD COLUMN IF NOT EXISTS calories DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS protein DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fat DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS carbs DOUBLE PRECISION NOT NULL DEFAULT 0;

UPDATE base_products AS bp
SET
    calories = i.calories,
    protein = i.protein,
    fat = i.fat,
    carbs = i.carbs
FROM ingredients AS i
WHERE bp.name = i.name;

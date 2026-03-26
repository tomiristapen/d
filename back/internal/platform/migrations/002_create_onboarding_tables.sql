CREATE TABLE IF NOT EXISTS profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    age INT NOT NULL,
    gender TEXT NOT NULL,
    height_cm INT NOT NULL,
    weight_kg INT NOT NULL,
    nutrition_goal TEXT NOT NULL,
    diet_type TEXT NOT NULL,
    religious_restriction TEXT NOT NULL,
    intolerances TEXT[] NOT NULL DEFAULT '{}',
    custom_allergies TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS profile_allergies (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    allergy TEXT NOT NULL,
    PRIMARY KEY (user_id, allergy)
);

CREATE TABLE IF NOT EXISTS ingredients (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

INSERT INTO ingredients(name) VALUES
    ('honey'),
    ('mango'),
    ('peanuts'),
    ('milk'),
    ('eggs'),
    ('fish'),
    ('soy'),
    ('wheat'),
    ('sesame')
ON CONFLICT (name) DO NOTHING;

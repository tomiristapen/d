CREATE TABLE IF NOT EXISTS base_products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_base_products_name ON base_products(name);

-- Minimal seed so fuzzy-matching has something to link to.
INSERT INTO base_products(name)
SELECT name FROM ingredients
ON CONFLICT (name) DO NOTHING;


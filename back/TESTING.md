# Testing

## Go tests (fast, local)

From `back/`:

```powershell
$env:GOCACHE="$PWD\\.gocache"
$env:GOTMPDIR="$PWD\\.gotmp"
$env:GOPATH="$PWD\\.gopath"
$env:GOMODCACHE="$PWD\\.gopath\\pkg\\mod"
go test ./...
```

## QA assets

- Auth/profile/onboarding Postman env: `back/qa/postman_auth_profile_onboarding_environment.json`
- Auth/profile/onboarding Postman collection: `back/qa/postman_auth_profile_onboarding_collection.json`
- Auth/profile/onboarding test matrix: `back/qa/auth_profile_onboarding_test_matrix.md`

## Manual + recipe API tests (Postman)

Import `back/postman_manual_recipe_collection.json`.

### Setup

- Set `base_url` (for Docker Compose default): `http://localhost:8080`
- Set `test_email` and `test_password`
- If the user already exists and is verified, run `Auth / POST /api/v1/auth/login`
- If the user does not exist yet, run `Auth / POST /api/v1/auth/register`, then `Auth / POST /api/v1/auth/send-verification-code`, paste the code into `verify_code`, run `Auth / POST /api/v1/auth/verify-email`, and then run `Auth / POST /api/v1/auth/login`
- If the user signs in with Google, run `Auth / POST /api/v1/auth/google`, and if they want password login later too, run `Auth / POST /api/v1/auth/set-password (auth required)`
- Run the whole `Setup - deterministic custom products` folder from top to bottom once
- Then run the `Manual - ...` and `Recipe - ...` folders in any order

### Notes

- The first setup request resets a per-run suffix and all later setup requests depend on it, so run the setup folder in order.
- Most positive manual/recipe cases in this collection use deterministic custom products created by the setup folder, so they do not depend on local seed data.
- Some requests intentionally send malformed JSON or omit `Content-Type` because those cases are covered by backend tests too.

## General API tests (legacy collection)

Import `back/postman_collection.json` if you still need the older broader collection.

### Product tests

`GET /products/by-barcode/{{barcode_ok}}` expects the barcode to already exist in your local DB.

If you want a deterministic "DB-only" product test, insert a product via `psql`:

```powershell
docker compose exec db psql -U postgres -d app
```

```sql
INSERT INTO products (barcode, name, brand, ingredients, calories, protein, fat, carbohydrates, confidence_score, source)
VALUES ('4870028002852','Kazakhstanskiy Milk Chocolate','Bayan Sulu','["sugar","cocoa butter"]'::jsonb,565,9.7,36.8,48.7,1,'manual')
ON CONFLICT (barcode) DO UPDATE SET
  name=EXCLUDED.name,
  brand=EXCLUDED.brand,
  ingredients=EXCLUDED.ingredients,
  calories=EXCLUDED.calories,
  protein=EXCLUDED.protein,
  fat=EXCLUDED.fat,
  carbohydrates=EXCLUDED.carbohydrates,
  confidence_score=EXCLUDED.confidence_score,
  source=EXCLUDED.source,
  updated_at=NOW();
```

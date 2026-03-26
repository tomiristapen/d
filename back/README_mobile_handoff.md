# Mobile Handoff README

This document is for the developer who will build the mobile client from scratch.

Current frontend scope:

- registration
- email verification
- email/password login
- login with email code
- Google login
- password setup after Google/code login
- password reset
- profile
- onboarding

Out of scope for now:

- barcode
- OCR
- manual food
- recipes
- diary

## 1. Source of truth

Primary API contract:

- Swagger file in repo:
  - `back/internal/platform/httpapi/swagger_assets/openapi.yaml`
- Swagger UI when backend is running:
  - `http://localhost:8080/swagger`
- Raw OpenAPI spec when backend is running:
  - `http://localhost:8080/swagger/openapi.yaml`

Current registered routes are defined in:

- `back/internal/platform/httpapi/router.go`

## 2. How to run backend

From repo root:

```powershell
cd back
Copy-Item .env.example .env
docker compose up --build
```

Health check:

```text
http://localhost:8080/api/v1/healthz
```

Expected response:

```json
{"status":"ok"}
```

Stop backend:

```powershell
cd back
docker compose down
```

## 3. Backend env

Backend env file:

- `back/.env`

Template:

- `back/.env.example`

Minimum required values:

- `DATABASE_URL`
- `JWT_ACCESS_SECRET`
- `JWT_REFRESH_SECRET`

Important auth-related values:

- `ACCESS_TOKEN_TTL=15m`
- `REFRESH_TOKEN_TTL=168h`
- `JWT_ISSUER=back-app`
- `GOOGLE_CLIENT_ID=...`

Notes:

- `GOOGLE_CLIENT_ID` can contain more than one client id separated by commas.
- This is useful when backend must accept Android, iOS, and Web Google ID tokens.
- For password reset / verification emails, SMTP settings are used if configured.

## 4. Suggested frontend env

The mobile app does not have to match any existing local mobile implementation.
If you are creating a new mobile client, the frontend should usually have these env values:

- `API_BASE_URL`
- `GOOGLE_CLIENT_ID`

Recommended examples:

- Android emulator:
  - `API_BASE_URL=http://10.0.2.2:8080`
- iOS simulator:
  - `API_BASE_URL=http://localhost:8080`
- Real Android/iPhone device on same Wi-Fi:
  - `API_BASE_URL=http://<YOUR_PC_IPV4>:8080`

Google sign-in:

- frontend should use a Google client id that can return an `id_token`
- backend `GOOGLE_CLIENT_ID` must include the client id audience that will appear in that token

## 5. Platform notes: Android vs iOS

Backend is the same for Android and iOS.

The API, payloads, tokens, onboarding logic, and profile logic do not change by platform.

Differences are only on the mobile side:

- local networking during development
- Google sign-in platform setup
- build/signing/tooling

In practice:

- Android emulator usually uses `10.0.2.2`
- iOS simulator usually uses `localhost`
- real devices use laptop LAN IP
- iOS builds require macOS + Xcode
- Android builds require Android SDK / Android Studio

So yes: Android and iOS differ mostly in the mobile implementation layer, not in backend behavior.

## 6. Auth and onboarding state machine

The app should treat the user lifecycle as a state machine:

1. `registered`
2. `email_verified`
3. `logged_in`
4. `profile_completed`

This is the most important contract to implement correctly.

Useful response fields:

- auth endpoints return:
  - `access_token`
  - `refresh_token`
  - `profile_completed`
  - `has_password`

Meaning:

- `profile_completed=false` -> send user into onboarding/profile flow
- `has_password=false` -> account exists but no email/password login yet, so frontend can offer `Set password`

## 7. Endpoints in current mobile scope

### Public

- `GET /api/v1/healthz`
- `GET /api/v1/onboarding/options`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/send-verification-code`
- `POST /api/v1/auth/verify-email`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/send-login-code`
- `POST /api/v1/auth/login-with-code`
- `POST /api/v1/auth/send-password-reset-code`
- `POST /api/v1/auth/reset-password`
- `POST /api/v1/auth/google`
- `POST /api/v1/auth/refresh`

### Requires Bearer token

- `POST /api/v1/auth/set-password`
- `GET /api/v1/profile`
- `PUT /api/v1/profile`
- `POST /api/v1/profile/reset`
- `GET /api/v1/onboarding/status`

## 8. How endpoints are connected

### Flow A: classic email flow

1. `POST /auth/register`
2. `POST /auth/verify-email`
3. `POST /auth/login`
4. if `profile_completed=false` -> onboarding screens
5. `GET /onboarding/options`
6. `PUT /profile`
7. `GET /onboarding/status` should now be `true`

### Flow B: login with code

1. `POST /auth/send-login-code`
2. `POST /auth/login-with-code`
3. if `has_password=false`, optionally offer `POST /auth/set-password`

### Flow C: Google login

1. `POST /auth/google`
2. if `has_password=false`, optionally offer `POST /auth/set-password`
3. later user can also log in via email/password

### Flow D: password reset

1. `POST /auth/send-password-reset-code`
2. `POST /auth/reset-password`
3. tokens are returned immediately
4. user does not need a separate login call right after reset

## 9. Important endpoint behavior

### `POST /api/v1/auth/register`

Request:

```json
{
  "email": "person@example.com",
  "password": "StrongPass1!",
  "confirm_password": "StrongPass1!"
}
```

Success:

- `201`
- body:

```json
{
  "status": "verification_sent"
}
```

### `POST /api/v1/auth/login`

Request:

```json
{
  "email": "person@example.com",
  "password": "StrongPass1!"
}
```

Success response shape:

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "profile_completed": false,
  "has_password": true
}
```

### `POST /api/v1/auth/google`

Request:

```json
{
  "id_token": "<google_id_token>"
}
```

Success response shape is the same `AuthResponse`.

### `POST /api/v1/auth/set-password`

Authenticated request:

```json
{
  "password": "StrongPass1!",
  "confirm_password": "StrongPass1!"
}
```

Success:

```json
{
  "status": "password_set"
}
```

### `GET /api/v1/onboarding/options`

This endpoint is the source of truth for dropdown/toggle values.

Frontend should not hardcode these option lists if it can avoid it.

It returns:

- `genders`
- `activity_levels`
- `nutrition_goals`
- `allergies`
- `intolerances`
- `diet_types`
- `religious_restrictions`

Each item shape:

```json
{
  "key": "milk",
  "label": "Milk"
}
```

### `PUT /api/v1/profile`

Important fields:

- `age`
- `gender`
- `height_cm`
- `weight_kg`
- `activity_level`
- `nutrition_goal`
- `allergies`
- `custom_allergies`
- `intolerances`
- `diet_type`
- `religious_restriction`

Important rule:

- `allergies` must contain canonical backend-supported keys from `GET /onboarding/options`
- if user enters something outside the canonical list, put it into `custom_allergies`

Example:

```json
{
  "age": 28,
  "gender": "female",
  "height_cm": 168,
  "weight_kg": 60,
  "activity_level": "moderate",
  "nutrition_goal": "healthy_eating",
  "allergies": ["milk"],
  "custom_allergies": ["honey"],
  "intolerances": ["lactose"],
  "diet_type": "vegetarian",
  "religious_restriction": "none"
}
```

## 10. Token lifecycle

Current backend defaults:

- access token: `15 minutes`
- refresh token: `7 days`

Frontend should:

1. store both tokens securely
2. attach access token to protected endpoints
3. on `401`, try `POST /auth/refresh`
4. replace both stored tokens with the refreshed pair
5. replace the previous refresh token too, because refresh tokens are rotated and single-use
6. only send user to login if refresh also fails

## 11. Recommended screens for current phase

Recommended mobile screens:

- Welcome
- Register
- Verify email
- Login
- Login with code
- Forgot password / Reset password
- Google login
- Set password
- Profile setup
- Dietary preferences / onboarding
- Profile view / edit

Not needed yet:

- food search
- barcode
- OCR
- recipes
- diary

## 12. Recommended implementation order

1. health check connectivity
2. register
3. verify email
4. login
5. token storage + refresh
6. onboarding options
7. profile create/update
8. onboarding status gate
9. Google login
10. set password
11. password reset

## 13. QA assets

If the mobile developer wants ready-made backend QA assets, use:

- Postman collection:
  - `back/qa/postman_auth_profile_onboarding_collection.json`
- Postman environment:
  - `back/qa/postman_auth_profile_onboarding_environment.json`
- Test matrix:
  - `back/qa/auth_profile_onboarding_test_matrix.md`

## 14. Current limitations / current behavior

- refresh tokens are stateful, rotated on refresh, and cannot be reused after successful refresh
- verify-email is currently idempotent for already-verified users
- profile validation enforces realistic ranges: `age 1-120`, `height_cm 30-300`, `weight_kg 2-500`
- food-related endpoints exist in backend, but they are not part of the current mobile implementation scope

## 15. Future barcode UX guidance

If barcode flow is added later, recommended mobile behavior:

- normalize scanned barcode before request:
  - keep digits only
  - ignore spaces, hyphens, and scanner formatting noise
  - treat UPC-A / EAN-13 zero-padded variants as the same lookup family
- after successful scan, show:
  - product name
  - nutrients per `100 g`
  - an input for `how much was eaten`
  - recalculated nutrients for the entered amount

Do not add a multi-product picker by barcode in v1 by default.

Why:

- barcode UX should usually feel deterministic: one scan -> one best product
- external datasets can sometimes contain duplicate or conflicting records for the same barcode, but that is usually a data quality problem, not a user choice problem
- showing several products after every scan adds friction and confusion for the common case

Recommended rule:

- first return one best product for the barcode
- only consider a manual choice UI later if backend starts surfacing true ambiguity from multiple trusted sources

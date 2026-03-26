# Auth / Profile / Onboarding Test Matrix

## Scope

This matrix covers the user identity state machine and the core profile/onboarding lifecycle:

- `registered`
- `verified`
- `logged_in`
- `profile_completed`

## Legend

- `Go` = automated Go tests under `back/internal/modules/...`
- `Postman` = request included in Postman collection
- `Manual` = requires manual setup/observation, usually email codes or Google token

## Register

| ID | Endpoint | Scenario | Type | Expected | Coverage |
| --- | --- | --- | --- | --- | --- |
| TC-AUTH-REG-001 | `POST /auth/register` | Valid email + password + confirm | Positive | `201`, `status=verification_sent` | Go, Postman |
| TC-AUTH-REG-002 | `POST /auth/register` | Email with subdomain | Positive | `201` | Go, Postman |
| TC-AUTH-REG-003 | `POST /auth/register` | Email already exists | Negative | `400` | Go, Postman |
| TC-AUTH-REG-004 | `POST /auth/register` | Password != confirm | Negative | `400` | Go, Postman |
| TC-AUTH-REG-005 | `POST /auth/register` | Invalid email (`abc`, `test@`) | Negative | `400` | Go, Postman |
| TC-AUTH-REG-006 | `POST /auth/register` | Empty fields | Negative | `400` | Go, Postman |
| TC-AUTH-REG-007 | `POST /auth/register` | Email length > 254 | Edge | `400` | Go, Postman |
| TC-AUTH-REG-008 | `POST /auth/register` | Password min length boundary | Boundary | `min-1 fail`, `min pass` | Go, Postman |
| TC-AUTH-REG-009 | `POST /auth/register` | Password with special chars | Edge | `201` | Go, Postman |
| TC-AUTH-REG-010 | `POST /auth/register` | Password max length boundary | Boundary | `max pass`, `max+1 fail` | Go, Postman |

## Verify Email

| ID | Endpoint | Scenario | Type | Expected | Coverage |
| --- | --- | --- | --- | --- | --- |
| TC-AUTH-VER-001 | `POST /auth/verify-email` | Correct code | Positive | `200` | Go, Postman |
| TC-AUTH-VER-002 | `POST /auth/verify-email` | Wrong code | Negative | `400` | Go, Postman |
| TC-AUTH-VER-003 | `POST /auth/verify-email` | Expired code | Negative | `400` | Go, Postman, Manual |
| TC-AUTH-VER-004 | `POST /auth/verify-email` | Email not registered | Negative | `400` | Go, Postman |
| TC-AUTH-VER-005 | `POST /auth/verify-email` | Verify already-verified email again | Edge | `200` idempotent | Go, Postman |

## Login

| ID | Endpoint | Scenario | Type | Expected | Coverage |
| --- | --- | --- | --- | --- | --- |
| TC-AUTH-LOGIN-001 | `POST /auth/login` | Correct email + password | Positive | `200`, tokens returned | Go, Postman |
| TC-AUTH-LOGIN-002 | `POST /auth/login` | Wrong password | Negative | `401` | Go, Postman |
| TC-AUTH-LOGIN-003 | `POST /auth/login` | Email not found | Negative | `401` | Go, Postman |
| TC-AUTH-LOGIN-004 | `POST /auth/login` | Email not verified | Negative | `401` | Go, Postman |
| TC-AUTH-LOGIN-005 | `POST /auth/login` | Login after reset password | Edge | `200` | Go, Postman |

## Login With Code

| ID | Endpoint | Scenario | Type | Expected | Coverage |
| --- | --- | --- | --- | --- | --- |
| TC-AUTH-CODE-001 | `POST /auth/send-login-code` + `POST /auth/login-with-code` | Correct code | Positive | `200` | Go, Postman |
| TC-AUTH-CODE-002 | `POST /auth/login-with-code` | Wrong code | Negative | `401` | Go, Postman |
| TC-AUTH-CODE-003 | `POST /auth/login-with-code` | Expired code | Negative | `401` | Go, Postman, Manual |
| TC-AUTH-CODE-004 | `POST /auth/login-with-code` | Reuse same code twice | Edge | Second attempt fails | Go, Postman |
| TC-AUTH-CODE-005 | `POST /auth/send-login-code` twice | Last code wins | Edge | Old code fails, latest works | Go, Postman, Manual |

## Refresh Token

| ID | Endpoint | Scenario | Type | Expected | Coverage |
| --- | --- | --- | --- | --- | --- |
| TC-AUTH-REFRESH-001 | `POST /auth/refresh` | Valid refresh token | Positive | `200`, new auth response | Go, Postman |
| TC-AUTH-REFRESH-002 | `POST /auth/refresh` | Invalid refresh token | Negative | `401` | Go, Postman |
| TC-AUTH-REFRESH-003 | `POST /auth/refresh` | Expired refresh token | Negative | `401` | Go, Postman |
| TC-AUTH-REFRESH-004 | `POST /auth/refresh` | Reuse same refresh token | Edge | Second use fails with `401` | Go, Postman |

## Set Password / Reset Password / Google

| ID | Endpoint | Scenario | Type | Expected | Coverage |
| --- | --- | --- | --- | --- | --- |
| TC-AUTH-SETPWD-001 | `POST /auth/set-password` | Set password after Google/code login | Positive | `200` | Go, Postman |
| TC-AUTH-SETPWD-002 | `POST /auth/set-password` | Password != confirm | Negative | `400` | Go, Postman |
| TC-AUTH-SETPWD-003 | `POST /auth/set-password` | No auth header | Negative | `401` | Go, Postman |
| TC-AUTH-SETPWD-004 | `POST /auth/set-password` | Overwrite existing password | Edge | Currently allowed | Go, Postman |
| TC-AUTH-GOOGLE-001 | `POST /auth/google` | Valid Google id token | Positive | `200` | Go, Postman, Manual |
| TC-AUTH-GOOGLE-002 | `POST /auth/google` | Invalid Google id token | Negative | `401` | Go, Postman |
| TC-AUTH-GOOGLE-003 | `POST /auth/google` -> `POST /auth/set-password` -> `POST /auth/login` | Google account gets email login | Flow | Full flow succeeds | Go, Postman |
| TC-AUTH-RESET-001 | `POST /auth/send-password-reset-code` -> `POST /auth/reset-password` | Reset password | Positive | `200`, tokens returned | Go, Postman |
| TC-AUTH-RESET-002 | `POST /auth/reset-password` | Wrong or expired code | Negative | `400` | Go, Postman |

## Profile / Onboarding

| ID | Endpoint | Scenario | Type | Expected | Coverage |
| --- | --- | --- | --- | --- | --- |
| TC-PROFILE-GET-001 | `GET /profile` | Authenticated existing profile | Positive | `200` | Postman |
| TC-PROFILE-GET-002 | `GET /profile` | No auth | Negative | `401` | Postman |
| TC-PROFILE-PUT-001 | `PUT /profile` | Valid full payload | Positive | `200` | Go, Postman |
| TC-PROFILE-PUT-002 | `PUT /profile` | Partial update on existing profile | Edge | `200` | Go, Postman |
| TC-PROFILE-PUT-003 | `PUT /profile` | `age=0` | Negative | `400` | Go, Postman |
| TC-PROFILE-PUT-004 | `PUT /profile` | `age=1` | Boundary | `200` | Go, Postman |
| TC-PROFILE-PUT-005 | `PUT /profile` | Invalid enum `gender=abc` | Negative | `400` | Go, Postman |
| TC-PROFILE-PUT-006 | `PUT /profile` | Empty body on first create | Negative | `400` | Go, Postman |
| TC-PROFILE-PUT-007 | `PUT /profile` | Very large height | Edge | `400` | Go, Postman |
| TC-PROFILE-RESET-001 | `POST /profile/reset` | Authenticated reset | Positive | `200` | Postman |
| TC-PROFILE-RESET-002 | `POST /profile/reset` | No auth | Negative | `401` | Postman |
| TC-ONBOARDING-001 | `GET /onboarding/status` | After register / before profile | Positive | `false` | Go, Postman |
| TC-ONBOARDING-002 | `GET /onboarding/status` | After profile save | Positive | `true` | Go, Postman |
| TC-ONBOARDING-003 | `GET /onboarding/options` | Public options catalog | Positive | `200`, options arrays returned | Go, Postman |

## End-to-End Flows

| ID | Flow | Steps | Expected | Coverage |
| --- | --- | --- | --- | --- |
| FLOW-001 | Classic email flow | register -> verify -> login -> put profile -> onboarding/status | Profile completed becomes `true` | Go, Postman |
| FLOW-002 | Email code flow | register -> verify -> send-login-code -> login-with-code | Login succeeds with code | Go, Postman |
| FLOW-003 | Google fallback | google login -> set-password -> email login | Email login works after password setup | Go, Postman |
| FLOW-004 | Password reset | send-reset-code -> reset-password -> login | New password works | Go, Postman |

## Known Current-Behavior Notes

- Refresh tokens are rotated and become invalid after first successful use.
- Verify-email is currently idempotent for already-verified users.
- Profile validation enforces `age 1-120`, `height_cm 30-300`, and `weight_kg 2-500`.

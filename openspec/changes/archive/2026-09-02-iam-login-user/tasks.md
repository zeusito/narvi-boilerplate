# Tasks: In-House Passwordless User Login (Email OTP & Bearer Sessions)

## 1. Schema & Migration Fixes
- [x] 1.1 Fix syntax defect (trailing comma before `FROM`) in `db/migrations/20260902035257_iam_sessions.sql`.

## 2. Models & Data Structures
- [x] 2.1 Define identity and verification domain models in `internal/iam/identity_models.go` and `internal/iam/verification_models.go`.
- [x] 2.2 Define session model and introspection view model in `internal/iam/session_models.go`.
- [x] 2.3 Define request/response DTOs for OTP dispatch, OTP verify, and introspection in `internal/iam/authn_models.go`.

## 3. Data Access Layer (Repositories)
- [x] 3.1 Define repository interfaces in `internal/iam/repository.go`.
- [x] 3.2 Implement `identityRepository` in `internal/iam/identity_repository.go` (lookup by email, update email_verified_at).
- [x] 3.3 Implement `verificationRepository` in `internal/iam/verification_repository.go` (create OTP, query active by identity, increment attempts, delete/invalidate).
- [x] 3.4 Implement `sessionRepository` in `internal/iam/session_repository.go` (insert session, introspect token view).

## 4. Business Logic (UseCases & SessionManager)
- [x] 4.1 Define `authUseCases` and helper interfaces in `internal/iam/usecases.go`.
- [x] 4.2 Implement `SendOTP` workflow in `internal/iam/authn_usecases.go` (anti-enumeration check, 60s cooldown check, prior OTP invalidation, OTP generation, HMAC hashing, mailer dispatch).
- [x] 4.3 Implement `VerifyOTP` workflow in `internal/iam/authn_usecases.go` (attempt tracking, brute force invalidation, single-use deletion, session generation with `tok_` prefix, default org resolution, pending invitations query).
- [x] 4.4 Implement `SessionManager` exported contract in `internal/iam/session_manager.go` for token introspection.

## 5. Transport Layer & Middleware
- [x] 5.1 Implement `authnMiddleware` in `internal/iam/authn_middleware.go` to extract Bearer token, hash, validate via `SessionManager`, and inject `PrincipalClaims` into Echo context.
- [x] 5.2 Implement `authnController` in `internal/iam/authn_controller.go` handling `POST /v1/auth/otp/send`, `POST /v1/auth/otp/verify`, and `GET /v1/auth/introspect`.

## 6. Factory & Dependency Wiring
- [x] 6.1 Update `internal/iam/factory.go` to initialize repositories, use cases, session manager, middleware, and register routes on Echo.
- [x] 6.2 Wire `iam.NewModule` in `cmd/main.go` with database connection, mailer, and configuration.

## 7. Verification & Automated Tests
- [ ] 7.1 Write integration tests in `internal/iam/authn_test.go` using `pkg/toolbox/testbox` validating OTP dispatch, cooldown, invalid code attempt counting, brute-force invalidation, successful verification, session creation, and introspection.
- [ ] 7.2 Run `make test` and `make lint` to verify code quality and passing assertions.


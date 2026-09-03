# Design: In-House Passwordless User Login (Email OTP & Bearer Sessions)

## Context & Architecture

This design aligns with the multi-concept module pattern documented in `.agents/skills/module-pattern/SKILL.md` and the architecture in `docs/iam-multi-tenant.md`.

All logic resides in `internal/iam/` separated into clean domain layers:
- **Transport**: Echo HTTP handlers in `authn_controller.go`.
- **Business Workflows**: `usecases.go` and `authn_usecases.go`.
- **Cross-Module API**: `session_manager.go` implementing `SessionManager`.
- **Data Access**: `repository.go`, `identity_repository.go`, `verification_repository.go`, `session_repository.go` using `uptrace/bun`.
- **Models & DTOs**: `authn_models.go`, `identity_models.go`, `session_models.go`.
- **Security & Middleware**: `authn_middleware.go` for Bearer token validation and context hydration.

```
                  ┌────────────────────────────────────────┐
                  │              HTTP Client               │
                  └───────────────────┬────────────────────┘
                                      │
              ┌───────────────────────┼───────────────────────┐
              ▼                       ▼                       ▼
POST /v1/auth/otp/send    POST /v1/auth/otp/verify    GET /v1/auth/introspect
              │                       │                       │
              │                       │             [ Bearer Middleware ]
              │                       │                       │
              └───────────────────────┼───────────────────────┘
                                      ▼
                        ┌───────────────────────────┐
                        │     authn_controller      │
                        └─────────────┬─────────────┘
                                      │
                        ┌─────────────▼─────────────┐
                        │      authn_usecases       │
                        └─────────────┬─────────────┘
                                      │
             ┌────────────────────────┼────────────────────────┐
             ▼                        ▼                        ▼
┌─────────────────────────┐ ┌────────────────────┐ ┌─────────────────────────┐
│   identityRepository    │ │ verificationRepo   │ │    sessionRepository    │
└─────────────────────────┘ └────────────────────┘ └─────────────────────────┘
             │                        │                        │
             └────────────────────────┼────────────────────────┘
                                      ▼
                            PostgreSQL Database
```

## Detailed Component Specifications

### 1. Data Models (`internal/iam/`)

- **Identity**: Maps to table `identities` (`id`, `email`, `first_name`, `last_name`, `state`, `email_verified_at`, `created_at`, `updated_at`).
- **Verification**: Maps to table `verifications` (`id`, `identity_id`, `kind`, `attempts`, `expires_at`, `created_at`).
  - `id`: Holds the HMAC-SHA256 hash of `identity_id + ":" + code` using `Iam.HmacSecret`.
  - `kind`: `'email_otp'`.
  - `attempts`: Counter incremented on wrong submissions (max 3).
  - `expires_at`: 10 minutes from creation.
- **IdentitySession**: Maps to table `identity_sessions` (`id`, `identity_id`, `organization_id`, `ip_address`, `user_agent`, `expires_at`, `created_at`).
  - `id`: SHA-256 hash of the full opaque token string.
  - `expires_at`: 24 hours from creation.
- **SessionIntrospectionView**: Maps to view `session_introspection_view` for single-query retrieval of identity, session, organization, and membership details.
  *(Note: fix the trailing comma syntax error in migration `20260902035257_iam_sessions.sql` line 31 before execution)*.

### 2. OTP Generation & Verification Logic

- **Code Generation**: 6-digit numeric string formatted with leading zeros (e.g. `fmt.Sprintf("%06d", randInt(0, 1000000))`) using `crypto/rand`.
- **HMAC Hashing**: Computed via `hasher.NewHmacSHA256(config.HmacSecret)` on `fmt.Sprintf("%s:%s", identityID, otpCode)`.
- **Dispatch Flow (`sendOTP`)**:
  1. Lookup identity by email (case-insensitive).
  2. If identity not found or suspended, return success without error (`anti-enumeration`).
  3. Check latest verification for cooldown: if created within past 60s, return `terrors.TooManyRequests`.
  4. Delete any existing pending verifications for this identity & `kind = 'email_otp'`.
  5. Insert new verification record with HMAC hash as `id`.
  6. Call `mailer.SendOTPCode(ctx, email, code)`.
- **Verification Flow (`verifyOTP`)**:
  1. Lookup identity by email. If not found, return `terrors.Unauthorized` ("invalid credentials").
  2. Query active verification: `WHERE identity_id = ? AND kind = 'email_otp' AND expires_at > now() ORDER BY created_at DESC LIMIT 1`.
  3. If none found, return `terrors.Unauthorized` ("expired or invalid code").
  4. If `attempts >= 3`, delete record and return `terrors.Unauthorized` ("maximum attempts exceeded").
  5. Compute expected hash `hmac(identity.id + ":" + code)`.
  6. If hash != `v.id`:
     - Increment `attempts = attempts + 1` in database.
     - If `attempts >= 3`, delete verification record.
     - Return `terrors.Unauthorized` ("invalid code").
  7. If hash == `v.id`:
     - Delete verification record (single-use consumption).
     - Update `identities.email_verified_at = now()`.
     - Generate token using `toolbox.GenerateOpaqueToken(hasher, "tok")`.
     - Resolve primary organization membership (first membership by `created_at ASC`).
     - Insert session into `identity_sessions`.
     - Query any pending invitations for the user's email.
     - Return raw token, identity, active organization, and pending invitations.

### 3. Session Introspection & Middleware

- **`SessionManager` Interface** (cross-module export):
  ```go
  type SessionManager interface {
      Introspect(ctx context.Context, token string) (*PrincipalClaims, error)
  }
  ```
- **Echo Middleware (`RequireAuth`)**:
  1. Read `Authorization` header, parse `Bearer <token>`.
  2. Compute SHA-256 hash of token.
  3. Call `SessionManager.Introspect(...)`.
  4. Set `PrincipalClaims` in Echo context (`c.Set("principal", claims)`).
  5. Propagate context down the handler chain.
- **Introspection Handler**:
  - Pulls `PrincipalClaims` from Echo context and serializes directly to JSON.

## Database & Migration Fix

Migration `20260902035257_iam_sessions.sql` contains a trailing comma before `FROM` in `session_introspection_view`:
```sql
COALESCE(om.role, '') AS organization_role, -- trailing comma must be removed
FROM identity_sessions s
```
This syntax defect will be corrected in the migration file so that `dbmate up` and test container initialization execute cleanly.

## Security Considerations

1. **Email Enumeration Defense**: `otp/send` returns identical HTTP 200 OK payloads regardless of user existence.
2. **Brute Force Rate Limiting**: Max 3 attempts per 6-digit OTP code before immediate deletion; 10-minute maximum lifespan; 60-second cooldown per email.
3. **Entropy & Storage**: 256-bit cryptographically secure random session tokens; only SHA-256 hashes are stored in the database.
4. **Timing Attack Protection**: HMAC comparison uses constant-time byte comparison (`hmac.Equal`).


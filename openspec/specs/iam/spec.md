# IAM Specification

## Purpose

In-house Identity and Access Management (IAM) providing passwordless email OTP verification, cryptographically secure opaque Bearer session tokens, and session introspection for multi-tenant B2B architectures.

## Requirements

### Requirement: Email OTP Dispatch (`POST /v1/auth/otp/send`)
The system MUST provide a public endpoint to generate and dispatch a 6-digit one-time passcode to verified or invited email addresses.

#### Scenario: OTP successfully sent to existing invited/active identity
- **GIVEN** an active identity with email `"user@example.com"` exists in the database
- **AND** no OTP has been dispatched to `"user@example.com"` within the last 60 seconds
- **WHEN** a client sends a `POST /v1/auth/otp/send` request with body `{"email": "user@example.com"}`
- **THEN** the system MUST respond with HTTP 200 OK and `{"sent": true}`
- **AND** an email containing a 6-digit numeric code MUST be queued/sent via the Mailer service
- **AND** an active verification record with `kind = 'email_otp'` and 10-minute expiry MUST be persisted.

#### Scenario: Silent success for uninvited/unknown email (Anti-enumeration)
- **GIVEN** no identity or invitation exists for `"stranger@example.com"`
- **WHEN** a client sends a `POST /v1/auth/otp/send` request with body `{"email": "stranger@example.com"}`
- **THEN** the system MUST respond with HTTP 200 OK and `{"sent": true}`
- **AND** the system MUST NOT dispatch any email or create any verification record.

#### Scenario: Dispatch cooldown violation
- **GIVEN** an OTP was already dispatched to `"user@example.com"` 30 seconds ago
- **WHEN** a client sends a `POST /v1/auth/otp/send` request with body `{"email": "user@example.com"}`
- **THEN** the system MUST respond with HTTP 429 Too Many Requests (or a rate-limit error) indicating cooldown in progress.

#### Scenario: Superseding previous pending OTP
- **GIVEN** an existing pending OTP exists for `"user@example.com"` created >60 seconds ago
- **WHEN** a client sends a `POST /v1/auth/otp/send` request for `"user@example.com"`
- **THEN** any prior unexpired OTP verification records for that identity MUST be invalidated/deleted.

### Requirement: Email OTP Verification & Session Issuance (`POST /v1/auth/otp/verify`)
The system MUST provide a public endpoint to verify a submitted OTP, record email verification, create a session, and issue an opaque Bearer token.

#### Scenario: Successful OTP verification with active organization
- **GIVEN** an active identity `"usr_123"` with email `"user@example.com"` having an unexpired OTP code `"654321"`
- **AND** the identity belongs to organization `"org_abc"` with role `"admin"`
- **WHEN** the client sends a `POST /v1/auth/otp/verify` request with `{"email": "user@example.com", "code": "654321"}`
- **THEN** the system MUST respond with HTTP 200 OK containing:
  - `token`: an opaque string starting with `tok_`
  - `identity`: identity profile data
  - `active_organization`: details of `"org_abc"` and role `"admin"`
  - `pending_invitations`: empty array
- **AND** `identities.email_verified_at` MUST be updated with current timestamp
- **AND** the verification record MUST be deleted (consumed)
- **AND** a session MUST be stored in `identity_sessions` with 24-hour expiration.

#### Scenario: Successful OTP verification for invited user without memberships
- **GIVEN** an invited identity `"usr_456"` with no organization memberships and 1 pending invitation
- **WHEN** the client submits the valid OTP code
- **THEN** the system MUST respond with HTTP 200 OK containing:
  - `token`: valid opaque Bearer token
  - `identity`: identity details
  - `active_organization`: `null`
  - `pending_invitations`: array containing the pending invitation details.

#### Scenario: Invalid OTP code submission (Attempt tracking)
- **GIVEN** an unexpired OTP exists with `attempts = 0`
- **WHEN** the client submits an incorrect code
- **THEN** the system MUST increment `attempts` to 1
- **AND** respond with HTTP 400 Bad Request or 401 Unauthorized indicating an invalid code.

#### Scenario: Exceeded maximum attempts (Brute-force protection)
- **GIVEN** an OTP record currently has `attempts = 2`
- **WHEN** the client submits an incorrect code
- **THEN** the verification record MUST be invalidated/deleted
- **AND** subsequent verification attempts for that code MUST be rejected.

### Requirement: Session Authentication Middleware & Introspection (`GET /v1/auth/introspect`)
The system MUST provide authentication middleware validating Bearer tokens and an introspection endpoint returning caller claims.

#### Scenario: Introspecting valid session token
- **GIVEN** a valid, unexpired session token `tok_xyz` associated with identity `"usr_123"` and active org `"org_abc"`
- **WHEN** the client sends a `GET /v1/auth/introspect` request with header `Authorization: Bearer tok_xyz`
- **THEN** the system MUST return HTTP 200 OK with `PrincipalClaims`:
  - `identity_id`: `"usr_123"`
  - `email`: `"user@example.com"`
  - `first_name`: string
  - `last_name`: string
  - `active_organization_id`: `"org_abc"`
  - `organization_name`: string
  - `organization_slug`: string
  - `organization_role`: string

#### Scenario: Missing or invalid Bearer token
- **WHEN** the client sends a `GET /v1/auth/introspect` request with an invalid or expired Bearer token
- **THEN** the middleware MUST intercept the request and respond with HTTP 401 Unauthorized.


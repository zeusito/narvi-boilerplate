# Proposal: In-House Passwordless User Login (Email OTP & Bearer Sessions)

## Why

The platform requires a secure, first-party authentication mechanism without external SaaS identity dependencies (such as Auth0, Clerk, or Kinde). Following the approved architecture specification in `docs/iam-multi-tenant.md`, this proposal establishes the foundational user authentication flow:
1. Dispatching a single-use 6-digit email One-Time Passcode (OTP).
2. Verifying the code and issuing a cryptographically secure 256-bit opaque Bearer session token.
3. Hydrating the user's identity and active tenant organization claims via session introspection.

## What

Implement the core authentication endpoints and middleware within the `internal/iam` module adhering to the multi-concept module pattern:

1. **`POST /v1/auth/otp/send`**:
   - Accepts `{ "email": string }`.
   - Anti-enumeration protection: Returns HTTP 200 `{"sent": true}` regardless of whether the email exists.
   - For valid, invited/registered identities, enforces a 60-second dispatch cooldown, invalidates any existing pending OTPs, generates a cryptographically random 6-digit code, stores its HMAC-SHA256 hash in `verifications.id` (10-minute TTL, 3 attempts), and sends it via `mailer.Mailer`.

2. **`POST /v1/auth/otp/verify`**:
   - Accepts `{ "email": string, "code": string }`.
   - Validates the code against the active verification record; increments failed attempts up to 3 before invalidation.
   - On success: marks `identities.email_verified_at = now()`, creates a server-side session in `identity_sessions` (24-hour TTL), resolves the default active organization, and returns the raw opaque Bearer token (`tok_<crockford-base32>`), identity profile, active organization context, and any pending invitations.

3. **`GET /v1/auth/introspect`**:
   - Authenticated endpoint protected by the Bearer token middleware.
   - Resolves and returns current principal claims (`identity_id`, `email`, `first_name`, `last_name`, `active_organization_id`, `organization_name`, `organization_slug`, `organization_role`).

4. **Authentication Middleware**:
   - Echo middleware extracting `Authorization: Bearer <token>`, hashing the token, querying the active session, and injecting `PrincipalClaims` into the request context.

## Key Design Decisions & Invariants

- **Invite-Only Enforcement**: Identities are pre-provisioned when invited. Unregistered emails receive a silent success response without dispatching emails.
- **Verification Storage**: Hashed code stored in `verifications.id` with `kind = 'email_otp'`. Single active OTP per user.
- **Brute-Force & Replay Defense**: Maximum 3 verification attempts; code expires in 10 minutes; deleted immediately upon successful verification.
- **Session Tokens**: 256-bit entropy generated with Crockford Base32 encoding and prefix `tok_`. Database stores only the SHA-256 / HMAC-SHA256 hash in `identity_sessions.id`.
- **Default Organization**: Active organization is defaulted to the user's primary/earliest organization membership. If the user has no memberships yet, `organization_id` is null and pending invitations are listed.
- **Scope Boundary**: Authentication-only. Role-based authorization via embedded OPA (`github.com/open-policy-agent/opa/rego`) is intentionally scoped for a dedicated follow-up change.

## Non-Goals

- Public self-registration / sign-up without invitation.
- Invitation dispatch and management (`/v1/invitations`).
- Organization context switching (`POST /v1/auth/session/organization`).
- Session revocation/logout (`DELETE /v1/auth/signout`).
- OPA Rego policy evaluation and permission checks (`RequirePermission`).
- Workspaces (to be introduced in a future multi-tenancy iteration).


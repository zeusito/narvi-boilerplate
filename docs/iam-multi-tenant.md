# Multi-Tenant B2B IAM & Authorization Design Document

**Status:** Approved Architecture Specification  
**Scope:** In-House Passwordless Authentication (Email OTP), Opaque Bearer Session Tokens, B2B Multi-Tenancy, Role-Based Access Control (RBAC), Embedded OPA (Open Policy Agent) Enforcement, and Strict Invite-Only Onboarding.

---

## 1. Executive Summary & Vision

This document outlines the reference Identity and Access Management (IAM) and Multi-Tenant architecture for the application boilerplate. The boilerplate is designed to serve as a high-performance, secure foundation for B2B SaaS applications.

Rather than relying on third-party federated identity providers (e.g., Auth0, Kinde, Clerk), this architecture implements a **first-party, in-house Identity and Access Management system** optimized for low operational footprint, data ownership, and sub-millisecond policy evaluation.

### Key Design Pillars:

1. **First-Party Passwordless Auth (Email OTP):** Authenticates users via secure, single-use 6-digit one-time passcodes (OTPs) with rate limiting, cryptographic hashing, and automated expiration.
2. **Opaque Bearer Session Tokens:** Issues cryptographically random 256-bit opaque tokens sent via standard `Authorization: Bearer <opaque_token>` HTTP headers. Session state and active organization context are persisted in PostgreSQL.
3. **Flat B2B Multi-Tenancy:** Decouples global human identities (`identities`) from tenant organizations (`organizations`). An identity can belong to multiple organizations via `organization_memberships`.
4. **Session-State Tenant Context:** Active organization context is bound to the server-side session record and switchable on-demand via `POST /v1/auth/session/organization`.
5. **Fixed System Roles as Collections of Permissions:** Pre-seeded hierarchical roles (`Owner`, `Admin`, `Member`, `Viewer`) composed of explicit `<resource>:<action>` permissions (e.g., `org:delete`, `members:invite`, `resources:create`).
6. **Embedded OPA Authorization Enforcer:** Utilizes Open Policy Agent's Go SDK (`github.com/open-policy-agent/opa/rego`) in-process for zero-network-overhead, sub-millisecond authorization decisions.
7. **Strict Invite-Only Onboarding:** Arbitrary public registrations are blocked; new identities are registered exclusively through verified email invitations. Bootstrap organizations are initialized via seed migrations or administrative CLI tools.

---

## 2. System Architecture & Entity Model

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                                CLIENT APPLICATION                                │
│                Sends: `Authorization: Bearer <opaque_session_token>`              │
└────────────────────────────────────────┬─────────────────────────────────────────┘
                                         │ HTTP REST API Requests
                                         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                           GO APPLICATION CORE (BACKEND)                          │
│                                                                                  │
│  1. Authentication Middleware                                                    │
│     ├── Validates Opaque Bearer Token against PostgreSQL `sessions`              │
│     └── Hydrates Principal Context: { IdentityID, ActiveOrgID, Role }            │
│                                                                                  │
│  2. Embedded OPA Authorization Middleware (github.com/open-policy-agent/opa/rego) │
│     ├── Loads Rego Policy Bundle in-process                                      │
│     ├── Evaluates: Input { principal, action, resource, target_org }             │
│     └── Enforces: Allow / Deny + Returns HTTP 403 Forbidden                      │
│                                                                                  │
│  3. Domain & Tenant Services                                                     │
│     ├── Auth / OTP Service (Generate, Rate-Limit, Verify)                        │
│     ├── Organization & Membership Service                                        │
│     └── Invitation State Machine Service (72h TTL)                               │
└────────────────────────────────────────┬─────────────────────────────────────────┘
                                         │
                                         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                             POSTGRESQL DATA STORE                                │
│                                                                                  │
│  ├── identities                     (Global User Records)                        │
│  ├── verifications                  (Hashed OTPs, TTL, Attempts)                 │
│  ├── sessions                       (Opaque Tokens, Active Org ID, Expiry)       │
│  ├── organizations                  (Tenant Entities)                            │
│  ├── organization_memberships       (Identity ↔ Org Links with System Roles)     │
│  └── invitations                    (Strict Invite-Only Onboarding Records)      │
└──────────────────────────────────────────────────────────────────────────────────┘
```



---

## 3. Role-Based Access Control (RBAC) & Permissions

The system implements a fixed, hierarchical role structure where each role represents a predefined set of granular permissions. Permissions use standard `<resource>:<action>` notation.

```
                  ┌────────────────────────┐
                  │         Owner          │
                  │   (All Permissions)    │
                  └───────────┬────────────┘
                              │ inherits
                  ┌───────────▼────────────┐
                  │         Admin          │
                  │ (Member & Org Control) │
                  └───────────┬────────────┘
                              │ inherits
                  ┌───────────▼────────────┐
                  │        Member          │
                  │ (Standard CRUD Ops)    │
                  └───────────┬────────────┘
                              │ inherits
                  ┌───────────▼────────────┐
                  │        Viewer          │
                  │   (Read-Only Access)   │
                  └────────────────────────┘
```

### 3.1 Permission Matrix

| Permission String | Description | `Viewer` | `Member` | `Admin` | `Owner` |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `org:read` | View organization profile and public metadata | ✅ | ✅ | ✅ | ✅ |
| `org:update` | Edit organization name, slug, and general settings | ❌ | ❌ | ✅ | ✅ |
| `org:delete` | Permanently delete the organization | ❌ | ❌ | ❌ | ✅ |
| `members:read` | List organization roster and member details | ✅ | ✅ | ✅ | ✅ |
| `members:invite` | Dispatch new email invitations to prospective members | ❌ | ❌ | ✅ | ✅ |
| `members:update_role` | Change another member's assigned role | ❌ | ❌ | ✅ (Up to Admin) | ✅ (Any) |
| `members:remove` | Evict members from the organization | ❌ | ❌ | ✅ (Members/Viewers) | ✅ (Any) |
| `invitations:read` | List pending, accepted, or revoked invitations | ❌ | ❌ | ✅ | ✅ |
| `invitations:revoke` | Revoke a pending organization invitation | ❌ | ❌ | ✅ | ✅ |
| `resources:read` | View application domain resources | ✅ | ✅ | ✅ | ✅ |
| `resources:create` | Create new application domain resources | ❌ | ✅ | ✅ | ✅ |
| `resources:update` | Modify existing application domain resources | ❌ | ✅ | ✅ | ✅ |
| `resources:delete` | Delete application domain resources | ❌ | ❌ | ✅ | ✅ |

---

## 4. Open Policy Agent (OPA) Authorization Engine

Authorization is enforced in-process within the Go backend using Open Policy Agent's Go SDK (`github.com/open-policy-agent/opa/rego`). This avoids external network hops while maintaining declarative, testable policies.

### 4.1 Policy Evaluation Input Schema

For every incoming HTTP request requiring authorization, the Go HTTP middleware constructs an OPA input document:

```json
{
  "principal": {
    "identity_id": "usr_01H9Z2W8Q6A0K1M3N5P7R9T1V3",
    "email": "alex@example.com",
    "active_organization_id": "org_01H9Z2W8Q6A0K1M3N5P7R9T1V1",
    "role": "admin"
  },
  "action": "members:invite",
  "resource": {
    "type": "organization",
    "organization_id": "org_01H9Z2W8Q6A0K1M3N5P7R9T1V1"
  }
}
```

### 4.2 Reference Rego Policy (`policy/authz.rego`)

```rego
package authz

import future.keywords.in

default allow = false

# Role to permission mapping
role_permissions := {
    "owner": [
        "org:read", "org:update", "org:delete",
        "members:read", "members:invite", "members:update_role", "members:remove",
        "invitations:read", "invitations:revoke",
        "resources:read", "resources:create", "resources:update", "resources:delete"
    ],
    "admin": [
        "org:read", "org:update",
        "members:read", "members:invite", "members:update_role", "members:remove",
        "invitations:read", "invitations:revoke",
        "resources:read", "resources:create", "resources:update", "resources:delete"
    ],
    "member": [
        "org:read",
        "members:read",
        "resources:read", "resources:create", "resources:update"
    ],
    "viewer": [
        "org:read",
        "members:read",
        "resources:read"
    ]
}

# Rule 1: Tenant Boundary Enforcement
# Principal's active organization must match the target resource's organization_id
tenant_match {
    input.principal.active_organization_id == input.resource.organization_id
}

# Rule 2: Permission Check
has_permission {
    role := input.principal.role
    permissions := role_permissions[role]
    input.action in permissions
}

# Final Decision
allow {
    tenant_match
    has_permission
}
```

### 4.3 In-Process Go Middleware Integration

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/open-policy-agent/opa/rego"
)

type AuthzEnforcer struct {
	query rego.PreparedEvalQuery
}

func NewAuthzEnforcer(ctx context.Context, regoPolicy string) (*AuthzEnforcer, error) {
	query, err := rego.New(
		rego.Query("data.authz.allow"),
		rego.Module("authz.rego", regoPolicy),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, err
	}
	return &AuthzEnforcer{query: query}, nil
}

func (e *AuthzEnforcer) RequirePermission(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal := GetPrincipal(r.Context())
			
			input := map[string]interface{}{
				"principal": map[string]interface{}{
					"identity_id":            principal.IdentityID,
					"email":                  principal.Email,
					"active_organization_id": principal.ActiveOrgID,
					"role":                   principal.Role,
				},
				"action": action,
				"resource": map[string]interface{}{
					"type":            "organization",
					"organization_id": principal.ActiveOrgID,
				},
			}

			results, err := e.query.Eval(r.Context(), rego.EvalInput(input))
			if err != nil || len(results) == 0 || !results[0].Bindings["x"].(bool) {
				http.Error(w, `{"error":"forbidden","message":"insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

---

## 5. Token & Session Architecture

### 5.1 Opaque Token Format & Transmission

- **Transmission:** Sent exclusively via the standard HTTP Authorization header:
  ```http
  Authorization: Bearer <opaque_token>
  ```
- **Token Generation:** 32 bytes of cryptographically secure randomness (`crypto/rand`), base64url-encoded (prefix: `tok_`).
- **Database Storage:** Only the SHA-256 hash (`token_hash`) of the token is stored in the database. The raw token string is provided to the client once upon successful OTP verification.
- **Session Sliding Window:** Sessions default to an active lifespan of 30 days. Each authenticated request updates `last_active_at`. Sessions inactive beyond the expiration threshold are rejected.

### 5.2 Active Organization Context Switching

When an identity belongs to multiple organizations:
1. The session record retains the currently selected `active_organization_id`.
2. The user switches organizations by calling:
   ```http
   POST /v1/auth/session/organization
   Content-Type: application/json

   {
     "organization_id": "org_target_0123"
   }
   ```
3. The backend verifies that the caller has an active `organization_memberships` record for `org_target_0123`, updates `sessions.active_organization_id`, and returns updated principal claims.
4. Subsequent API requests evaluate permissions against the newly active organization.

---

## 6. User Lifecycle & Authentication Flows

### 6.1 Passwordless Sign-In via Email OTP

```
CLIENT                              BACKEND / API                       POSTGRESQL / EMAIL
  │                                       │                                     │
  │ 1. POST /v1/auth/otp/send             │                                     │
  │    { "email": "user@corp.com" }       │                                     │
  │ ────────────────────────────────────> │                                     │
  │                                       │ 2. Check Pending Invites / Members  │
  │                                       │ ──────────────────────────────────> │
  │                                       │    (Deny if no invite or member)    │
  │                                       │ 3. Generate 6-Digit Code & Hash     │
  │                                       │ 4. Insert email_verification_codes  │
  │                                       │ ──────────────────────────────────> │
  │                                       │ 5. Dispatch Email with OTP Code     │
  │                                       │ ──────────────────────────────────> │
  │ <──────────────────────────────────── │                                     │
  │    200 OK: { "sent": true }           │                                     │
  │                                       │                                     │
  │ 6. POST /v1/auth/otp/verify           │                                     │
  │    { "email": "...", "code": "..." }  │                                     │
  │ ────────────────────────────────────> │                                     │
  │                                       │ 7. Validate Hash & Expiry (10m TTL) │
  │                                       │ ──────────────────────────────────> │
  │                                       │ 8. Upsert Identity & Create Session │
  │                                       │ 9. Resolve Default Active Org       │
  │                                       │ ──────────────────────────────────> │
  │ <──────────────────────────────────── │                                     │
  │    200 OK: {                          │                                     │
  │      "token": "tok_...",              │                                     │
  │      "identity": { ... },             │                                     │
  │      "active_organization": { ... },  │                                     │
  │      "pending_invitations": [ ... ]   │                                     │
  │    }                                  │                                     │
```

### 6.2 Strict Invite-Only Onboarding Lifecycle

To enforce strict tenancy and prevent unauthorized registrations:
1. **Public sign-up is closed.** The `/v1/auth/otp/send` endpoint checks if the requested email matches an active row in `organization_memberships` or a pending row in `invitations`. If no association exists, the request returns `403 Forbidden` (or a generic success response to prevent user enumeration while declining to send the OTP).
2. **Invitation Dispatch:** An authenticated `Admin` or `Owner` creates an invitation via `POST /v1/invitations`.
3. **Invitation Acceptance:** Upon completing OTP verification, a newly registered identity fetches their pending invites (`GET /v1/me/invitations`) and calls `POST /v1/me/invitations/{id}/accept`.
4. **Membership Grant:** The backend inserts an `organization_memberships` record with the invited role, transitions the invitation status to `accepted`, and sets the newly joined organization as the active organization in the session.

---

## 7. REST API Specifications

### 7.1 Authentication & Session Management

| Method | Endpoint | Auth Required | Description |
| :--- | :--- | :---: | :--- |
| `POST` | `/v1/auth/otp/send` | No | Generates and emails a 6-digit verification OTP (subject to strict invite/membership check). |
| `POST` | `/v1/auth/otp/verify` | No | Verifies the OTP, provisions identity, creates a session, and returns an opaque Bearer token. |
| `GET` | `/v1/auth/introspect` | Yes | Returns authenticated identity claims, active organization, and resolved role/permissions. |
| `POST` | `/v1/auth/session/organization` | Yes | Updates the active organization context for the current session. |
| `DELETE`| `/v1/auth/signout` | Yes | Revokes the active session token, invalidating it from PostgreSQL. |

### 7.2 Organization & Membership Management

| Method | Endpoint | Required Permission | Description |
| :--- | :--- | :---: | :--- |
| `GET` | `/v1/organizations` | Authenticated | Lists all organizations where the caller holds active membership. |
| `GET` | `/v1/organizations/{id}` | `org:read` | Returns organization details for the active organization. |
| `PATCH`| `/v1/organizations/{id}` | `org:update` | Updates organization metadata (e.g., name). |
| `DELETE`| `/v1/organizations/{id}`| `org:delete` | Permanently removes the organization and cascades associated data. |
| `GET` | `/v1/members` | `members:read` | Lists all members and assigned roles in the active organization. |
| `PATCH`| `/v1/members/{identityId}`| `members:update_role` | Updates the role of a member within the active organization. |
| `DELETE`| `/v1/members/{identityId}`| `members:remove` | Removes a member from the active organization. |

### 7.3 Invitations Management

| Method | Endpoint | Required Permission / Scope | Description |
| :--- | :--- | :---: | :--- |
| `POST` | `/v1/invitations` | `members:invite` | Creates and emails a new membership invitation (72h TTL). |
| `GET` | `/v1/invitations` | `invitations:read` | Lists all pending/historical invitations for the active organization. |
| `DELETE`| `/v1/invitations/{id}` | `invitations:revoke` | Revokes an active pending invitation. |
| `GET` | `/v1/me/invitations` | Authenticated (Self) | Returns all pending invitations matching the caller's verified email. |
| `POST` | `/v1/me/invitations/{id}/accept` | Authenticated (Self) | Accepts an invitation and inserts the membership into the target organization. |
| `POST` | `/v1/me/invitations/{id}/decline`| Authenticated (Self) | Declines a pending invitation. |

---

## 8. Security Guarantees & Operational Hardening

1. **Token Invalidation on Revocation:** Deleting a session row immediately rejects subsequent Bearer requests since validation queries the database index on each authenticated request.
2. **OTP Brute-Force Defense:** Verification attempts are capped at 3 tries per code. After 3 failed attempts, the code is invalidated (`attempts_remaining = 0`). Codes automatically expire after 10 minutes.
3. **Tenant Data Isolation:** All domain database queries are parameterized with `organization_id = $1` resolved from the validated `sessions.active_organization_id`.
4. **Auditability:** All invitation state changes (`pending` $\rightarrow$ `accepted` | `declined` | `revoked`) record the actor's identity ID and timestamp for enterprise compliance.

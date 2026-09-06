## ADDED Requirements

### Requirement: Role-to-Permission Expansion
The system MUST expand a principal's active organization role into an explicit set of granular permissions before evaluating authorization.

#### Scenario: Expanding owner role permissions
- **GIVEN** a principal with active organization role `"owner"`
- **WHEN** the system expands the role into permissions
- **THEN** the permissions set MUST contain all organization, member, invitation, and resource management permissions (`org:read`, `org:update`, `org:delete`, `members:read`, `members:invite`, `members:update_role`, `members:remove`, `invitations:read`, `invitations:revoke`, `resources:read`, `resources:create`, `resources:update`, `resources:delete`).

#### Scenario: Expanding admin role permissions
- **GIVEN** a principal with active organization role `"admin"`
- **WHEN** the system expands the role into permissions
- **THEN** the permissions set MUST include member, invitation, and resource operations and org update/read, but MUST NOT include `org:delete`.

#### Scenario: Expanding member role permissions
- **GIVEN** a principal with active organization role `"member"`
- **WHEN** the system expands the role into permissions
- **THEN** the permissions set MUST include `org:read`, `members:read`, `resources:read`, `resources:create`, and `resources:update`, and MUST NOT include administrative permissions (`org:update`, `org:delete`, `members:invite`, `members:remove`, `members:update_role`, `invitations:revoke`, `resources:delete`).

#### Scenario: Expanding viewer role permissions
- **GIVEN** a principal with active organization role `"viewer"`
- **WHEN** the system expands the role into permissions
- **THEN** the permissions set MUST strictly contain read-only permissions (`org:read`, `members:read`, `resources:read`).

#### Scenario: Principal with unknown or empty role
- **GIVEN** a principal with an empty or unrecognized organization role
- **WHEN** the system expands the role into permissions
- **THEN** the permissions set MUST be empty.

### Requirement: OPA Policy Authorization Guard
The system MUST enforce authorization via an in-process Open Policy Agent (OPA) evaluation engine using embedded Rego policy before executing protected route actions.

#### Scenario: Authorized access with matching tenant and permission
- **GIVEN** an authenticated principal in active organization `"org_alpha"` holding permission `"org:read"`
- **WHEN** the client invokes an endpoint protected by permission `"org:read"` for resource organization `"org_alpha"`
- **THEN** the OPA guard MUST allow the request to proceed to the handler.

#### Scenario: Forbidden access due to missing permission
- **GIVEN** an authenticated principal in active organization `"org_alpha"` holding role `"viewer"` (lacking `"org:update"`)
- **WHEN** the client invokes an endpoint protected by permission `"org:update"`
- **THEN** the OPA guard MUST intercept the request and respond with HTTP 403 Forbidden
- **AND** the response body MUST follow standard error format `{"code": "ActionForbidden", "message": "insufficient permissions"}`.

#### Scenario: Forbidden access due to cross-tenant boundary violation
- **GIVEN** an authenticated principal with active organization `"org_alpha"` and role `"owner"`
- **WHEN** the client invokes an endpoint protected by permission `"org:read"` targeting resource organization `"org_beta"`
- **THEN** the OPA guard MUST intercept the request and respond with HTTP 403 Forbidden
- **AND** the response body MUST follow standard error format `{"code": "ActionForbidden", "message": "insufficient permissions"}`.

#### Scenario: Forbidden access without active organization
- **GIVEN** an authenticated principal who has no active organization selected (`active_organization_id` is empty)
- **WHEN** the client invokes an endpoint protected by a tenant-scoped permission
- **THEN** the OPA guard MUST intercept the request and respond with HTTP 403 Forbidden.

## MODIFIED Requirements

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
  - `permissions`: array of permission strings expanded from the active organization role

#### Scenario: Missing or invalid Bearer token
- **WHEN** the client sends a `GET /v1/auth/introspect` request with an invalid or expired Bearer token
- **THEN** the middleware MUST intercept the request and respond with HTTP 401 Unauthorized.

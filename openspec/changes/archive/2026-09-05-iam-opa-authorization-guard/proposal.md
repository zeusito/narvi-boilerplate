## Why

The backend currently validates identity sessions via `RequireAuth`, but lacks a declarative, sub-millisecond authorization layer to enforce tenant boundaries and role permissions. As defined in `docs/iam-multi-tenant.md`, endpoints must be guarded by Open Policy Agent (OPA) evaluations so that users can perform only actions permitted by their role in the active tenant organization.

Establishing an embedded OPA authorization engine now provides the foundational security guardrails needed before introducing organization, member, and domain resource mutation endpoints.

## What Changes

- **Embedded OPA Engine (`pkg/opa`)**: A generic, in-process Rego evaluation service leveraging `github.com/open-policy-agent/opa/rego` with embedded default policies (`//go:embed`) and pre-compiled query evaluation (`rego.PrepareForEval`).
- **Role-to-Permission Expansion (`internal/iam`)**: Go-side definition of fixed hierarchical roles (`Owner`, `Admin`, `Member`, `Viewer`) and standard `<resource>:<action>` permissions, expanding active roles into explicit permission slices passed to OPA.
- **Echo Authorization Guard Middleware (`internal/iam`)**: `RequirePermission(action, ...opts)` middleware that extracts active session context, maps permissions, resolves resource organization context, evaluates OPA policy, and returns standard HTTP 403 `terrors.Forbidden("insufficient permissions")` on denial.
- **Rego Authorization Policy (`policy/authz.rego`)**: Rego policy verifying tenant boundary isolation (`principal.active_organization_id == resource.organization_id`) and permission containment (`action in principal.permissions`).
- **Principal Claims Enrichment**: Enriches `PrincipalClaims` with expanded permissions or helper accessors for downstream controllers.
- **Verification & Demonstration Routes**: Protected test/sample endpoints verifying allow/deny decisions across tenant and role scenarios.

## Capabilities

### Modified Capabilities
- `iam`: Adding requirements for embedded OPA authorization enforcer, role-to-permission mapping, tenant boundary enforcement, and `RequirePermission` Echo route guard.

## Architectural Boundaries & Permission Implications

- **Separation of Concerns**: Generic policy evaluation engine lives in `pkg/opa`; IAM-specific roles, permissions, context hydration, and route guards live in `internal/iam`.
- **Tenant Isolation**: OPA input requires matching `principal.active_organization_id` and `resource.organization_id`. Cross-tenant actions are rejected with HTTP 403.
- **Domain vs Policy Boundary**: Rego policy strictly validates tenant isolation and permission strings. Complex business constraints (e.g., preventing the last Owner from demoting themselves or leaving) remain in domain service handlers.

## Impact

- **backend-api**:
  - Adds dependency `github.com/open-policy-agent/opa` in `go.mod`.
  - Introduces `pkg/opa` package with embedded Rego.
  - Updates `internal/iam` with permissions constants, role expansion, and `RequirePermission` middleware.
  - No database migration required; roles are already stored on `organization_memberships.role` and `PrincipalClaims.OrganizationRole`.
- **frontend-hub**:
  - No breaking changes. Future client features can utilize introspected permissions or handle standardized HTTP 403 error responses (`ActionForbidden`).

## Non-Goals

- Full CRUD endpoints for organizations and members (`/v1/organizations`, `/v1/members`), which will be specified in dedicated subsequent changes.
- Invitation state machine and acceptance endpoints (`/v1/invitations`).
- Custom or dynamic tenant roles stored in database (future enterprise enhancement; currently fixed system roles).
- Remote OPA bundle server integration (in-process embedded Rego satisfies sub-millisecond, low-footprint goals).

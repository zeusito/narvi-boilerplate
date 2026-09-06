## Context

Currently, the `internal/iam` module authenticates requests using opaque Bearer tokens via `RequireAuth`, populating `PrincipalClaims` in the Echo context with identity details and `OrganizationRole`. However, there is no enforcement layer to restrict actions according to tenant boundaries or role permissions.

As specified in `docs/iam-multi-tenant.md`, the platform utilizes Open Policy Agent (OPA) embedded in-process to provide declarative, sub-millisecond authorization.

## Goals / Non-Goals

**Goals:**
- Provide an in-process, zero-network-overhead policy evaluation engine in `pkg/opa` using `github.com/open-policy-agent/opa/rego`.
- Define strongly-typed roles and granular permissions in `internal/iam` matching the approved permission matrix.
- Expand organization roles into explicit permissions in Go and feed `principal.permissions` into the OPA input.
- Provide an Echo middleware `RequirePermission(action, ...opts)` that checks active organization context, evaluates OPA policy, and returns HTTP 403 `terrors.Forbidden("insufficient permissions")` on denial.
- Support default active organization tenant matching with extensible resource organization extraction for path parameters.
- Provide comprehensive unit tests covering all matrix roles, tenant mismatches, and edge cases.

**Non-Goals:**
- Implementing complete Organization or Member CRUD endpoints (deferred to subsequent changes).
- Dynamic database-stored custom roles (fixed hierarchical roles: Owner, Admin, Member, Viewer).
- Remote OPA bundle server synchronization or daemon sidecars.

## Decisions

### 1. Split Architecture: Generic Enforcer (`pkg/opa`) and IAM Domain Logic (`internal/iam`)
- **Choice:** Create a reusable, domain-agnostic OPA evaluator in `pkg/opa` that manages policy compilation (`rego.PrepareForEval`), input binding, and result extraction. Place role/permission constants, role expansion, and Echo middleware in `internal/iam`.
- **Rationale:** Keeps `pkg/opa` clean, reusable for non-IAM policies in the future, and prevents OPA library abstractions from leaking directly into HTTP controllers.
- **Alternatives Considered:** 
  - *Consolidated in `internal/iam`*: Simpler initial layout, but tightly couples OPA query preparation with IAM HTTP handling.
  - *Standalone OPA daemon/microservice*: Introduces network overhead, extra operational infrastructure, and failure points contrary to the boilerplate's low-footprint philosophy.

### 2. Go-Side Role-to-Permission Expansion
- **Choice:** Go defines the mapping of each role (`Owner`, `Admin`, `Member`, `Viewer`) to its slice of permission strings (`[]string`). The middleware resolves the user's permissions and passes `principal.permissions` into OPA.
- **Rationale:** 
  - Keeps Rego policy clean and minimal (`input.action in input.principal.permissions`).
  - Allows compile-time validation and autocompletion for permission constants in Go.
  - Makes it trivial to extend or cache permissions in the future without modifying Rego files.
- **Alternatives Considered:**
  - *Mapping defined entirely in Rego*: Couples permission taxonomy changes to Rego syntax; Go code still requires permission constants for route decoration.

### 3. Embedded Rego Policy with Pre-compiled Query
- **Choice:** Use Go 1.16+ `//go:embed` to embed `policy/authz.rego` into the binary. Call `rego.PrepareForEval(ctx)` once at application/module startup to compile the AST.
- **Rationale:** Guarantees single-binary deployment with zero runtime file dependencies and sub-millisecond evaluation per request. Allows an optional configuration path override for local policy overrides or testing.
- **Alternatives Considered:**
  - *Filesystem lookup on every request*: Slower and introduces runtime filesystem dependency.
  - *Re-preparing query per evaluation*: Substantial CPU overhead and unnecessary allocation.

### 4. Policy Input Schema & Hybrid Guard Middleware
- **Choice:** The Echo middleware `RequirePermission(action string, opts ...Option)` extracts claims, populates OPA input, and evaluates:
  ```json
  {
    "principal": {
      "identity_id": "usr_...",
      "email": "user@example.com",
      "active_organization_id": "org_...",
      "role": "admin",
      "permissions": ["org:read", "members:invite", "..."]
    },
    "action": "members:invite",
    "resource": {
      "type": "organization",
      "organization_id": "org_..."
    }
  }
  ```
  By default, `resource.organization_id` defaults to `principal.active_organization_id`. An optional extractor (e.g., `WithOrgParam("id")`) can extract the target organization ID from route params (such as `/v1/organizations/:id`).
- **Rationale:** Covers 90% of tenant-scoped routes with zero boilerplate, while providing seamless support for routes where the target resource organization is specified in the URL.
- **Alternatives Considered:**
  - *Manual checks inside controller functions*: High boilerplate and prone to developer omission.

### 5. Rego Policy Logic & Domain Boundary
- **Choice:** Rego rules enforce:
  1. `tenant_match`: `input.principal.active_organization_id == input.resource.organization_id`
  2. `has_permission`: `input.action in input.principal.permissions`
  3. `allow`: `tenant_match && has_permission`
  Fine-grained business state (e.g., preventing demoting an organization's only Owner) is enforced in the domain service layer.
- **Rationale:** Keeps policy evaluation purely stateless and lightning fast without requiring OPA to query the database.

## Risks / Trade-offs

- **[Risk] Increased binary size & dependencies**: Adding OPA Go SDK adds dependencies to `go.mod`.
  - *Mitigation*: The OPA Go SDK is standard, battle-tested, and well-supported in Go. The in-process speed and declarative policy outweigh minor dependency footprint.
- **[Risk] Missing Active Organization Context**: If a user is authenticated but has not selected an active organization, tenant-scoped permission checks fail.
  - *Mitigation*: The guard returns standard HTTP 403 Forbidden `terrors.Forbidden("insufficient permissions")`, prompting the client to switch or select an active organization context.

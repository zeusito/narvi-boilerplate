## 1. OPA Engine Package (`pkg/opa`)

- [x] 1.1 Add `github.com/open-policy-agent/opa` to `backend-api/go.mod` and verify `go mod tidy` succeeds without compilation errors.
- [x] 1.2 Create embedded Rego policy `pkg/opa/policy/authz.rego` with `tenant_match` and `has_permission` rules, verifying formatting and syntax.
- [x] 1.3 Implement generic evaluator `pkg/opa/enforcer.go` using `rego.PrepareForEval(ctx)` and verify with table-driven unit tests in `pkg/opa/enforcer_test.go`.

## 2. IAM Roles, Permissions & Context (`internal/iam`)

- [x] 2.1 Define roles (`Owner`, `Admin`, `Member`, `Viewer`) and permission constants in `internal/iam/permissions.go` adhering to `docs/iam-multi-tenant.md`.
- [x] 2.2 Implement `ExpandRolePermissions(role string) []string` mapping hierarchy in `internal/iam/permissions.go` and verify with unit tests covering each role.
- [x] 2.3 Update `PrincipalClaims` in `internal/iam/claims.go` to include `Permissions []string`, populating it during session introspection.

## 3. Authorization Guard Middleware (`internal/iam`)

- [x] 3.1 Implement `RequirePermission(action string, opts ...Option)` Echo middleware in `internal/iam/authz_middleware.go` mapping claims to OPA input and returning HTTP 403 `terrors.Forbidden`.
- [x] 3.2 Add resource organization extractor support (e.g. `WithOrgParam(param string)`) to `RequirePermission` for routes with URL target parameters.
- [x] 3.3 Wire the OPA enforcer into IAM module initialization/factory (`internal/iam/factory.go`).

## 4. Verification & Integration Tests

- [x] 4.1 Implement comprehensive unit tests in `internal/iam/authz_middleware_test.go` verifying allow, permission denial, cross-tenant denial, and missing active org.
- [x] 4.2 Run full test suite (`go test -v ./pkg/opa/... ./internal/iam/...`) and lint checks to verify zero regressions.

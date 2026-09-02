---
name: module-pattern
description: Guidelines and templates for creating and testing standard internal modules in this project.
---

# Module Pattern Skill

Use this skill when creating a new module or adding functionality to an existing one in `internal/`. This ensures strict adherence to the project architecture: Chi (Router) -> Service -> Repository -> Bun (ORM).

## Module Architecture & Organization

Modules follow a modular architecture. Depending on the size and scope of the module, choose between **Single-Concept** or **Multi-Concept** file organization.

### 1. Single-Concept Modules (Small / Simple)

For focused or small modules, single files per layer are appropriate:

| File                    | Purpose                              | Patterns to Follow                                        |
| :---------------------- | :----------------------------------- | :-------------------------------------------------------- |
| `factory.go`            | **Entry Point.** Wires dependencies. | Use `NewModule(mux *chi.Mux, db *bun.DB)`                 |
| `controller.go`         | **Transport Layer.** HTTP handlers.  | Use `chi` for routing, `render` for responses.            |
| `service.go`            | **Business Logic Interface.**        | Define the `Service` interface and constructor.           |
| `service_default.go`    | **Logic Implementation.**            | Implement `Service`. Use `PrincipalClaims`.               |
| `repository.go`         | **Data Access Interface.**           | Define the `Repository` interface and constructor.        |
| `repository_default.go` | **DB Implementation.**               | Use `uptrace/bun`. Implementation should be isolated.     |
| `models.go`             | **Data Structures.**                 | Domain models (Bun tags), Request/Response DTOs, Filters. |

### 2. Multi-Concept Modules (Large / Complex)

When a module covers multiple sub-domains, entities, or concepts (such as `iam` with Authn/Authz, Organizations, Sessions, Identities, Verifications), organize files using the `<concept>_<layer>.go` naming convention:

| File                      | Purpose                                        | Patterns to Follow                                                                                          |
| :------------------------ | :--------------------------------------------- | :---------------------------------------------------------------------------------------------------------- |
| `factory.go`              | **Entry Point.** Wires module dependencies.    | Initializes repositories, use cases, managers, and controllers.                                             |
| `<concept>_controller.go` | **Transport Layer per Concept.**               | e.g. `authn_controller.go`, `org_controller.go`. Handles routing and HTTP responses.                        |
| `usecases.go`             | **Internal Workflow Registry.**                | Defines internal `*UseCases` interfaces (`authUseCases`, `organizationUseCases`) and constructor functions. |
| `<concept>_usecases.go`   | **Business Logic Implementation per Concept.** | Implements domain workflows (e.g. `authn_usecases.go`, `org_usecases.go`).                                  |
| `<concept>_manager.go`    | **Exported Contract Implementation.**          | Implements public cross-module contracts (e.g. `session_manager.go` implementing `SessionManager`).         |
| `repository.go`           | **Data Access Registry.**                      | Defines sub-repository interfaces (`identityRepository`, `sessionRepository`) and constructors.             |
| `<entity>_repository.go`  | **DB Implementation per Entity/Concept.**      | Implements Bun queries (e.g. `identity_repository.go`, `org_repository.go`, `session_repository.go`).       |
| `<concept>_models.go`     | **Data Structures per Concept.**               | DTOs and domain structs per concept (e.g. `authn_models.go`, `identity_models.go`, `org_models.go`).        |
| `<concept>_middleware.go` | **Module Middleware.**                         | HTTP middlewares or guards (e.g. `authn_middleware.go`, `authz_guard.go`, `claims.go`).                     |

## Naming Conventions for Services & Interfaces

To maintain a clean distinction between **Cross-Module APIs**, **Controller Workflows**, and **Internal Helpers**, adhere strictly to the following interface and component naming rules:

| Category               | Convention  | Visibility                   | Purpose & Examples                                                                                                                                |
| :--------------------- | :---------- | :--------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Exported Contracts** | `*Manager`  | **Exported** (`PascalCase`)  | Public contracts exposed for cross-module consumption.<br>_(e.g., `SessionManager` in `session_manager.go`)_                                      |
| **Internal Workflows** | `*UseCases` | **Unexported** (`camelCase`) | Business logic interfaces orchestrating HTTP controller actions.<br>_(e.g., `authUseCases` in `usecases.go`, implemented in `authn_usecases.go`)_ |
| **Internal Helpers**   | `*er`       | **Unexported** (`camelCase`) | Focused utility or single-responsibility behavioral interfaces.<br>_(e.g., `authenticator`, `introspector`, `hasher`, `courier`)_                 |

### Example (`usecases.go` & `session_manager.go`)

```go
// Exported Contract (Defined and implemented in session_manager.go for cross-module consumption)
type SessionManager interface {
    Introspect(ctx context.Context, token string) PrincipalClaims
}

// Internal Workflows (Defined in usecases.go, implemented in authn_usecases.go, consumed by authn_controller.go)
type authUseCases interface {
    signInWithMagicLink(ctx context.Context, data *NewMagicLinkRequest) error
    verifyMagicLink(ctx context.Context, data *VerifyMagicLinkRequest) (SignInResponse, error)
    signOut(ctx context.Context, claims PrincipalClaims) error
}

// Internal Helper (Single responsibility)
type introspector interface {
    introspectToken(ctx context.Context, token string) PrincipalClaims
}
```

## Implementation Rules

1. **Dependency Injection**: Dependencies like `*bun.DB` or other services must be passed via constructors in `factory.go` or sub-constructors.
2. **Context & Logging**: Every Service/Repo method must accept `context.Context`. Use `toolbox.GetRequestID(ctx)` and `claims.PrincipalID` in every log line.
3. **Error Handling**: Use `pkg/terrors` (e.g., `terrors.RecordNotFound`, `terrors.OperationFailed`). NEVER return raw DB errors from the service layer.
4. **Pagination & Filters**: Controllers should use a `Filters` struct from `models.go` (or `models_<concept>.go`) and pass it to the service.

## Inter-Module Boundaries & Dependencies

To maintain modularity, avoid circular dependencies, and preserve domain invariants, adhere strictly to the following rules:

### 1. Database Table Ownership
- **Strict Isolation:** A module's repository must **ONLY** query and mutate tables owned by its domain.
- **No Cross-Module Joins:** Never write SQL queries that join or mutate tables belonging to another module inside domain repositories.
- **Data Invariants:** Bypassing another module's repository violates its business invariants, validations, and lifecycle hooks.

### 2. Synchronous Cross-Module Communication (Manager / Reader Pattern)
When Module B needs data or capabilities from Module A:
1. **Define an Exported Contract in Module A:** Define an interface (e.g., `SessionManager`, `IdentityReader`) in `<concept>_manager.go` or `<concept>_models.go`.
2. **Use Pure DTOs:** The exported contract methods must accept and return plain DTOs, never exposing internal ORM/Bun entities.
3. **Expose via Module Factory:** Expose the manager via the `*Module` struct returned by `NewModule(...)` (e.g., `func (m *Module) SessionManager() SessionManager`).
4. **Inject in `cmd/main.go`:** Wire dependencies in `cmd/main.go` (the Composition Root):

```go
// cmd/main.go
iamMod := iam.NewModule(myRouter.Mux, dbPool.Conn, mailer, configStore.Iam)
contentMod := content.NewModule(myRouter.Mux, dbPool.Conn, iamMod.SessionManager())
```

### 3. Cross-Domain Reads & Analytics (CQRS / Projections)
When a feature requires joining data across multiple modules (e.g., an admin dashboard displaying user info, billing status, and project counts):
- **Primary Approach (API Composition):** Fetch data from each module's exported manager/reader concurrently in the controller using `errgroup.Group` and stitch the response DTO.
- **High-Performance Reporting:** If complex SQL joins or materialized views are required for performance, create a dedicated, read-only package (e.g., `internal/reporting` or `internal/analytics`). This package acts as a CQRS projection layer and is explicitly read-only, keeping transactional write domains strictly isolated.

## Testing Guidelines

This project prioritizes integration tests over mocking. Tests must use a real Postgres container.

### 1. TestMain Setup

Use `pkg/testbox` to initialize a container. Include `../../db/schema.sql` and a module-specific seed file.

```go
func TestMain(m *testing.M) {
 ctx := context.Background()

 // Infer root directory
 _, filename, _, _ := runtime.Caller(0)
 rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

 // Define paths to SQL files
 schemaPath := filepath.Join(rootDir, "db", "schema.sql")
 testDataPath := filepath.Join(rootDir, "db", "testdata", "iam.sql")

 conn, cleanup, err := testbox.InitPostgresqlContainer(ctx, []string{
  schemaPath,
  testDataPath,
 })

 if err != nil {
  log.Fatal().Err(err).Msg("failed to initialize postgresql container")
 }

 // Defer the cleanup function to close the container
 defer cleanup()

 // Assign to global variable for use in subtests
 testDB = conn

 // Run all tests in this package
 m.Run()
}
```

### 2. Test Style

- Use `github.com/stretchr/testify/assert`.
- Use `t.Context()` for the context.
- Avoid table-driven tests; prefer detailed assertion blocks for readability.
- Always include a "Seed" file in `db/testdata/` for consistency.

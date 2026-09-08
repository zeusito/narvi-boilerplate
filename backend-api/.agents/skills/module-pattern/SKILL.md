---
name: module-pattern
description: Guidelines and templates for creating and testing standard internal modules in this project.
---

# Module Pattern Skill

Use this skill when creating a new module or adding functionality to an existing one in `internal/`. This ensures strict adherence to the project architecture: (Router) -> Controller -> Service -> Repository.

---

## Module Architecture & Layer Responsibilities

Every module follows a strict **Controller-Service-Repository** strategy:

```
[ HTTP Request ]
       │
       ▼
┌──────────────┐
│  Controller  │  Transport: Chi routing, DTO binding, syntactic validation, claims extraction
└──────┬───────┘
       │  ctx (context.Context), Request DTOs
       ▼
┌──────────────┐
│   Service    │  Business Logic: Domain rules, semantic validation, transaction coordination
└──────┬───────┘
       │  ctx (context.Context), Domain Models, bun.Tx / bun.IDB
       ▼
┌──────────────┐
│  Repository  │  Data Access: Bun queries, table ownership, mapping SQL/DB errors -> terrors
└──────┬───────┘
       │
       ▼
[ PostgreSQL Database ]
```

### 1. Controller (Transport Layer)
- **HTTP Routing & Binding:** Handles HTTP routes (`*chi.Mux` or `chi.Router` groups), parses path/query parameters, and binds JSON payloads (using `router.BindBody`) to dedicated **Request DTO** structs.
- **Syntactic Validation:** Validates request structure using `go-playground/validator` struct tags (e.g. `validate:"required,email"`). Returns HTTP 400 immediately if input is structurally malformed.
- **Context Extraction:** Uses standard `r.Context()` and extracts authentication claims/headers via `authz` helpers.
- **Strict Boundary Rule:** **NEVER pass HTTP transport objects (`http.ResponseWriter`, `*http.Request`) into Service or Repository methods.** Always pass standard Go `context.Context` and pure DTOs/types.
- **Response Mapping:** Calls the Service, formats successful results into **Response DTOs** (via `router.RenderJSON`), and returns typed errors via `router.RenderError`.

### 2. Service (Business Logic Layer)
- **Pure Go Logic:** Contains domain business workflows, lifecycle rules, and state transitions. Completely independent of HTTP frameworks or web transports (can run safely in CLI, workers, or tests).
- **Semantic & Domain Validation:** Validates business invariants (e.g. account balance, invitation expiration, permission checks). Returns typed errors from `pkg/terrors` (e.g. `terrors.PreconditionFailed`, `terrors.Forbidden`, `terrors.UnAuthorized`).
- **Orchestration & Transactions:** Coordinates one or more repositories, external clients (mailers, payment processors), or internal helper interfaces. Manages multi-repository atomic transactions when modifying multiple entities.

### 3. Repository (Data Access Layer)
- **Data Encapsulation:** Encapsulates all database interactions using `uptrace/bun`.
- **Table Ownership:** Strictly queries and mutates only tables owned by this module.
- **Strict Error Rule:** **Repositories must NEVER return raw SQL or storage errors** (`sql.ErrNoRows`, unique constraint strings, driver connection errors). All database outcomes must be translated into typed errors from `pkg/terrors`.
- **Transaction Support:** Accepts `bun.IDB` or provides a `WithTx(tx bun.Tx)` helper to participate in transactions coordinated by the service.

### 4. DTOs vs. Domain Models
- **Strict Isolation:** Never bind HTTP requests directly to Bun database models (prevents mass-assignment security vulnerabilities).
- **Request DTOs:** Define the client input contract with validation tags.
- **Response DTOs:** Define the public output representation, shielding internal DB fields (passwords, tokens, internal state machine data).
- **Domain Models:** Define the persistence entities with Bun tags (`bun:"table:identities,alias:i"`).

#### DTO Naming Conventions
Never use generic suffixes like `*DTO`, `*Input`, or `*Output`. Use explicit action- and direction-based names:

| Purpose | Pattern | Examples |
| :--- | :--- | :--- |
| **Incoming Body** | `[Action][Entity]Request` | `CreateOrgRequest`, `VerifyOTPRequest`, `UpdateProfileRequest` |
| **Incoming Query / Filters** | `[Entity]Filters` | `IdentityFilters`, `OrgFilters`, `SessionFilters` |
| **Outgoing Entity** | `[Entity]Response` | `OrgResponse`, `SignInResponse`, `ProfileResponse` |
| **Collection Item / Light View** | `[Entity]Summary` | `OrgMemberSummary`, `IdentitySummary` |
| **Paginated List Response** | `List[Entities]Response` | `ListOrgsResponse`, `ListMembersResponse` |

---

## Module Organization

Depending on the size and scope of the module, choose between **Single-Concept** or **Multi-Concept** file organization.

### 1. Single-Concept Modules (Small / Simple)

For focused or small modules, single files per layer are appropriate:

| File                    | Purpose                              | Patterns to Follow                                                                              |
| :---------------------- | :----------------------------------- | :---------------------------------------------------------------------------------------------- |
| `factory.go`            | **Entry Point.** Wires dependencies. | Use `NewModule(mux *chi.Mux, db *bun.DB, ...)`                                                  |
| `controller.go`         | **Transport Layer.** HTTP handlers.  | Use `chi` for routing, bind/validate DTOs via `pkg/router`, call service, return JSON.         |
| `service.go`            | **Business Logic Interface.**        | Define the `Service` interface and constructor.                                                 |
| `service_default.go`    | **Logic Implementation.**            | Implement `Service`. Orchestrate repository calls and business logic.                           |
| `repository.go`         | **Data Access Interface.**           | Define the `Repository` interface and constructor.                                              |
| `repository_default.go` | **DB Implementation.**               | Use `uptrace/bun`. Map all DB/storage errors to `pkg/terrors`.                                   |
| `models.go`             | **Data Structures.**                 | Domain models (Bun tags), Request/Response DTOs, Filters.                                       |

### 2. Multi-Concept Modules (Large / Complex)

When a module covers multiple sub-domains, entities, or concepts (such as `iam` with Authn/Authz, Organizations, Sessions, Identities, Verifications), organize files using the `<concept>_<layer>.go` naming convention:

| File                      | Purpose                                        | Patterns to Follow                                                                                           |
| :------------------------ | :--------------------------------------------- | :----------------------------------------------------------------------------------------------------------- |
| `factory.go`              | **Entry Point.** Wires module dependencies.    | Initializes repositories, services, and controllers.                                                         |
| `<concept>_controller.go` | **Transport Layer per Concept.**               | e.g. `authn_controller.go`, `org_controller.go`. Handles routing and HTTP responses.                         |
| `services.go`             | **Internal Service Registry.**                 | Defines internal `*Service` interfaces (`authService`, `organizationService`) and constructor helpers.        |
| `<concept>_service.go`    | **Business Logic Implementation per Concept.** | Implements domain business services and exported contracts (e.g. `authn_service.go`, `session_service.go`).   |
| `repository.go`           | **Data Access Registry.**                      | Defines sub-repository interfaces (`identityRepository`, `sessionRepository`) and constructors.              |
| `<entity>_repository.go`  | **DB Implementation per Entity/Concept.**      | Implements Bun queries (e.g. `identity_repository.go`, `org_repository.go`). Maps DB errors to `terrors`.   |
| `<concept>_models.go`     | **Data Structures per Concept.**               | DTOs and domain structs per concept (e.g. `authn_models.go`, `identity_models.go`, `org_models.go`).         |
| `<concept>_middleware.go` | **Module Middleware.**                         | HTTP middlewares or guards (e.g. `authn_middleware.go`, `authz_guard.go`, `claims.go`).                      |

---

## Naming Conventions for Services & Interfaces

Adhere strictly to the following interface and component naming rules:

| Category              | Convention | Visibility                   | Purpose & Examples                                                                                                                              |
| :-------------------- | :--------- | :--------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------- |
| **Exported Services** | `*Service` | **Exported** (`PascalCase`)  | Public domain service contracts exposed for cross-module consumption.<br>_(e.g., `SessionService` in `session_service.go`)_                     |
| **Internal Services** | `*Service` | **Unexported** (`camelCase`) | Business logic interfaces orchestrating domain actions inside the module.<br>_(e.g., `authService` in `services.go`, implemented in `authn_service.go`)_ |
| **Internal Helpers**  | `*er`      | **Unexported** (`camelCase`) | Focused utility or single-responsibility behavioral interfaces.<br>_(e.g., `authenticator`, `introspector`, `hasher`, `courier`)_               |

### Interface Definition Example (`services.go` & `session_service.go`)

```go
// Exported Domain Service (Defined and implemented in session_service.go for cross-module consumption)
type SessionService interface {
    Introspect(ctx context.Context, token string) (*PrincipalClaims, error)
}

// Internal Service (Defined in services.go, implemented in authn_service.go, consumed by authn_controller.go)
type authService interface {
    SendOTP(ctx context.Context, req *SendOTPRequest) error
    VerifyOTP(ctx context.Context, req *VerifyOTPRequest) (*SignInResponse, error)
}

// Internal Helper (Single responsibility)
type introspector interface {
    introspectToken(ctx context.Context, token string) (*PrincipalClaims, error)
}
```

---

## Repository Layer & Error Encapsulation

### Strict Rule: Never Return Storage/SQL Errors

Repositories are the boundary between domain data requirements and physical database storage. **They must never leak database internals or SQL errors.**

- **Forbidden:** Returning `sql.ErrNoRows`, `bun.ErrNoRows`, Postgres driver constraint errors (e.g. unique violation error strings), or connection errors directly to the caller.
- **Mandatory:** Map all database errors into typed domain errors using `backend-api/pkg/terrors`.

### Error Translation Mapping

| Database Outcome                                  | Typed Error from `pkg/terrors`                             | Description / Example                                          |
| :------------------------------------------------ | :--------------------------------------------------------- | :------------------------------------------------------------- |
| `errors.Is(err, sql.ErrNoRows)`                   | `terrors.RecordNotFound("entity not found")`               | Record does not exist for the given query criteria.            |
| Duplicate key / unique constraint violation       | `terrors.RecordAlreadyExists("entity already exists")`     | Conflict on unique field (e.g. email, slug, external ID).      |
| Foreign key constraint violation                  | `terrors.PreconditionFailed("referenced record missing")`  | Related foreign entity does not exist.                         |
| Connection error, query syntax, or mutation error | `terrors.OperationFailed("storage operation failed")`      | Query execution failed; log raw error internally with zerolog. |

### Exemplary Repository Implementation with Transaction Support

```go
package iam

import (
    "context"
    "database/sql"
    "errors"
    "strings"

    "backend-api/pkg/terrors"

    "github.com/rs/zerolog/log"
    "github.com/uptrace/bun"
    "github.com/uptrace/bun/driver/pgdriver"
)

type defaultIdentityRepo struct {
    db bun.IDB
}

func newIdentityRepository(db bun.IDB) identityRepository {
    return &defaultIdentityRepo{db: db}
}

// WithTx returns a new repository instance participating in an active transaction
func (r *defaultIdentityRepo) WithTx(tx bun.Tx) identityRepository {
    return &defaultIdentityRepo{db: tx}
}

func (r *defaultIdentityRepo) Create(ctx context.Context, identity *Identity) error {
    _, err := r.db.NewInsert().
        Model(identity).
        Returning("*").
        Exec(ctx)

    if err != nil {
        var pgErr pgdriver.Error
        if errors.As(err, &pgErr) && pgErr.IntegrityViolation() {
            return terrors.RecordAlreadyExists("identity with this email already exists")
        }

        log.Error().Err(err).Str("email", identity.Email).Msg("failed to insert identity into database")
        return terrors.OperationFailed("failed to create identity")
    }

    return nil
}

func (r *defaultIdentityRepo) FindActiveByEmail(ctx context.Context, email string) (*Identity, error) {
    var identity Identity
    err := r.db.NewSelect().
        Model(&identity).
        Where("lower(email) = ?", strings.ToLower(strings.TrimSpace(email))).
        Where("state = ?", IdentityStateActive).
        Scan(ctx)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, terrors.RecordNotFound("identity not found")
        }

        log.Error().Err(err).Str("email", email).Msg("failed to query identity by email")
        return nil, terrors.OperationFailed("failed to retrieve identity")
    }

    return &identity, nil
}
```

---

## Transaction Management across Repositories

When a domain business operation involves mutating multiple database entities atomically (e.g. creating an Organization and its initial Owner Identity):

1. **Orchestrated by the Service:** The Service controls transaction boundaries because business logic determines what constitutes an atomic unit of work.
2. **Coordinated via `bun.DB.RunInTx`:** The service passes the active transaction `bun.Tx` to repositories using their `WithTx(tx)` helper.
3. **Automatic Rollback on `terrors`:** If any repository operation returns an error (e.g. `terrors.RecordAlreadyExists`), returning that error aborts the transaction cleanly.

```go
func (s *defaultOrgService) RegisterOrganization(ctx context.Context, req *RegisterOrgRequest) (*OrgResponse, error) {
    var createdOrg *Organization

    err := s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
        org := &Organization{Name: req.Name}
        if err := s.orgRepo.WithTx(tx).Create(ctx, org); err != nil {
            return err // Rollback & propagate typed terror
        }

        identity := &Identity{Email: req.OwnerEmail, OrgID: org.ID}
        if err := s.identityRepo.WithTx(tx).Create(ctx, identity); err != nil {
            return err // Rollback & propagate typed terror
        }

        createdOrg = org
        return nil
    })

    if err != nil {
        return nil, err // Typed terror propagated directly to controller
    }

    return toOrgResponse(createdOrg), nil
}
```

---

## Validation Boundaries: Transport vs. Domain

To keep code maintainable and prevent logic leakage, strictly separate **Transport Validation** from **Domain Validation**:

| Category | Where | Technology / Pattern | Purpose & Examples |
| :--- | :--- | :--- | :--- |
| **Transport Validation** | **Controller** | `go-playground/validator` tags on DTOs | **Syntactic & Structural Checks:** Required fields, string lengths, UUID formats, email syntax, numeric ranges.<br>_(e.g., `validate:"required,email"`)_ |
| **Domain Validation** | **Service** | Pure Go business logic, `pkg/terrors` | **Semantic & State Checks:** Account balances, expired tokens, duplicate slug within tenant, business invariants.<br>_(e.g., `return terrors.Forbidden("plan limit reached")`)_ |

---

## Implementation Rules

1. **Dependency Injection**: Dependencies (`*bun.DB`, services, repositories) must be passed explicitly via constructors in `factory.go` or package sub-constructors.
2. **Context & Structured Logging**: Every Service and Repository method must accept standard `context.Context` as its first parameter. Log operations with Zerolog using contextual tags (e.g. `request_id`, `principal_id`).
3. **No Framework Leakage**: Controllers extract `r.Context()` and pass pure DTOs. Never import `github.com/go-chi/chi/v5` inside service or repository files.
4. **Controller-Service-Repository Flow**:
   - Controllers never touch repositories; they call services.
   - Services orchestrate business logic and interact with repositories.
   - Repositories interact only with the database and return typed `terrors`.
5. **Error Handling Across Layers**:
   - Repositories return `*terrors.Terror` (e.g. `RecordNotFound`, `RecordAlreadyExists`, `OperationFailed`).
   - Services validate domain rules and propagate or create `*terrors.Terror` (e.g. `UnAuthorized`, `Forbidden`, `PreconditionFailed`).
   - Controllers handle HTTP errors using `router.RenderError(ctx, w, err)` for standardized HTTP response serialization.
6. **Pagination & Filters**: Controllers parse query params into a `Filters` struct from `models.go` (or `<concept>_models.go`) and pass it to the service layer.

---

## Inter-Module Boundaries & Dependencies

To maintain modularity, avoid circular dependencies, and preserve domain invariants, adhere strictly to the following rules:

### 1. Database Table Ownership
- **Strict Isolation:** A module's repository must **ONLY** query and mutate tables owned by its domain.
- **No Cross-Module Joins:** Never write SQL queries that join or mutate tables belonging to another module inside domain repositories.
- **Data Invariants:** Bypassing another module's repository violates its business invariants, validations, and lifecycle hooks.

### 2. Synchronous Cross-Module Communication (Service Pattern)
When Module B needs data or capabilities from Module A:
1. **Define an Exported Service Interface in Module A:** Define an interface (e.g., `SessionService`, `IdentityService`) in `<concept>_service.go` or `services.go`.
2. **Use Pure DTOs:** The exported service methods must accept and return plain DTOs or domain models, never exposing internal ORM/Bun entities or database handles.
3. **Expose via Module Factory:** Expose the service via the `*Module` struct returned by `NewModule(...)` (e.g., `func (m *Module) SessionService() SessionService` or field `SessionService SessionService`).
4. **Inject in `cmd/main.go`:** Wire dependencies in `cmd/main.go` (the Composition Root):

```go
// cmd/main.go
iamMod := iam.NewModule(myRouter.Mux, dbPool.Conn, mailer, configStore.Iam)
contentMod := content.NewModule(myRouter.Mux, dbPool.Conn, iamMod.SessionService)
```

### 3. Acyclic Dependencies (No Circular Module Imports)
- Dependencies between modules must always flow in one direction.
- If Module A needs Module B, Module B must **never** import Module A.
- If bidirectional communication is required, either:
  1. Invert the dependency using **Domain Events / Asynchronous Pub-Sub** (e.g. Module A publishes an event that Module B listens to).
  2. Or extract shared concepts into a dedicated shared module.

### 4. Cross-Domain Reads & Analytics (CQRS / Projections)
When a feature requires joining data across multiple modules (e.g., an admin dashboard displaying user info, billing status, and project counts):
- **Primary Approach (API Composition):** Fetch data from each module's exported service concurrently in the service/controller layer using `errgroup.Group` and stitch the response DTO.
- **High-Performance Reporting:** If complex SQL joins or materialized views are required for performance, create a dedicated, read-only package (e.g., `internal/reporting` or `internal/analytics`). This package acts as a CQRS projection layer and is explicitly read-only, keeping transactional write domains strictly isolated.

---

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

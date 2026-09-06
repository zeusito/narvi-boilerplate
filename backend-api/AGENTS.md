# AI Agents

This document provides context and guidelines for AI agents working on this project.

## Project Overview

This is our main API for our platform. We aim for code that is highly maintainable, testable, and scalable. Stability is our priority.

- **Frameworks**: Echo v5 (Router), Bun (ORM), Zerolog (Logging).
- **Structure**:
  - `cmd/`: Application entry points.
  - `internal/`: Core business logic.
  - `pkg/`: Shared infrastructure and tools.

**File Organization & Architecture:**
We follow a modular architecture with strict separation of concerns under `internal/`. For full details on module layout (single-concept vs. multi-concept), file naming conventions, and layer templates, refer to the [module-pattern](.agents/skills/module-pattern/SKILL.md) skill.


## Coding Conventions

- **Interface Naming:**
  - Exported cross-module services: `*Service` (e.g. `SessionService`, `IdentityService`).
  - Unexported internal services: `*Service` (e.g. `authService`).
  - Unexported internal helpers: `*er` (e.g. `authenticator`, `introspector`).
- **DTO Naming:** Use `[Action][Entity]Request` for request bodies, `[Entity]Response` for outputs, `[Entity]Filters` for query parameters, and `[Entity]Summary` for list items. Never use generic suffixes like `*DTO` or `*Input`.
- **Dependency Injection:** Pass dependencies (DB, other services) into the `NewModule` or `NewService` constructors.
- **Router:** We use `echo`. Controllers should accept `*echo.Echo` (or `echo.Router`) and register their own sub-routes.
- **Database:** We use `uptrace/bun`. Repositories should accept `*bun.DB` (or `bun.IDB`). Repositories must never leak SQL or storage errors and must return typed errors from `pkg/terrors`.
- **Error Handling:** Use `pkg/terrors` for typed errors across repository, service, and controller layers.
- **Idiomatic Go**: Follow standard Go practices and patterns.
- **Separation of Concerns**: Keep business logic in `internal/` and infrastructure in `pkg/`.
- **Consistency**: Match the existing coding style in the repository.
- **Automation**: Use the `Makefile` for tasks like building, running, and testing.

## Testing Rules

- **Required for Go changes:** run `make test` (or `go test -v ./... --race`) before handing off work.
- **Scoped changes:** also run tests for the modified package(s) when feasible (e.g., `go test ./internal/wallets`).
- **DB-dependent tests:** if you add or modify tests that require Postgres, document the setup in the PR or task notes and ensure migrations are up to date.
- **Test style:** avoid table-driven tests; prefer assertion-style tests.
- **Avoid mocking:** prefer testing against a real Postgres instance (e.g., using `testcontainers`) rather than mocking the database layer. If mocking is necessary, use interfaces and dependency injection to facilitate testing.

## Task Execution

Before starting a task:

1. Review `README.md` for project context.
2. Check existing patterns in `internal/iam` or `internal/healthcheck`.
3. Verify changes using `make test` and `make lint`.

# Instructions for agents

This file is the operational entry point for any AI agent working on this Go backend. It is deliberately short — the detail lives in `docs/ai/`.

## Project goal

This repository is a Go backend service (typically a REST API), built with hexagonal architecture, consuming the shared library `github.com/raykavin/gobox` for cross-cutting concerns (logging, config, HTTP response, pagination, OIDC, database, healthcheck). Confirm the specific business purpose in the repository's `README.md` before acting.

## Sources of truth (in this order)

1. Explicit requirements of the current task.
2. Local repository instructions (`.claude/`, `CLAUDE.md`, the project's own `README.md`) — always take priority over this document.
3. Already-implemented public contracts (HTTP routes, database schema, response format) — don't break them without authorization.
4. This `AGENTS.md`.
5. Topic-specific documents in `docs/ai/instructions/` (go, api, database, testing, security).
6. `docs/ai/backend-development-standard.md`.
7. Conventions inferred from existing code in the repository itself.
8. General engineering recommendations.

If two sources of the same priority level conflict, **stop and ask for a decision** — don't choose silently.

## How to start a task

1. Read the repository's `README.md` and any local `.claude/`/`CLAUDE.md` — they take priority over this file.
2. Look for an equivalent implementation already in the repository before creating a new abstraction (e.g., before writing pagination, look for `gobox/pagination`; before writing an HTTP response envelope, look for `gobox/httpserver/respond`).
3. Before adding a new external dependency, check whether an equivalent solution already exists in `github.com/raykavin/gobox` or another dependency already present in `go.mod`.

## Architecture

Hexagonal / ports-and-adapters. See `docs/ai/backend-development-standard.md`, section "Standard architecture", for the full directory tree.

## Directory structure

```
cmd/<binary>/main.go      — bootstrap only: flags, config, calls internal/app, starts/stops the process
internal/domain/          — entities and rules, no I/O
internal/port/            — interfaces (contracts)
internal/usecase/         — 1 file per business operation
internal/adapter/inbound/http/{handler,middleware,response,router}
internal/adapter/outbound/<x>/   — concrete implementations of port
internal/app/ (or internal/platform/) — manual DI, called once by cmd/*/main.go
internal/config/          — config loading/validation
scripts/sh/*.sh           — build/deploy/test/swagger/healthcheck, invoked by the Makefile
```

## Dependency rules

- `domain` → nothing from `adapter`/`infra`/`config`/`app`.
- `usecase` → only `domain` and `port` (interfaces), never `adapter/outbound` directly.
- `adapter/outbound/*` → implements `port`; the only place with GORM/drivers/third-party clients.
- `adapter/inbound/http/handler` → validates input, calls 1 usecase, formats output; **never** implements business rules.
- Composition/DI → centralized in `internal/app/` (or `internal/platform/`), manual, no DI framework (don't introduce Uber Fx, Wire, etc. without an explicit decision).

## Implementation conventions

- 1 file per operation in `usecase/` (`create.go`, `update.go`, `find_or_create.go`), `snake_case.go`.
- Errors: local sentinels (`var ErrX = errors.New(...)`) + `fmt.Errorf("%w: ...", ErrX)`. No central error package.
- DTOs validated via `binding:"required"` (Gin/go-playground/validator), never standalone `validate:`.
- Logging via `github.com/raykavin/gobox/logger` — never write a new zerolog wrapper.
- Pagination via `github.com/raykavin/gobox/pagination` — never reimplement manual `LIMIT`/`OFFSET` in a project that already uses GORM.

## APIs and contracts

- Framework: Gin. Fixed prefix `/api/v1`. Resources in plural kebab-case.
- `GET /resource/:id` for reading by ID; `POST /resource/search` (filter in the body) for search/listing.
- Authentication: **always** `github.com/raykavin/gobox/oidcauth` + `gobox/httpserver/middlewares`. Never implement OIDC/JWT verification from scratch.
- Authorization: declarative per route (middleware), never inside the usecase. Permission named `"resource:action"`.
- Every new endpoint must have tests for: authorization (access denied without permission), success, invalid input, and non-existent resource (when applicable).
- Explicitly map domain errors → HTTP status in the handler (`errors.Is`/`errors.As`); never let every error fall through to 500 by default.
- Keep Swagger annotations (`@Success`/`@Failure`) in sync with the handler's actual behavior.

## Persistence and migrations

- ORM: GORM over Postgres/SQLite is the standard for a general relational domain.
- **Don't modify a migration that's already been applied; create a new one** — unless the project uses `db.AutoMigrate` without versioned SQL (acceptable in an early stage; evaluate the risk of running it in production — see `docs/ai/instructions/database.md`).
- Every new environment/config variable must be validated at startup and documented in the `*.example.yml` file.
- Any schema change that removes or renames a column, or that runs against a large table, **requires explicit human approval** — see the section below.
- Multi-step transactions: use a `TransactionRunner` propagated via `context.Context`; don't write multi-step usecases without a transaction.

## Security

See `docs/ai/instructions/security.md` for the full list. Non-negotiable: authentication always via `gobox/oidcauth`; CORS always via `gobox/httpserver/middlewares`; CSRF required with a session cookie; secrets only via a versioned `*.example.yml` + the real file in `.gitignore` — never commit a real value.

## Testing and validation

Minimum sequence before considering any task complete:
```
go build ./...
go vet ./...
golangci-lint run     # if the repository has a .golangci.yml
go test -v -race ./...
```
If a step can't be run (e.g., `golangci-lint` not installed in the environment), state that clearly in the final report — don't omit the failure or pretend it passed. Don't remove or reduce a test just to make validation pass. Prefer hand-written functional mocks or SQLite in-memory for tests involving GORM/transactions, rather than relying on automatic mock generation.

## Documentation

Update `README.md` (sections: Architecture, Configuration, Local execution, Scripts/Makefile, Tests, API documentation) and Swagger annotations in the same PR that changes the behavior they describe. If the repository has its own `.claude/`/`CLAUDE.md`, update it too when the change affects what it describes.

## Operations that require explicit human approval

Stop and ask for confirmation before:
- Any incompatible API change (changing response format, removing a field, changing an existing endpoint's HTTP verb).
- Changing the authentication/authorization mechanism.
- A migration that removes/renames a column, or `DROP`/`ALTER` on a table with real data.
- Deleting relevant files or modules.
- Changing framework, database, or structural technology (e.g., moving away from GORM, swapping Gin for another framework).
- Adding a high-impact dependency (new ORM, new DI framework, new message broker).
- Changing shared infrastructure (`gobox`) in a way that affects other consumers without coordinating the version.
- Changing the build/deploy pipeline in production.
- Reducing test coverage or removing an existing test.
- Any broad change outside the requested scope.

## Forbidden actions

- Don't invent a requirement or architectural decision that wasn't requested.
- Don't locally copy/reimplement something that already exists in `gobox`.
- Don't expose secrets in code, logs, documentation, or responses — not even "temporarily" or in an example file.
- Don't run `--no-verify`, skip hooks, or disable lint/tests to make a task "pass".

## Definition of done

A task is complete when: (1) build/vet/lint/test pass (see Testing section); (2) public contracts preserved or the break explicitly authorized; (3) relevant documentation updated; (4) no operation from the "requires approval" list was executed without confirmation; (5) the agent clearly reported what changed, why, and how it was validated.

## Related documents

- [docs/ai/backend-development-standard.md](docs/ai/backend-development-standard.md) — detailed development standard.
- [docs/ai/new-project-checklist.md](docs/ai/new-project-checklist.md) — project kickoff checklist.
- [docs/ai/feature-template.md](docs/ai/feature-template.md) — recommended flow for implementing a feature.
- [docs/ai/instructions/](docs/ai/instructions/) — go.md, api.md, database.md, testing.md, security.md (per-topic detail).

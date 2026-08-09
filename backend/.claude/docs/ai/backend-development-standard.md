# Backend development standard (Go)

Reference standard for Go backend services built from this template.

## Engineering principles

1. Simplicity and readability over speculative abstraction — manual and explicit composition, no DI framework.
2. Reuse via `gobox` (`github.com/raykavin/gobox`) is the declared intent — before writing a logger, config loader, HTTP envelope, pagination, OIDC client, or database wrapper, check whether it already exists in `gobox`.
3. Smallest coherent change: usecases are one file per operation (`create.go`, `update.go`, `find_or_create.go`) — don't group unrelated operations in the same file/function.

## Standard architecture

Hexagonal / ports-and-adapters, with this folder split under `internal/`:

```
internal/
  domain/    GORM entities (BaseModel: ID/CreatedAt/UpdatedAt/DeletedAt), value objects, invariant rules — NO import of adapter/infra/config
  port/      interfaces: repository and external-service contracts, consumed by usecase
  usecase/   1 file per business operation; depends only on port, never directly on adapter/outbound
  adapter/
    inbound/http/   handler (bind+validate+delegate), middleware, response (envelope), router
    outbound/<x>/    concrete implementation of port (postgres, external integrations)
  app/       manual DI: bootstrap.go wires everything and hands ready-to-use dependencies to cmd/*/main.go
  config/    configuration loading and validation per entry point (api, worker)
  dto/       HTTP-layer request/response objects
```

**Dependency rules (forbidden/allowed):**
- `domain` doesn't import anything from `adapter`, `infra`, `config`, or `app`.
- `usecase` imports `domain` and `port`; never imports `adapter/outbound` directly (only the interface in `port`).
- `adapter/outbound/*` implements the `port` interfaces; it's the only place where GORM/database drivers/third-party HTTP clients appear.
- `adapter/inbound/http/handler` doesn't implement business rules — validates input, calls 1 usecase, formats the output.
- Wiring all of this together is centralized in `internal/app/` (or `internal/platform/`), called once in each `cmd/*/main.go`.

**Dependency injection: manual, no framework.** Don't introduce Uber Fx, Wire, or equivalent without an explicit decision.

**Multiple binaries under `cmd/`.** Each process with a distinct lifecycle (HTTP API, async worker, migrator, diagnostic CLI) is its own binary at `cmd/<name>/main.go`, all consuming the same `internal/app`. Don't put business logic in `main.go` beyond: parsing flags, loading config, calling bootstrap, starting/stopping the process.

**Two HTTP servers per service.** A business server (routes `/api/v1/...`) and a health-check server on a separate, independently configurable port. Prefer, when practical, exposing `/healthz` (liveness) and `/readyz` (readiness) as distinct endpoints on the health server.

## Recommended structure for a new project

See [new-project-checklist.md](new-project-checklist.md) for the full step-by-step. Summary of the initial tree:
```
cmd/api/main.go
cmd/worker/main.go        # if there is async processing
configs/{api,worker}.example.yml
internal/{domain,port,usecase,adapter,app,config,dto}/
scripts/sh/{build,deploy,test,swagger,healthcheck}.sh
Makefile
Dockerfile
.devcontainer/
.golangci.yml
```

## Code conventions

| Topic | Rule |
|---|---|
| File name | `snake_case.go`, one file per operation in `usecase/` |
| DTOs | `binding:"required"` (Gin/validator tags), never standalone `validate:` |
| Domain errors | local sentinels (`var ErrX = errors.New(...)`) + `fmt.Errorf("%w: ...", ErrX)`; no central error package |
| Entity ID | `uint` autoincrement via an embedded `BaseModel` is the standard; use `uuid.UUID` when the ID needs to be generated/exposed externally without coordination |
| Entity base | `BaseModel{ID, CreatedAt, UpdatedAt time.Time, DeletedAt gorm.DeletedAt}` embedded — see `instructions/database.md` |
| Pagination | `github.com/raykavin/gobox/pagination` — `Page`/`Result[T]`/`Scope` — never reimplement |
| Logging | `github.com/raykavin/gobox/logger` (zerolog underneath) — never create a new local wrapper |
| HTTP response | `github.com/raykavin/gobox/httpserver/respond` — import, don't copy |

Full detail per topic in `instructions/go.md`.

## APIs

- Framework: **Gin**. Fixed route prefix `/api/v1`. Resources in plural kebab-case (`bank-accounts`, `document-types`).
- Read by ID: `GET /resource/:id`. Search with filter: `POST /resource/search` with the filter in the body — this is the pattern to follow even though it looks non-RESTful, because it's what `gobox/pagination` expects (a complex filter doesn't fit well in a querystring).
- Update: `PATCH` by default; use `PUT` only when the entire resource is replaced.
- Authentication: `github.com/raykavin/gobox/oidcauth` + `gobox/httpserver/middlewares`. **Never create a new OIDC verifier from scratch.**
- Authorization: declarative per route/group, never inside the usecase. Permission named `"resource:action"` (e.g., `"invoice:validate"`) for human users; OAuth scope (`"service:webhook:write"`) for M2M clients.
- Validation: `go-playground/validator` via Gin tags. Handle the bind error explicitly — don't let the validator's raw message (in English) leak to the client; parse it field-by-field.
- Pagination: always via `gobox/pagination`. Never implement manual `LIMIT`/`OFFSET` in new code if the project already uses GORM.
- Ingestion idempotency (file upload, import, webhook): SHA-256 hash of the content + composite unique index with the parent scope (`(parent_id, file_hash)`) + find-or-create flow in the usecase. Use GORM's `clause.OnConflict{DoNothing:true}` when you don't need to return the existing record.
- Documentation: `swaggo/swag`, annotations on handlers, generated via `make back-swagger`. **Keep the `@Failure`/`@Success` annotations in sync with the handler's actual `switch`/error mapping.**

## Data

- ORM: GORM over Postgres (production) / SQLite (development) is the standard for services with a relational domain. Plain pgx without an ORM is acceptable only when the entity is simple (few columns, no relations) and there's strong performance/control pressure — it's not the standard for new general relational-domain services.
- Migrations vs. `AutoMigrate`: `db.AutoMigrate` at boot is acceptable in an early/low-risk stage, but has real risk in production (no rollback, no version history, index creation doesn't run `CONCURRENTLY`). For critical data, prefer versioned migrations (`golang-migrate` + a dedicated `migrator` binary).
- Transactions: use a `TransactionRunner`/Unit-of-Work propagated via `context.Context`. Don't write multi-step usecases without a transaction.
- N+1: use GORM's `Preload` to hydrate associations on reads; reserve manual `Joins` for search filters.
- Soft delete: `gorm.DeletedAt` embedded in `BaseModel`. If an entity has a unique index that needs to ignore soft-deleted rows, use a partial unique index (`uniqueIndex:...,where:deleted_at IS NULL`).

Full detail in `instructions/database.md`.

## Errors

- Local sentinels per package (`var ErrNotFound = errors.New(...)`), never a central exceptions package.
- Wrap with `fmt.Errorf("%w: %w", ...)` (double-wrap, Go 1.20+) or `errors.Join` when applicable.
- In the HTTP handler, explicitly map the domain error to an HTTP status (`errors.Is`/`errors.As`) — never let every error fall through to 500 by default.

## Logging

- `github.com/raykavin/gobox/logger` (zerolog underneath), configurable (console/JSON). Never write a new zerolog wrapper.
- Never log a full webhook/integration payload carrying personal data at `debug` level without redacting sensitive fields.

## Security

See `instructions/security.md` for the full list. Non-negotiable rules:
- Authentication always via `gobox/oidcauth` — never reimplement JWT verification/introspection.
- CORS always via `gobox/httpserver/middlewares` (current version, allowlist) — never copy/reimplement CORS locally.
- CSRF required on any route that accepts a session cookie.
- Secrets: only via a versioned `configs/*.example.yml` + the real file in `.gitignore`. Never commit a file with a real value, not even as a "temporary example".

## Tests

- Minimum command: `go test -v -race ./...`.
- `go build ./...` and `go vet ./...` before the test.
- A `.golangci.yml` with `govet, errcheck, staticcheck, ineffassign, unused, gofmt, goimports` as a starting point for new projects.
- Mocks: prefer hand-written functional mocks (structs with `func` fields) or SQLite in-memory for repository tests involving a real transaction/GORM, rather than relying on automatic mock generation.

Full detail in `instructions/testing.md`.

## Documentation

- README with the sections: Architecture, Configuration, Local execution, Scripts/Makefile, Tests, API documentation.
- AI-agent documentation (`.claude/`, `AGENTS.md`) must be updated in the same PR that changes the structure it describes.

## Observability

- Structured logging via `gobox/logger` is the mandatory baseline.
- OpenTelemetry+Prometheus via `gobox/telemetry` is recommended when the project needs tracing/metrics, instead of reimplementing it.

## Infrastructure

- Multi-stage Dockerfile: build on `golang:1.26-alpine3.22`, runtime on `alpine:3.22`, non-root user, `HEALTHCHECK` via `scripts/sh/healthcheck.sh`. **Confirm the referenced script exists** before considering the build ready.
- `scripts/sh/build.sh` must point to the Dockerfile's real path. Test `make back-build` before considering a new project ready.
- `make {build,deploy,test,mocks,swagger}` delegating to `scripts/sh/*.sh` is the standard automation pattern.
- **CI/CD**: a new project should include a basic pipeline (lint + build + test) from the start. See `new-project-checklist.md`.

## Task completion criteria

A change is only complete when:
1. `go build ./...` with no error.
2. `go vet ./...` with no error.
3. `golangci-lint run` with no error, if the project has a `.golangci.yml` (or one was added as part of the task).
4. `go test -v -race ./...` with no failure, including new/updated tests for the changed behavior.
5. Public contracts (HTTP routes, database schema, response format) preserved, unless explicitly authorized.
6. Documentation (README/Swagger/godoc comment) updated if the documented behavior changed.

## Decisions requiring human approval

See the "Operations that require approval" section in [../../AGENTS.md](../../AGENTS.md).

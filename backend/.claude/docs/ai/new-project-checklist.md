# Backend project kickoff checklist

Use as a starting point for a new Go backend service built from this template. Items marked **[decision]** require confirmation from the person responsible before proceeding — don't decide alone.

## 1. Architectural decisions

- [ ] Confirm the service's domain/purpose.
- [ ] **[decision]** Entity ID type: `uint` autoincrement (recommended default) or `uuid.UUID` — decide based on whether IDs will be exposed/generated externally.
- [ ] **[decision]** Schema strategy: `db.AutoMigrate` at boot (simpler, no history/rollback) vs. versioned SQL migrations + a `migrator` binary (safer for critical data).
- [ ] Confirm whether there will be async processing — if so, plan `cmd/worker` from the start.
- [ ] Manual DI via `internal/app/` (bootstrap/factory) — don't introduce a DI framework (Fx, Wire) unless explicitly decided otherwise.

## 2. Initial structure

- [ ] Create the tree: `cmd/api/`, `cmd/worker/` (if applicable), `internal/{domain,port,usecase,adapter/{inbound/http,outbound},app,config,dto}/`.
- [ ] `Makefile` with `build, deploy, test, swagger` targets delegating to `scripts/sh/*.sh` — **verify that the paths referenced by the scripts actually exist**.
- [ ] `.golangci.yml` (`govet, errcheck, staticcheck, ineffassign, unused, gofmt, goimports`).
- [ ] `.devcontainer/` bringing up only the auxiliary services actually used (Postgres if GORM; don't copy observability services "just in case" unless the new project will actually instrument telemetry).

## 3. Configuration

- [ ] `configs/api.example.yml` (+ `worker.example.yml` if applicable) versioned, with placeholders (`"changeme"`), YAML.
- [ ] The real file (`configs/api.yml`) in `.gitignore` from the first commit — confirm the `.gitignore` rule doesn't have an accidental exception.
- [ ] Every configuration variable validated at startup (fail fast if required and missing) and documented in the `.example.yml`.
- [ ] Config loaded via `github.com/raykavin/gobox/config` (`Loader[T]` over Viper) — don't write a new loader.

## 4. Secrets management

- [ ] No real value in any versioned file, not even in a "temporary configuration" commit.
- [ ] Runtime secrets (OIDC client secret, database DSN, partner token) only via a real (ignored) config file or an environment variable injected at deploy time.

## 5. Database

- [ ] Connection via `github.com/raykavin/gobox/database/gorm` (factory with retry+pool), not `gorm.Open` directly.
- [ ] `BaseModel` with `ID/CreatedAt/UpdatedAt/DeletedAt(gorm.DeletedAt)` embedded in every persisted entity — ID type per the item 1 decision.
- [ ] If choosing versioned migrations: use `github.com/raykavin/gobox/database/migrate`, a dedicated `cmd/migrator` binary, real `.up.sql`/`.down.sql` files (not empty placeholders) from the first migration.
- [ ] Unit-of-Work: implement `port.TransactionRunner` before the first multi-step usecase that writes to more than one table.
- [ ] Pagination via `gobox/pagination`, not a custom implementation.

## 6. Authentication and authorization

- [ ] Authentication via `github.com/raykavin/gobox/oidcauth` — don't write your own OIDC verifier.
- [ ] CORS/CSRF via `github.com/raykavin/gobox/httpserver/middlewares` — don't reimplement.
- [ ] Define the permission naming convention (`"resource:action"`) from the first endpoint, documented in a single file (`internal/domain/valueobject/permissions.go` or equivalent).

## 7. Logging

- [ ] `github.com/raykavin/gobox/logger` — don't create a local zerolog wrapper.
- [ ] Decide from the start which payload/request fields are sensitive (PII, credentials) and ensure redaction before logging at debug level.

## 8. Error handling

- [ ] Local error sentinels per package (`var ErrX = errors.New(...)`), no central exceptions package.
- [ ] Explicit domain-error → HTTP-status mapping in every handler, from the first endpoint (don't let it fall through to 500 by default).

## 9. Health checks

- [ ] Health-check server on a port separate from the business server.
- [ ] Prefer distinct `/healthz` (liveness) and `/readyz` (readiness) endpoints.
- [ ] Real, tested `scripts/sh/healthcheck.sh` — confirm the `Dockerfile` only `COPY`s a script that actually exists.

## 10. Tests

- [ ] `go test -v -race ./...` working from the first commit.
- [ ] `.golangci.yml` configured and `golangci-lint run` passing.
- [ ] Authorization/success/invalid-input/not-found tests for the first endpoint, as a model for the ones that follow.
- [ ] Prefer hand-written functional mocks over relying on automatic generation (`mockgen`) without first testing whether it actually produces useful artifacts.

## 11. Quality

- [ ] `go build ./... && go vet ./... && golangci-lint run && go test -v -race ./...` as the validation sequence documented in the README.
- [ ] Include a basic CI pipeline (lint+build+test) from the start.

## 12. Documentation

- [ ] `README.md` with the sections: Architecture, Configuration, Local execution, Scripts/Makefile, Tests, API documentation.
- [ ] Swagger (`swaggo/swag`) from the first endpoint, generated via `make back-swagger`.
- [ ] The project's own `AGENTS.md`/`.claude/` (if the team uses AI agents on it) kept in sync with the real structure from the start.

## 13. Containers

- [ ] Multi-stage Dockerfile `golang:1.26-alpine3.22` → `alpine:3.22`, non-root user, `HEALTHCHECK` via a real script.
- [ ] Test `make back-build` and `docker build` locally before considering the project ready.

## 14. Pipeline and deploy

- [ ] Basic CI (lint+build+test) — see item 11.
- [ ] `make back-deploy` → `scripts/sh/deploy.sh` (`docker push`) as the starting standard; introduce Kubernetes/Helm/automated rollback only on explicit request.

## 15. Readiness criteria

See the "Definition of done" section in `../../AGENTS.md`. Summary: build/vet/lint/test pass; contracts preserved; documentation updated; no operation from the "requires approval" list executed without confirmation.

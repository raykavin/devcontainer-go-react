# Specific instructions — Go

Complements `AGENTS.md` and `backend-development-standard.md`.

## Version and module

- Go 1.26.x is this template's target version.
- Native Go Modules (`go.mod`/`go.sum`); no vendoring.
- A local `replace` for your own not-yet-published libraries is acceptable (e.g., `replace github.com/raykavin/gocnab => ./pkg/gocnab`) — but prefer publishing as a versioned module (semver tag) once it stabilizes.

## gobox — platform library, check before reimplementing

Before writing any of these, check the corresponding subpackage in `github.com/raykavin/gobox`:

| Concern | gobox subpackage | Note |
|---|---|---|
| Logging | `logger` | zerolog underneath, logrus-like interface |
| Config | `config` | generic `Loader[T]` over Viper, hot-reload via fsnotify |
| HTTP server | `httpserver` | wraps Gin; `respond/` = response envelope; `middlewares/` = auth/CORS/CSRF/oidcauth |
| OIDC authentication | `oidcauth` | local JWT verification + remote introspection, pluggable cache |
| Pagination | `pagination` | `Page`/`Result[T]`/`Query`/`Filter` — `Scope()` is specific to `*gorm.DB` |
| Database (GORM) | `database/gorm` | connection factory with retry and configurable pool — does **not** define `BaseModel`, soft delete, or transactions; that's left to the consuming project |
| Migrations | `database/migrate` | wrapper over `golang-migrate`; only exposes `Migrate()` (Up) and `Populate()` (seeds) |
| Generic SQL | `database/sql` | `Connector[T]` over plain `database/sql`, no extra pooling, no transaction |
| Healthcheck | `healthcheck` | concurrent probe via the `Pinger` interface |
| Telemetry | `telemetry` | OpenTelemetry bootstrap (OTLP traces + Prometheus metrics) |
| Retry | `retry` | pure function `Do(ctx, maxAttempts, backoff, shouldRetry, fn)` |
| HTTP client | `httpclient` | thin wrapper over `net/http` + decompression |
| OAuth2 client-credentials | `oauth2` | `TokenManager` with in-memory token cache |

**Contribution rule**: if a cross-cutting concern is being reimplemented locally instead of importing `gobox` (e.g., HTTP response envelope, OIDC verifier, logging wrapper), prefer migrating to the shared lib as part of the task — **don't decide alone to do a broad, unrequested migration**; record the decision if the migration is bigger than the requested scope.

## Code conventions

- File name: `snake_case.go`. One file per operation in `usecase/` (`create.go`, `update.go`, `get_by_id.go`, `find_or_create.go`, `filter.go`).
- Packages: short, lowercase name, no underscore, one concept per package.
- Errors: local sentinels per package (`var ErrNotFound = errors.New("...")`), never a central errors/exceptions package. Wrap with `fmt.Errorf("%w: %w", ErrX, cause)` (double-wrap, Go 1.20+) or `errors.Join`. In the HTTP handler, map explicitly via `errors.Is`/`errors.As` — never let it fall through to 500 by default.
- Generics: used deliberately for reusable infrastructure types (`pagination.Result[T]`, `database/sql.Connector[T]`, `config.Loader[T]`) — don't force generics into domain/business code where there's no real reuse.
- `context.Context` as the first parameter in every I/O/network operation.
- Input validation: Gin's `binding:"required"` tags (which use `go-playground/validator` underneath). Never use the `validate:"..."` tag alone outside Gin's bind flow.
- Comments: only where they explain a non-obvious decision (e.g., why a column has no FK, why a specific race condition was fixed a certain way). Don't document the obvious.

## Test style

- Framework: the stdlib `testing` is the standard. `testify` (`assert`/`require`/`mock`) is acceptable when already in use in the project — don't introduce it in a project that currently only uses stdlib without a concrete need.
- Naming: `TestXxx_Scenario`, with subtests via `t.Run(...)` for variations. Table-driven when there are multiple input/output cases.
- See `instructions/testing.md` for commands, mocks, and coverage.

## Linting

- `.golangci.yml` with `disable-all: true`, enabling `govet, errcheck, staticcheck, ineffassign, unused, gofmt, goimports`, 3m timeout. Adopt exactly this set in new projects or when adding lint to an existing one.

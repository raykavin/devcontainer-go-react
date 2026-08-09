# Specific instructions — Tests and quality

Complements `AGENTS.md`.

## Minimum validation command

Run in this order before considering any task complete:

```sh
cd backend
go build ./...
go vet ./...
golangci-lint run          # only if .golangci.yml exists in the repository
go test -v -race ./...
```

- `go test -v -race ./...` is the non-negotiable minimum bar, always with `-race`.
- Run `go build`/`go vet` as isolated steps even if they seem redundant with the test — they're cheap and catch problems the test alone won't.
- If the repository has no `.golangci.yml`, don't invent a different linter set than the one recommended in `instructions/go.md` — either skip this step, or (if the task includes "improve quality") add exactly that same set.
- If any step can't run in your environment (tool not installed, no network access to fetch a dependency), state that explicitly in the task's final report — never omit it or pretend it passed.

## Assertion framework

- The stdlib `testing` is the standard. `testify` (`assert`/`require`/`mock`) is acceptable when already used in the project. Don't introduce `testify` in a project that currently only uses stdlib, unless the task explicitly asks for it.

## Naming and style

- `TestXxx_Scenario` with subtests via `t.Run("case description", func(t *testing.T) {...})`.
- Table-driven for multiple input/output cases (a `mutate func(*Input)` per case is a useful pattern).
- Name the test after the behavior, not the implementation — the test name should document the condition it exists to prevent.

## Mocks

Prefer one of these three patterns, in this order of preference for most cases:
1. **Hand-written functional structs** — `type mockSubRepo struct { SearchFn func(...) ...; GetByIDFn func(...) ... }`, implementing the `port` interface with substitutable function fields per test. This is the simplest to maintain and the default recommendation.
2. **`testify/mock`** — when `testify` is already used in the project.
3. **Real SQLite in-memory** — to test a real transaction/GORM instead of mocking every collaborator. Prefer this option when the test needs to validate a real multi-table transaction.

Don't rely on `make back-mocks`/`mockgen` without first confirming it actually generates useful artifacts in the project — it's common for this Makefile target to exist by copy from another project without ever having been used for real.

## Coverage

Don't add an arbitrary minimum coverage gate unless the task explicitly asks for it.

## Required tests by change type

- **New HTTP endpoint**: authorization test (no permission → 403/401), success, invalid input (400), non-existent resource (404) when applicable.
- **New domain rule/usecase**: unit test covering the happy path and at least one relevant business-error condition.
- **New migration/schema change**: if the project uses versioned SQL, test that the migration applies without error against a clean database (via `cmd/migrator` or an equivalent script); if it uses `AutoMigrate`, verify that `AutoMigrate` doesn't fail against the current schema.
- **Concurrent code** (queues, workers, workflows): always write a test, even if the rest of the file/package has low coverage — concurrent code without tests is where subtle bugs most often go unnoticed.

## CI/CD

New projects should include a basic pipeline (lint + build + test) from the start — see `new-project-checklist.md`. For existing repositories without CI, adding it is a recommended improvement, not an action to take without the task explicitly asking for it (changes to the build/deploy pipeline require approval — see `AGENTS.md`).

# Feature implementation template

Recommended flow for implementing a feature in this backend, from understanding the requirement to final validation.

## 1. Understand the expected behavior

- Re-read the task's requirement. If something is ambiguous in a way that would change the implementation (e.g., HTTP verb, response format, whether a transaction is needed), ask before writing code — don't assume.
- Identify which layer the change actually belongs to: `domain` (rule/invariant), `usecase` (orchestration), `adapter/inbound/http` (API contract), `adapter/outbound` (external integration/database).

## 2. Identify affected contracts

- Does the change alter an existing HTTP route (request/response format, status code, verb)? If so, it's a potentially incompatible change — see "Operations that require approval" in `AGENTS.md` before proceeding.
- Does the change alter the database schema (new column, new index, new relationship)? Check whether the project uses `AutoMigrate` or versioned migrations (`instructions/database.md`) and follow the mechanism already in use.
- Does the change affect a contract consumed by another service (e.g., webhook payload, format accepted by an external integration)? Treat it as a cross-service change — it requires more caution and possibly approval.

## 3. Locate existing examples

Before writing any new code, look for a case already implemented in the same repository:
- New listing endpoint with a filter → look at an existing `Search`/`List` in `internal/usecase/` + `internal/adapter/outbound/postgres/`.
- New idempotent ingestion operation (upload, import) → check whether a content-hash find-or-create pattern already exists in the project.
- New multi-step usecase that writes to more than one table → check whether a `TransactionRunner`/Unit-of-Work already exists.
- New authenticated endpoint → see how `gobox/oidcauth`+`middlewares` are already used in the same repository before touching authentication.
- Before creating a new abstraction (generic helper, wrapper, HTTP client): check whether it already exists in `github.com/raykavin/gobox` (see the table in `instructions/go.md`).
- Before adding a new dependency to `go.mod`: check whether `gobox` or an existing dependency already solves the same problem.

## 4. Define domain rules

- Business rules/invariants live in `internal/domain` (or as an entity method). Don't put business rules in a `handler` or in `adapter/outbound`.
- If the rule depends on external state (another service, config), that's `usecase` orchestration, not a pure `domain` rule.

## 5. Evaluate security and authorization

- Every new route needs: an authentication middleware (`gobox/oidcauth`) + an authorization middleware with a permission specific to the operation (don't reuse a generic permission from another endpoint).
- If the operation processes sensitive data (national ID, banking data, PII), confirm it won't be logged without redaction, especially at debug level.
- See the full checklist in `instructions/security.md`.

## 6. Evaluate persistence and migrations

- New entity/column: follow the schema mechanism already used in the repository (`AutoMigrate` or a versioned SQL migration — `instructions/database.md`).
- New multi-table write operation: use (or create, if it doesn't yet exist in the repository) a `TransactionRunner`.
- New query with a relationship: use `Preload`, not manual `Joins` (reserve `Joins` for filtering).
- If the change touches a migration already applied in production: **don't edit it** — create a new one.

## 7. Implement the smallest coherent change

- One usecase per operation (`internal/usecase/<context>/<operation>.go`).
- Don't refactor unrelated code "while you're in there" — if you spot an improvement opportunity outside the scope, report it at the end instead of executing it without asking.
- Reuse the response envelope, pagination, logger, and authentication already in use in the repository.

## 8. Create or update tests

- New endpoint: test for authorization, success, invalid input, non-existent resource.
- New usecase: happy path + at least one business-error condition.
- Concurrent code (queue, worker, workflow): always test, even if it looks simple.
- Use a hand-written functional mock or SQLite in-memory for a real transaction (see `instructions/testing.md`).

## 9. Run validations

```sh
go build ./...
go vet ./...
golangci-lint run     # if .golangci.yml exists in the repository
go test -v -race ./...
```
If any step can't run, state that explicitly in the final report.

## 10. Update documentation

- Swagger annotations (`@Summary/@Router/@Success/@Failure`) in sync with the actual behavior — regenerate via `make back-swagger`.
- README (relevant section: Architecture/Configuration/API) if the documented behavior changed.
- The repository's own `.claude/`/`AGENTS.md`, if it exists and the change affects what it describes.

## 11. Report

In the task's final report, include: what changed (files/routes/schema), why (link to the requirement), which decisions you made and why (e.g., which HTTP verb you chose and why, whether you followed `AutoMigrate` or a versioned migration), what risks remain, and how you validated it (commands run and result — including failures or skipped steps). If anything touched one of the operations that require approval (`AGENTS.md`), confirm that approval was obtained before executing, not after.

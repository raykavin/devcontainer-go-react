# Specific instructions — APIs and contracts

Complements `AGENTS.md`.

## Framework and routes

- Gin is the standard HTTP framework. Plain `net/http` is acceptable only for simple, isolated cases (e.g., a webhook receiver with a single route).
- Fixed route prefix: `/api/v1`. Don't create `/v2` without first recording a versioning decision — there's no defined deprecation strategy by default.
- Resources in plural kebab-case: `bank-accounts`, `document-types`.
- Read by ID: `GET /resource/:id`.
- Search/listing with a filter: `POST /resource/search` with the filter in the request body — **don't** use `GET` with a querystring for composite filters; the filter integrates directly with `gobox/pagination.NewFilterBuilder()`.
- Update: `PATCH` by default; use `PUT` only when the entire resource is replaced. When editing an existing endpoint, keep the verb already in use.

## Response envelope

`gobox/httpserver/respond` defines the canonical format:
```go
type Response struct {
    Success bool          `json:"success"`
    Message string        `json:"message,omitempty"`
    Data    any           `json:"data,omitempty"`
    Errors  []*APIError   `json:"errors,omitempty"`
}
```
with helpers `OK/Created/Accepted/BadRequest/NotFound/InternalServerError/...`.

Import `gobox/httpserver/respond` directly — don't copy the code into the local repository. Swapping an existing project's response envelope is an incompatible contract change and requires human approval.

## Domain error → HTTP status

When creating/editing a handler: explicitly map each relevant domain error to the correct HTTP status via `errors.Is`/`errors.As`. Don't let every error fall through to 500 by default. If you find a Swagger annotation that doesn't match the code's actual behavior, fix one of the two sides (preferably the code, if the documented contract is what the consumer expects) and flag the inconsistency.

## Authentication and authorization

- Authentication: always via `gobox/oidcauth` (Bearer/cookie OIDC against the configured OIDC provider). **Never** implement JWT verification/introspection from scratch.
- Human authorization: declarative RBAC per route, permission named `"resource:action"` (e.g., `"invoice:validate"`, `"company:read"`). The check always happens in the middleware/router, never inside the usecase.
- M2M authorization (service-to-service, webhooks): OAuth scope (`RequireScope`) + `azp` (authorized party) verification when the client is known.

## Webhooks

Two authentication mechanisms are valid, depending on the source system's capability:
- **Full M2M OIDC**: client_credentials Bearer, scope + `azp` required. Use when the source system speaks OIDC natively.
- **Static token**: fixed token in the query string, compared with `subtle.ConstantTimeCompare` over SHA-256, **and redacted from the access log before writing**. Use only when the source system doesn't offer OIDC — always implement constant-time comparison and log redaction.

## Input validation

- `go-playground/validator` via Gin's `binding:"required"` tags (`ctx.ShouldBindJSON`).
- Handle the bind error explicitly. Don't let the validator's raw message (English, format `"Key: '...' Error:Field validation..."`) leak into the response's `details`/`error` field — parse it field-by-field.

## Pagination

Always via `github.com/raykavin/gobox/pagination`:
```go
filters := pagination.NewFilterBuilder().WhereIf(cond, "column", pagination.Eq, value).Build()
sorts := pagination.NewSortBuilder().OrderBy("created_at", pagination.Desc).Build()
query := pagination.NewQuery(page, filters, sorts)
db.Scopes(pagination.Scope(query, &total)).Find(&results)
result := pagination.NewResult(results, int(total), query.Page)
```
Pagination input always through the request **body** (inside the filter of `POST /resource/search`), never through a `?page=` querystring. Note: `pagination.Scope` is specific to `*gorm.DB` — if the project uses plain pgx, implement manual `LIMIT`/`OFFSET` from `Query.Page.Offset()`/`PerPage`, replicating the same field names for consistency.

## Ingestion idempotency

For file upload, spreadsheet import, or async processing of a return/callback: SHA-256 hash of the content + composite unique index `(parent_scope_id, file_hash)` + find-or-create flow in the usecase. Explicitly document the accepted race window (concurrent double-upload of the same file).

## OpenAPI documentation

`swaggo/swag`, generated via `make back-swagger` from `@Summary/@Router/@Success/@Failure` annotations on handlers. Every task that changes an endpoint's behavior must regenerate the documentation and verify the annotations match the actual code.

## Changes that require human approval

Any change that breaks backward compatibility of an existing endpoint: removing/renaming a response field, changing the response envelope, changing the HTTP verb, changing the pagination format, removing a required permission. See "Operations that require approval" in `AGENTS.md`.

# Specific instructions — Security

Complements `AGENTS.md`. No real secret value should ever be reproduced in any document or artifact in this repository — only mechanisms.

## Authentication

- **Always** via `github.com/raykavin/gobox/oidcauth` against the configured OIDC provider (Keycloak or equivalent). Never implement JWT verification/introspection from scratch — this tends to create divergent forks that don't receive security fixes made in `gobox`.
- A homegrown JWT + bcrypt + local RBAC (e.g., Casbin) is an acceptable exception only when the service doesn't depend on corporate/federated SSO (standalone product, single-org RBAC). Don't migrate to OIDC without confirming the deployment context changed.

## Authorization

- Declarative RBAC per route, permission named `"resource:action"`, or OAuth scope for M2M. The check always happens in the middleware/router — never inside the usecase.
- Avoid leaking the authorization decision into the data layer via `ctx.Value` beyond what's needed for auditing — this mixes declarative authorization with downstream imperative decisions and makes it harder to audit "who can do what" by just looking at the router.

## CORS

- Always via `gobox/httpserver/middlewares` (current version, restrictive by origin allowlist). Never copy/reimplement CORS locally — a permissive fallback (`Access-Control-Allow-Origin: *` combined with `Access-Control-Allow-Credentials: true`) is a common mistake in local implementations and shouldn't exist in new code.

## CSRF

- Required on any route that accepts a session cookie: HttpOnly cookie + CSRF middleware via `gobox/httpserver/middlewares`.
- If a route reads an OIDC session cookie as an authentication fallback but has no CSRF middleware, that's a security gap — flag it and treat it as a change that requires explicit approval (see `AGENTS.md`).

## Secrets management

- Standard: a versioned `configs/*.example.yml` (with placeholders like `"changeme"`), the real file (`configs/*.yml` without the `.example` suffix) always in `.gitignore`.
- **Never** commit a configuration file with a real value, not even "temporarily", not even on a feature branch.
- If you find any file that looks like it contains a real credential/token during a task (even if untracked, e.g. in `temp/`), **don't reproduce the value anywhere** (code, log, response, commit) — just report the location to the person responsible.

## Webhooks

- Static token: always compare with `subtle.ConstantTimeCompare` over a hash (never string `==`), and **redact the token from any access log/URL before writing it**.
- M2M OIDC: require a known scope + `azp` (authorized party), not just a generic valid Bearer token.

## Rate limiting

Not implemented by default in this template — it's a recommendation for projects that expose login and webhook endpoints publicly, not an existing guarantee.

## Sensitive data in logs

- There's no automatic PII masking in `gobox/logger` — each project needs to handle this on a case-by-case basis.
- If a task involves logging a payload containing a national ID, email, phone number, or banking data, explicitly redact those fields before logging, especially at `debug`/`info` level — don't assume automatic protection exists.

## Environment separation

Segregate by a different config file per process (`api.yml` vs `api.dev.yml` vs `api.prod.yml`), not by an `APP_ENV` variable read at runtime. Handle `.gitignore`/deploy scripts with care — an accidentally versioned real config file is the most common leak vector.

## Security checklist for any new endpoint

1. Authentication middleware (`gobox/oidcauth`) applied to the correct route group.
2. Authorization middleware with a permission/scope specific to the operation, not a reused generic permission.
3. CORS via `gobox/httpserver/middlewares`, not reimplemented.
4. CSRF if the route depends on a session cookie.
5. Input validation via binding+validator, without leaking the validator's raw message to the client.
6. No sensitive data logged without redaction, even at debug level.
7. If the endpoint is a webhook: constant-time comparison and token redaction in logs.

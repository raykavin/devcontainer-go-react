# AGENTS.md

Guidance for AI coding agents (Claude Code, etc.) working in this repository.

## Repository layout

```
backend/       Go application (module name = $PROJECT_NAME from devcontainer.json) + production Dockerfile
frontend/      React application (Vite; package manager = $PKG_MANAGER from devcontainer.json) + production Dockerfile
Makefile       Entry point for all common tasks
```

Both toolchains (Go and Node.js) are available in this single container. You can build, run, test, and lint both sides from the same shell.

## Commands

| Task                    | Command                                      |
| ----------------------- | -------------------------------------------- |
| Run backend             | `make back-run` (serves on :8080)            |
| Run frontend dev server | `make front-dev` (Vite on :5173)             |
| Run both                | `make dev`                                   |
| All tests               | `make test`                                  |
| All linters             | `make lint`                                  |
| Backend tests only      | `make back-test`                             |
| Frontend build          | `make front-build`                           |
| Production images       | `make back-build` / `make front-build-image` |

## Conventions

- Go code follows idiomatic Go: minimal abstraction, compile-time safety, no framework magic.
- All code comments in English.
- Backend lives entirely under `backend/`; never place Go files at the repository root.
- Frontend is TypeScript + TSX; run `$PKG_MANAGER run typecheck` inside `frontend/` to type-check via `tsconfig.json`.
- Lint before finishing any task: `make lint`.
- `gopls` ignores `frontend/` (see `.devcontainer/devcontainer.json`), so Go tooling stays fast.
- Production Dockerfiles live directly in `backend/Dockerfile` and `frontend/Dockerfile`, matching their own build context — files referenced by `COPY` must exist inside those folders.

## Caches (do not delete)

- `/go/bin` and `/go/pkg` are persistent volumes (Go tools and modules).
- `frontend/node_modules` is a persistent volume; do not remove it wholesale use the package manager.
- `~/.claude` is a persistent volume (Claude Code credentials and settings).

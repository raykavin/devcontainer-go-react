# Full Stack Devcontainer Template (Go + React)

[![Go Devcontainer](https://img.shields.io/badge/Go%20+%20React%20Devcontainer-blue?style=flat&logo=go)](https://github.com/raykavin/devcontainer-full)

Single-container full stack template (Go + React) optimized for AI coding agents interactive setup and persistent caches.

Merges [devcontainer-go](https://github.com/raykavin/devcontainer-go) and a React (Vite) frontend setup.

## Purpose

Provide one consistent environment where an AI agent (or you) can build, run, test, and lint both the backend and the frontend from the same shell no container hopping, no split terminals, no per-machine setup.

## Project Layout

```
.devcontainer/   Development container: compose, Dockerfile, post-create
backend/         Go module (API on :8080) + auxiliary scripts (scripts/sh)
frontend/        React + TypeScript (Vite on :5173)
deploy/          Production: Dockerfiles (docker/) and docker-compose.yml
Makefile         Task entry point
AGENTS.md        Instructions for AI coding agents
init.sh          Interactive configurator
```

The Vite dev server proxies `/api` to the Go backend on `:8080` (see `frontend/vite.config.ts`), so the frontend calls the API without CORS setup during development.

## How to Use

### 1. Clone

```sh
git clone https://github.com/raykavin/devcontainer-full.git my-project
cd my-project
```

### 2. Configure

```sh
./init.sh
```

Asks each option with a sensible default (press Enter to accept):

```
Project name [my-app]:
Go version (1.26/1.25) [1.26]:
Node.js version (24/22/20) [24]:
Package manager (pnpm/npm) [pnpm]:
Timezone [America/Belem]:
Git user.name / user.email / origin URL / initial branch
```

At the end it offers to delete itself and open VS Code directly inside the dev container (`code --folder-uri vscode-remote://dev-container+...`), so the build starts without the "Reopen in Container" click. Answer `n` to keep the script; it is idempotent and can be re-run at any time (then use "Dev Containers: Rebuild Container").

All values live in `.devcontainer/devcontainer.json` (`containerEnv`) and `.devcontainer/docker-compose.yml` (build args), so editing by hand also works.

### 3. First start

`post-create.sh` runs automatically and:

1. Initializes (or validates) the Git repository and configures `user.name` / `user.email` / `origin`
2. Sets up the private Go proxy if `GO_PROXY_HOST` / `GO_PROXY_PORT` are set
3. Runs `go mod init $PROJECT_NAME` in `backend/` (if `go.mod` is absent) and installs Go tools (`gopls`, `goimports`, `dlv`, `staticcheck`, `golangci-lint`)
4. Installs frontend dependencies with the chosen package manager
5. Sets fish as the default shell

### Reuse in an existing project

Copy `.devcontainer/`, `deploy/`, `Makefile`, and `AGENTS.md` into any project that follows the `backend/` + `frontend/` layout. The post-create script skips every step whose result already exists (`.git`, `go.mod`).

## Development vs Production

|               | Development                                                               | Production                                                        |
| ------------- | ------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Compose       | `.devcontainer/docker-compose.yml`                                        | `deploy/docker-compose.yml`                                       |
| Backend image | `mcr.microsoft.com/devcontainers/go` (prebuilt, MCR)                      | `deploy/docker/backend.Dockerfile` (multi-stage Alpine, non-root) |
| Frontend      | Vite dev server on :5173                                                  | `deploy/docker/frontend.Dockerfile` (nginx, non-root)             |
| Node.js       | Copied from the official `node` image on Docker Hub (multi-stage, no nvm) | Official `node` image (build stage only)                          |

The production Dockerfiles live in `deploy/docker/` but keep `backend/` and `frontend/` as build contexts, so `COPY` paths inside them are unchanged:

```sh
docker build -f deploy/docker/backend.Dockerfile  -t myapp-api:1.0.0 backend/
docker build -f deploy/docker/frontend.Dockerfile -t myapp-web:1.0.0 frontend/
# or: make back-build / make front-build-image
# or: docker compose -f deploy/docker-compose.yml up -d --build
```

## Makefile Targets

```
make back-run           # Run the Go backend
make front-dev          # Start the Vite dev server
make dev                # Run both in parallel
make test               # Backend + frontend tests
make lint               # golangci-lint + frontend lint
make back-build         # Production backend image
make front-build        # Production frontend build (dist/)
make front-build-image  # Production frontend image
```

## Persistent Caches

Named volumes survive container rebuilds:

```
go_bin        →  /go/bin
go_pkg        →  /go/pkg
node_modules  →  /workspaces/app/frontend/node_modules
pnpm_store    →  ~/.local/share/pnpm/store
claude_code   →  ~/.claude   (Claude Code credentials, settings, history — if installed manually)
```

## AI Agents

- `AGENTS.md` at the root documents every command and convention Claude Code and similar tools read it automatically
- The Claude Code CLI is **not** installed automatically; install it manually inside the container if you use it (`npm install -g @anthropic-ai/claude-code`) — its state persists across rebuilds via the `claude_code` volume
- `gopls` is configured to ignore `frontend/`, so Go tooling stays fast in the monorepo

## Notes

- This is a personal template, meant to evolve over time
- Default shell inside the container is fish
- Go versions offered by `init.sh` are the ones with `-trixie` tags on `mcr.microsoft.com/devcontainers/go`

---

Built to make full stack development with AI agents easier.

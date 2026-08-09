#!/bin/sh
set -e

log() { echo "> $*"; }
info() { echo "  $*"; }

WORKSPACE="/workspaces/app"
BACKEND_DIR="$WORKSPACE/backend"
FRONTEND_DIR="$WORKSPACE/frontend"

log "Configuring Git"
cd "$WORKSPACE"

if [ ! -d ".git" ]; then
  info "Initializing repository (branch: $GIT_INIT_DEFAULT_BRANCH)"
  git init --initial-branch="$GIT_INIT_DEFAULT_BRANCH"
else
  info "Repository already exists, skipping init"
fi

git config --global --add safe.directory "$WORKSPACE"
git config --global user.name  "$GIT_CONFIG_DEV_USERNAME"
git config --global user.email "$GIT_CONFIG_DEV_EMAIL"

if git remote | grep -q "^origin$"; then
  CURRENT_URL="$(git remote get-url origin)"
  if [ "$CURRENT_URL" != "$GIT_REPO_ADDRESS" ]; then
    info "Updating origin URL"
    git remote set-url origin "$GIT_REPO_ADDRESS"
  else
    info "Origin already configured correctly"
  fi
else
  info "Adding origin remote"
  git remote add origin "$GIT_REPO_ADDRESS"
fi

if git show-ref --verify --quiet "refs/heads/$GIT_INIT_DEFAULT_BRANCH"; then
  git branch --set-upstream-to="origin/$GIT_INIT_DEFAULT_BRANCH" "$GIT_INIT_DEFAULT_BRANCH" 2>/dev/null || true
fi

if [ -n "$GO_PROXY_HOST" ] && [ -n "$GO_PROXY_PORT" ] && [ -n "$GIT_CONFIG_DEV_USERNAME" ] && [ -n "$GIT_CONFIG_TOKEN" ]; then
  log "Go private proxy env vars detected, running proxy-setup.sh"
  sh "$WORKSPACE/.devcontainer/proxy-setup.sh"
else
  info "Go private proxy env vars not set, skipping proxy-setup.sh"
fi

log "Setting up backend (Go)"
cd "$BACKEND_DIR"

# /go/bin and /go/pkg are named volumes and are owned by root on first creation
log "Adjusting Go directory permissions"
sudo chmod -R 777 /go/bin /go/pkg

export PATH="$PATH:/go/bin"

if [ ! -f go.mod ]; then
  info "Creating module: $PROJECT_NAME"
  go mod init "$PROJECT_NAME"
else
  info "go.mod already exists, skipping"
fi

go mod tidy

if [ -n "$GOPRIVATE" ]; then
  log "Setting private Go repositories"
  go env -w GOPRIVATE="$GOPRIVATE"
fi

log "Installing Go development tools"
go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

log "Verifying installed Go tools"
gopls version
dlv version
golangci-lint version

log "Running initial backend lint"
golangci-lint run || true

log "Setting up frontend (React, package manager: $PKG_MANAGER)"

# node_modules and the Claude Code state are named volumes, owned by root on first creation
sudo chown -R vscode:vscode "$FRONTEND_DIR/node_modules" /home/vscode/.local/share/pnpm /home/vscode/.claude 2>/dev/null || true

if [ "$PKG_MANAGER" = "pnpm" ]; then
  if ! command -v pnpm > /dev/null 2>&1; then
    info "Enabling pnpm via corepack"
    sudo corepack enable
    corepack prepare pnpm@latest --activate
  fi
fi

cd "$FRONTEND_DIR"

if [ ! -f package.json ]; then
  info "No package.json found in frontend/, skipping frontend setup"
else
  log "Installing frontend dependencies"
  "$PKG_MANAGER" install

  log "Verifying installed Node tools"
  node --version
  "$PKG_MANAGER" --version
  npx tsc --version

  log "Running initial frontend lint"
  "$PKG_MANAGER" run lint || true
fi

# Safe to change the login shell now: everything that runs `su vscode -c` already finished
log "Setting fish as default shell for vscode user"
sudo chsh -s /usr/bin/fish vscode || true

cd "$WORKSPACE"
log "Checking out develop branch"
if git show-ref --verify --quiet "refs/heads/develop"; then
  if git rev-parse --abbrev-ref HEAD | grep -q "^develop$"; then
    info "Already on develop branch"
  else
    info "Switching to develop branch"
    git checkout develop
  fi
fi

echo ""
log "Post-create completed successfully"
info "Backend:  make back-run   (http://localhost:8080)"
info "Frontend: make front-dev  (http://localhost:5173)"

#!/usr/bin/env bash
set -euo pipefail

SWAG_VERSION="${SWAG_VERSION:-v1.16.4}"
MAIN_FILE="${MAIN_FILE:-cmd/api/main.go}"
DOCS_OUTPUT_DIR="${DOCS_OUTPUT_DIR:-./internal/adapter/inbound/http/docs}"

# walk up from the script's location to the nearest go.mod
find_module_root() {
  local dir
  dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  while [[ "$dir" != "/" ]]; do
    [[ -f "$dir/go.mod" ]] && { echo "$dir"; return 0; }
    dir="$(dirname "$dir")"
  done
  return 1
}

MODULE_ROOT="$(find_module_root)" || {
  echo "error: no go.mod found above $(dirname "${BASH_SOURCE[0]}")" >&2
  exit 1
}
cd "$MODULE_ROOT"

GOBIN="$(go env GOBIN)"
[[ -n "$GOBIN" ]] || GOBIN="$(go env GOPATH)/bin"
export PATH="$GOBIN:$PATH"

if ! command -v swag >/dev/null 2>&1; then
  echo "Installing swag CLI (${SWAG_VERSION})..."
  go install "github.com/swaggo/swag/cmd/swag@${SWAG_VERSION}"
fi

echo "Generating Swagger documentation (module: ${MODULE_ROOT})..."
rm -rf "$DOCS_OUTPUT_DIR"
swag init \
  --parseDependency \
  --parseInternal \
  -g "$MAIN_FILE" \
  -o "$DOCS_OUTPUT_DIR"

echo "Done — output: ${DOCS_OUTPUT_DIR}"
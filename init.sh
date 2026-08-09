#!/bin/sh
# ============================================================================
# Interactive devcontainer configurator
#
# Run this BEFORE opening the folder in VS Code. It asks a few questions and
# rewrites the values inside .devcontainer/devcontainer.json and
# .devcontainer/docker-compose.yml. Safe to re-run at any time.
#
# Usage:  ./init.sh
# ============================================================================
set -e

DEVCONTAINER_JSON=".devcontainer/devcontainer.json"
COMPOSE_FILE=".devcontainer/docker-compose.yml"

[ -f "$DEVCONTAINER_JSON" ] || { echo "Run this script from the repository root."; exit 1; }

# --- helpers ----------------------------------------------------------------
ask() {
  # ask "Question" "default"
  printf "%s [%s]: " "$1" "$2" >&2
  read -r answer
  [ -z "$answer" ] && answer="$2"
  echo "$answer"
}

ask_choice() {
  # ask_choice "Question" "opt1 opt2 ..." "default"
  while true; do
    v=$(ask "$1 ($(echo "$2" | tr ' ' '/'))" "$3")
    for opt in $2; do
      [ "$v" = "$opt" ] && echo "$v" && return
    done
    echo "  Invalid option: $v" >&2
  done
}

json_set() {
  # json_set KEY VALUE  -> rewrites "KEY": "..." inside devcontainer.json
  sed -i "s|\"$1\": \"[^\"]*\"|\"$1\": \"$2\"|" "$DEVCONTAINER_JSON"
}

# --- current values as defaults ----------------------------------------------
cur() { grep -o "\"$1\": \"[^\"]*\"" "$DEVCONTAINER_JSON" | head -1 | sed 's/.*: "\(.*\)"/\1/'; }

echo "=== Full Stack Go + React devcontainer setup ==="
echo ""

PROJECT_NAME=$(ask  "Project name"                     "$(cur PROJECT_NAME)")
GO_VERSION=$(ask_choice "Go version"                   "1.26 1.25"    "$(grep -o 'GO_VERSION: "[^"]*"' "$COMPOSE_FILE" | sed 's/.*"\(.*\)"/\1/')")
NODE_VERSION=$(ask_choice "Node.js version"            "24 22 20"     "$(grep -o 'NODE_VERSION: "[^"]*"' "$COMPOSE_FILE" | sed 's/.*"\(.*\)"/\1/')")
PKG_MANAGER=$(ask_choice "Package manager"             "pnpm npm"     "$(cur PKG_MANAGER)")
TIMEZONE=$(ask     "Timezone"                          "$(cur TZ)")
echo ""
GIT_USERNAME=$(ask "Git user.name"                     "$(cur GIT_CONFIG_DEV_USERNAME)")
GIT_EMAIL=$(ask    "Git user.email"                    "$(cur GIT_CONFIG_DEV_EMAIL)")
GIT_REPO=$(ask     "Git remote (origin) URL"           "$(cur GIT_REPO_ADDRESS)")
GIT_BRANCH=$(ask   "Initial branch name"               "$(cur GIT_INIT_DEFAULT_BRANCH)")

# --- devcontainer.json --------------------------------------------------------
echo ""
echo "> Updating $DEVCONTAINER_JSON"
sed -i "s|\"name\": \"[^\"]*\"|\"name\": \"$PROJECT_NAME DevContainer (Go + React)\"|" "$DEVCONTAINER_JSON"
json_set PROJECT_NAME             "$PROJECT_NAME"
json_set TZ                       "$TIMEZONE"
json_set PKG_MANAGER              "$PKG_MANAGER"
json_set GIT_CONFIG_DEV_USERNAME  "$GIT_USERNAME"
json_set GIT_CONFIG_DEV_EMAIL     "$GIT_EMAIL"
json_set GIT_REPO_ADDRESS         "$GIT_REPO"
json_set GIT_INIT_DEFAULT_BRANCH  "$GIT_BRANCH"

# --- docker-compose.yml -------------------------------------------------------
echo "> Updating $COMPOSE_FILE"
sed -i "s|container_name: .*|container_name: dev_$PROJECT_NAME|" "$COMPOSE_FILE"
sed -i "s|GO_VERSION: \"[^\"]*\"|GO_VERSION: \"$GO_VERSION\"|"   "$COMPOSE_FILE"
sed -i "s|NODE_VERSION: \"[^\"]*\"|NODE_VERSION: \"$NODE_VERSION\"|" "$COMPOSE_FILE"
sed -i "s|TZ: \"[^\"]*\"|TZ: \"$TIMEZONE\"|"                     "$COMPOSE_FILE"

echo ""
echo "> Done. Summary:"
echo "    Project:      $PROJECT_NAME"
echo "    Go:           $GO_VERSION"
echo "    Node:         $NODE_VERSION ($PKG_MANAGER)"
echo "    Timezone:     $TIMEZONE"
echo "    Git:          $GIT_USERNAME <$GIT_EMAIL> -> $GIT_REPO ($GIT_BRANCH)"
echo ""

# --- open in VS Code + self-delete -------------------------------------------
OPEN_NOW=$(ask_choice "Delete init.sh and open VS Code in the dev container now?" "y n" "y")

if [ "$OPEN_NOW" = "y" ]; then
  # Encode the absolute folder path as lowercase hex, as required by the
  # vscode-remote://dev-container+<hex> URI scheme
  FOLDER="$(pwd)"
  HEX=$(printf '%s' "$FOLDER" | od -A n -t x1 | tr -d ' \n')

  rm -- "$0"
  echo "> init.sh removed"

  if command -v code > /dev/null 2>&1; then
    echo "> Opening VS Code directly in the dev container (build starts automatically)"
    exec code --folder-uri "vscode-remote://dev-container+${HEX}/workspaces/app"
  else
    echo "'code' CLI not found. Open the folder manually and choose 'Reopen in Container'."
  fi
else
  echo "Open the folder in VS Code and choose 'Reopen in Container'."
  echo "If the container was already built, run 'Dev Containers: Rebuild Container'."
fi

#!/bin/sh
# Runs on every container start (and once from post-create.sh).
#
# Named volumes are created owned by root when their mount point does not exist
# in the image. Fix only what is wrong, so a healthy start costs a few `stat`
# calls instead of a recursive chown of the pnpm store.

for dir in \
  /home/vscode/.claude \
  /home/vscode/.local/share/pnpm/store \
  /workspaces/frontend/node_modules; do
  [ -d "$dir" ] || continue
  if [ "$(stat -c %U "$dir")" != "vscode" ]; then
    echo "> Fixing ownership of $dir"
    sudo chown -R vscode:vscode "$dir"
  fi
done

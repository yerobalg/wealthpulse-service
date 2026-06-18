#!/bin/bash
#
# Rename the Go module path across the whole project.
#
# Reads the current module path from go.mod and replaces every occurrence with
# the new one in .go, .mod, .md, and Makefile files.
#
# Usage:
#   ./scripts/rename_module.sh github.com/<your-org>/<your-repo>
#
set -e

NEW="$1"

if [ -z "$NEW" ]; then
  echo "usage: $0 <new-module-path>" >&2
  echo "example: $0 github.com/acme/my-service" >&2
  exit 1
fi

if [ ! -f go.mod ]; then
  echo "error: go.mod not found — run this from the project root" >&2
  exit 1
fi

OLD=$(grep '^module ' go.mod | awk '{print $2}')

if [ -z "$OLD" ]; then
  echo "error: could not read current module path from go.mod" >&2
  exit 1
fi

if [ "$OLD" = "$NEW" ]; then
  echo "module path is already '$NEW' — nothing to do"
  exit 0
fi

echo "Renaming module path:"
echo "  from: $OLD"
echo "  to:   $NEW"

files=$(grep -rl "$OLD" . \
  --exclude-dir=.git \
  --include='*.go' \
  --include='*.mod' \
  --include='*.md' \
  --include='Makefile' || true)

if [ -z "$files" ]; then
  echo "no files reference '$OLD'"
  exit 0
fi

# sed -i.bak then removing the backup works on both BSD (macOS) and GNU sed.
echo "$files" | while IFS= read -r f; do
  sed -i.bak "s|$OLD|$NEW|g" "$f"
  rm -f "$f.bak"
  echo "  updated $f"
done

echo "Done. Verify with: go build ./..."

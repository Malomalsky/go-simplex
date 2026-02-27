#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: ./scripts/release-notes.sh <version> [changelog_path]

Extracts the changelog section body for a version heading:
  ## [<version>] - YYYY-MM-DD
EOF
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 1
fi

VERSION="$1"
CHANGELOG_PATH="${2:-CHANGELOG.md}"

if [[ ! -f "$CHANGELOG_PATH" ]]; then
  echo "error: changelog file not found: $CHANGELOG_PATH" >&2
  exit 1
fi

set +e
NOTES="$(
  awk -v version="$VERSION" '
BEGIN {
  target = "^## \\[" version "\\]([[:space:]]|$)"
  in_section = 0
  found = 0
}
$0 ~ /^## \[/ {
  if (in_section) {
    exit
  }
  if ($0 ~ target) {
    in_section = 1
    found = 1
    next
  }
}
in_section {
  print
}
END {
  if (!found) {
    exit 2
  }
}
' "$CHANGELOG_PATH"
)"
AWK_STATUS=$?
set -e

if [[ "$AWK_STATUS" -ne 0 ]]; then
  if [[ "$AWK_STATUS" -eq 2 ]]; then
    echo "error: version section not found in $CHANGELOG_PATH: $VERSION" >&2
    exit 1
  fi
  echo "error: failed to extract release notes from $CHANGELOG_PATH" >&2
  exit 1
fi

if [[ -z "${NOTES//[[:space:]]/}" ]]; then
  echo "No changelog notes provided for $VERSION."
  exit 0
fi

echo "$NOTES"

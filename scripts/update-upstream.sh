#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
UPSTREAM_BRANCH="${1:-stable}"
BASE_URL="https://raw.githubusercontent.com/simplex-chat/simplex-chat/${UPSTREAM_BRANCH}"

echo "Updating upstream snapshots from branch: ${UPSTREAM_BRANCH}"

curl -fsSL "${BASE_URL}/bots/api/COMMANDS.md" -o "${ROOT_DIR}/spec/upstream/COMMANDS.md"
curl -fsSL "${BASE_URL}/packages/simplex-chat-client/types/typescript/src/commands.ts" -o "${ROOT_DIR}/spec/upstream/commands.ts"
curl -fsSL "${BASE_URL}/packages/simplex-chat-client/types/typescript/src/events.ts" -o "${ROOT_DIR}/spec/upstream/events.ts"
curl -fsSL "${BASE_URL}/packages/simplex-chat-client/types/typescript/src/responses.ts" -o "${ROOT_DIR}/spec/upstream/responses.ts"
curl -fsSL "${BASE_URL}/packages/simplex-chat-client/types/typescript/src/types.ts" -o "${ROOT_DIR}/spec/upstream/types.ts"

echo "Regenerating Go files..."
(cd "${ROOT_DIR}" && go run ./cmd/simplexgen)

echo "Done."

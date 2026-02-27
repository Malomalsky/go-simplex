#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

echo "[compat] regenerate generated files"
go run ./cmd/simplexgen

echo "[compat] check generated files are up-to-date"
git diff --exit-code -- \
  sdk/command/generated_catalog.go \
  sdk/command/generated_requests.go \
  sdk/client/generated_senders.go \
  sdk/types/generated_tags.go \
  sdk/types/generated_records.go \
  sdk/types/generated_types.go

count_or_zero() {
  local pattern="$1"
  local file="$2"
  local result
  result="$(rg -c "$pattern" "$file" 2>/dev/null || true)"
  if [[ -z "$result" ]]; then
    echo 0
    return
  fi
  echo "$result"
}

upstream_cmds="$(count_or_zero '^export interface ' spec/upstream/commands.ts)"
generated_reqs="$(count_or_zero '^type [A-Za-z0-9_]+ struct' sdk/command/generated_requests.go)"
generated_senders="$(count_or_zero '^func \(c \*Client\) Send[A-Za-z0-9_]+' sdk/client/generated_senders.go)"

upstream_events="$(count_or_zero '^  export interface ' spec/upstream/events.ts)"
generated_event_tags="$(count_or_zero '^\s*EventType' sdk/types/generated_tags.go)"

upstream_responses="$(count_or_zero '^  export interface ' spec/upstream/responses.ts)"
generated_response_tags="$(count_or_zero '^\s*ResponseType' sdk/types/generated_tags.go)"

echo "[compat] upstream commands: $upstream_cmds"
echo "[compat] generated request structs: $generated_reqs"
echo "[compat] generated sender funcs: $generated_senders"
echo "[compat] upstream events: $upstream_events"
echo "[compat] generated event tags: $generated_event_tags"
echo "[compat] upstream responses: $upstream_responses"
echo "[compat] generated response tags: $generated_response_tags"

if [[ "$upstream_cmds" != "$generated_reqs" ]]; then
  echo "[compat] mismatch: command interfaces != generated request structs" >&2
  exit 1
fi
if [[ "$upstream_cmds" != "$generated_senders" ]]; then
  echo "[compat] mismatch: command interfaces != generated sender funcs" >&2
  exit 1
fi
if [[ "$upstream_events" != "$generated_event_tags" ]]; then
  echo "[compat] mismatch: upstream events != generated event tags" >&2
  exit 1
fi
if [[ "$upstream_responses" != "$generated_response_tags" ]]; then
  echo "[compat] mismatch: upstream responses != generated response tags" >&2
  exit 1
fi

echo "[compat] OK"

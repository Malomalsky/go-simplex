#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

WS_URL="${SIMPLEX_WS_URL:-}"
STARTED_PID=""
LOG_FILE=""

cleanup() {
  if [[ -n "$STARTED_PID" ]]; then
    if kill -0 "$STARTED_PID" >/dev/null 2>&1; then
      kill "$STARTED_PID" >/dev/null 2>&1 || true
      wait "$STARTED_PID" >/dev/null 2>&1 || true
    fi
  fi
}
trap cleanup EXIT INT TERM

if [[ -z "$WS_URL" ]]; then
  SIMPLEX_BIN="${SIMPLEX_BIN:-simplex-chat}"
  SIMPLEX_WS_PORT="${SIMPLEX_WS_PORT:-5225}"
  WS_URL="ws://localhost:${SIMPLEX_WS_PORT}"
  LOG_FILE="${SIMPLEX_LOG_FILE:-/tmp/go-simplex-integration-${SIMPLEX_WS_PORT}.log}"

  if ! command -v "$SIMPLEX_BIN" >/dev/null 2>&1; then
    echo "[integration] error: '$SIMPLEX_BIN' not found" >&2
    echo "[integration] install SimpleX CLI or set SIMPLEX_BIN/SIMPLEX_WS_URL" >&2
    exit 1
  fi

  echo "[integration] starting $SIMPLEX_BIN -p $SIMPLEX_WS_PORT"
  "$SIMPLEX_BIN" -p "$SIMPLEX_WS_PORT" >"$LOG_FILE" 2>&1 &
  STARTED_PID="$!"

  echo "[integration] waiting for websocket at $WS_URL"
  ready=0
  for _ in $(seq 1 60); do
    if go run ./cmd/simplex-smoke --ws "$WS_URL" >/dev/null 2>&1; then
      ready=1
      break
    fi
    sleep 1
  done

  if [[ "$ready" -ne 1 ]]; then
    echo "[integration] websocket was not ready in time: $WS_URL" >&2
    if [[ -n "$LOG_FILE" && -f "$LOG_FILE" ]]; then
      echo "[integration] last simplex log lines:" >&2
      tail -n 50 "$LOG_FILE" >&2 || true
    fi
    exit 1
  fi
fi

export SIMPLEX_WS_URL="$WS_URL"

echo "[integration] running tests against $SIMPLEX_WS_URL"
go test -tags=integration ./integration/... -v "$@"

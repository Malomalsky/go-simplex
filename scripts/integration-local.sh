#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

NO_START=0
VERBOSE=0
WAIT_TIMEOUT="${SIMPLEX_WAIT_TIMEOUT:-60}"
SIMPLEX_BIN="${SIMPLEX_BIN:-simplex-chat}"
SIMPLEX_WS_PORT="${SIMPLEX_WS_PORT:-5225}"
WS_URL="${SIMPLEX_WS_URL:-}"
STARTED_PID=""
LOG_FILE=""
TEST_ARGS=()

usage() {
  cat <<'EOF'
Usage: ./scripts/integration-local.sh [options] [-- go_test_args...]

Options:
  --no-start        do not launch simplex-chat; use existing SIMPLEX_WS_URL or ws://localhost:<port>
  --port <port>     websocket port for local simplex-chat start/default URL (default: 5225)
  --timeout <sec>   websocket readiness timeout in seconds when auto-starting simplex-chat (default: 60)
  --verbose         print extra diagnostics while waiting/running tests
  -h, --help        show this help

Environment:
  SIMPLEX_WS_URL      full websocket URL (if set, script does not auto-start simplex-chat)
  SIMPLEX_BIN         simplex binary name/path for auto-start (default: simplex-chat)
  SIMPLEX_LOG_FILE    log path for auto-started simplex process
EOF
}

log() {
  echo "[integration] $*"
}

debug() {
  if [[ "$VERBOSE" -eq 1 ]]; then
    log "$*"
  fi
}

is_uint() {
  [[ "$1" =~ ^[0-9]+$ ]]
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --no-start)
      NO_START=1
      shift
      ;;
    --port)
      if [[ $# -lt 2 ]]; then
        echo "[integration] error: --port expects value" >&2
        exit 1
      fi
      SIMPLEX_WS_PORT="$2"
      shift 2
      ;;
    --timeout)
      if [[ $# -lt 2 ]]; then
        echo "[integration] error: --timeout expects value" >&2
        exit 1
      fi
      WAIT_TIMEOUT="$2"
      shift 2
      ;;
    --verbose)
      VERBOSE=1
      shift
      ;;
    --)
      shift
      TEST_ARGS+=("$@")
      break
      ;;
    *)
      TEST_ARGS+=("$1")
      shift
      ;;
  esac
done

if ! is_uint "$SIMPLEX_WS_PORT"; then
  echo "[integration] error: --port must be a non-negative integer (got '$SIMPLEX_WS_PORT')" >&2
  exit 1
fi
if ! is_uint "$WAIT_TIMEOUT"; then
  echo "[integration] error: --timeout must be a non-negative integer (got '$WAIT_TIMEOUT')" >&2
  exit 1
fi

if [[ -z "$WS_URL" ]]; then
  WS_URL="ws://localhost:${SIMPLEX_WS_PORT}"
fi

LOG_FILE="${SIMPLEX_LOG_FILE:-/tmp/go-simplex-integration-${SIMPLEX_WS_PORT}.log}"

cleanup() {
  if [[ -n "$STARTED_PID" ]]; then
    if kill -0 "$STARTED_PID" >/dev/null 2>&1; then
      kill "$STARTED_PID" >/dev/null 2>&1 || true
      wait "$STARTED_PID" >/dev/null 2>&1 || true
    fi
  fi
}
trap cleanup EXIT INT TERM

if [[ "$NO_START" -eq 0 && -z "${SIMPLEX_WS_URL:-}" ]]; then
  if ! command -v "$SIMPLEX_BIN" >/dev/null 2>&1; then
    echo "[integration] error: '$SIMPLEX_BIN' not found" >&2
    echo "[integration] install SimpleX CLI or set SIMPLEX_BIN/SIMPLEX_WS_URL" >&2
    exit 1
  fi

  log "starting $SIMPLEX_BIN -p $SIMPLEX_WS_PORT"
  "$SIMPLEX_BIN" -p "$SIMPLEX_WS_PORT" >"$LOG_FILE" 2>&1 &
  STARTED_PID="$!"
  debug "simplex log file: $LOG_FILE"

  log "waiting for websocket at $WS_URL (timeout ${WAIT_TIMEOUT}s)"
  ready=0
  for attempt in $(seq 1 "$WAIT_TIMEOUT"); do
    if go run ./cmd/simplex-smoke --ws "$WS_URL" >/dev/null 2>&1; then
      ready=1
      break
    fi
    debug "websocket not ready yet (${attempt}/${WAIT_TIMEOUT})"
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
elif [[ "$NO_START" -eq 1 ]]; then
  log "--no-start enabled; using existing websocket at $WS_URL"
fi

export SIMPLEX_WS_URL="$WS_URL"

log "running tests against $SIMPLEX_WS_URL"
go test -tags=integration ./integration/... -v "${TEST_ARGS[@]}"

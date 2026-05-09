#!/usr/bin/env bash
# Run the collector against the real message queue (HTTP) and optionally POST a sample CSV.
#
# Prerequisite: your telemetry-message-queue exposes HTTP compatible with:
#   - POST {MQ_HTTP_BASE}{MQ_HTTP_ENQUEUE_PATH}  (optional) body = CSV file
#   - GET  {MQ_HTTP_BASE}{MQ_HTTP_PULL_PATH}?limit=N  → 200 JSON:
#        {"items":[{"id":"optional","body":"<one CSV data row or JSON telemetry>"}]}
#   - POST {MQ_HTTP_BASE}{MQ_HTTP_ACK_PATH}  body {"id":"..."}  (optional, when id returned)
#
# Defaults match a typical dev layout (host maps container 8080 → 9002).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export QUEUE_BACKEND="${QUEUE_BACKEND:-http}"
export MQ_HTTP_BASE="${MQ_HTTP_BASE:-http://host.docker.internal:9002}"
export MQ_HTTP_PULL_PATH="${MQ_HTTP_PULL_PATH:-/pull}"
export MQ_HTTP_ACK_PATH="${MQ_HTTP_ACK_PATH:-/ack}"
export MQ_HTTP_ENQUEUE_PATH="${MQ_HTTP_ENQUEUE_PATH:-/enqueue}"

SKIP_ENQUEUE="${SKIP_ENQUEUE:-0}"
CSV_FILE="${1:-$ROOT/scripts/sample_telemetry.csv}"

echo "==> QUEUE_BACKEND=${QUEUE_BACKEND} MQ_HTTP_BASE=${MQ_HTTP_BASE}"
echo "==> pull=${MQ_HTTP_PULL_PATH} ack=${MQ_HTTP_ACK_PATH} enqueue=${MQ_HTTP_ENQUEUE_PATH}"

if [[ "${SKIP_ENQUEUE}" != "1" ]]; then
  echo "==> POST sample CSV to queue (${CSV_FILE})"
  if ! curl -fsS -X POST "${MQ_HTTP_BASE}${MQ_HTTP_ENQUEUE_PATH}" \
    -H "Content-Type: text/csv" \
    --data-binary @"${CSV_FILE}"; then
    echo "WARN: enqueue failed (your queue may use a different path or body). Set SKIP_ENQUEUE=1 to skip and enqueue manually."
  fi
else
  echo "==> SKIP_ENQUEUE=1 — not posting CSV (enqueue manually if needed)."
fi

echo "==> Starting collector (Ctrl+C to stop)"
exec go run ./cmd/collector

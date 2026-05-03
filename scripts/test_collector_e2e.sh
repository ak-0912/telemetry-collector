#!/usr/bin/env bash
# End-to-end: seed mock queue file, run collector, verify PostgreSQL row.
# Assumes Postgres is reachable and the telemetry schema already exists (owned by your DB service).
# Set E2E_WITH_DOCKER_COMPOSE=1 to start compose Postgres and verify via docker exec (table must exist).
#
# Default DATABASE_URL uses host.docker.internal:5433 (DB on the host / another compose stack).
# Run this script from the devcontainer or any environment where that hostname resolves.
# Optional: COLLECTOR_DATABASE_URL overrides only the collector process if verify DSN must differ.
# E2E_VERIFY_RETRIES: seconds to poll for the row (default 30).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

default_telemetry_dsn="postgres://telemetry:telemetry@host.docker.internal:5433/telemetry?sslmode=disable"
default_compose_dsn="postgres://telemetry:telemetry@localhost:5432/telemetry?sslmode=disable"

# Uses local psql when installed; otherwise postgres:16-alpine (needs Docker).
psql_remote() {
  if command -v psql >/dev/null 2>&1; then
    psql "$DATABASE_URL" "$@"
    return
  fi
  if [[ "$(uname -s)" == "Linux" ]]; then
    docker run --rm --add-host=host.docker.internal:host-gateway postgres:16-alpine psql "$DATABASE_URL" "$@"
  else
    docker run --rm postgres:16-alpine psql "$DATABASE_URL" "$@"
  fi
}

query_db() {
  local q="$1"
  if [[ "${E2E_WITH_DOCKER_COMPOSE:-0}" == "1" ]]; then
    docker compose exec -T postgres psql -U telemetry -d telemetry -tAqc "$q"
  else
    psql_remote -tAqc "$q"
  fi
}

explain_db() {
  if [[ "${E2E_WITH_DOCKER_COMPOSE:-0}" == "1" ]]; then
    docker compose exec -T postgres psql -U telemetry -d telemetry -c \
      "SELECT id, metric_name, uuid, host_name, value FROM telemetry ORDER BY id DESC LIMIT 5;"
  else
    psql_remote -c \
      "SELECT id, metric_name, uuid, host_name, value FROM telemetry ORDER BY id DESC LIMIT 5;"
  fi
}

if [[ "${E2E_WITH_DOCKER_COMPOSE:-0}" == "1" ]]; then
  export DATABASE_URL="${DATABASE_URL:-$default_compose_dsn}"
  echo "==> NOTE: E2E_WITH_DOCKER_COMPOSE=1 writes to THIS repo's Postgres (localhost:5432, user telemetry)."
  echo "    It does NOT use host.docker.internal:5433. Point pgAdmin at :5432 to see these rows."
  echo "==> Ensuring Postgres is up (docker compose)"
  docker compose up -d postgres

  echo "==> Waiting for Postgres"
  for _ in $(seq 1 30); do
    if docker compose exec -T postgres pg_isready -U telemetry -d telemetry >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done
  docker compose exec -T postgres pg_isready -U telemetry -d telemetry
else
  export DATABASE_URL="${DATABASE_URL:-${POSTGRES_DSN:-$default_telemetry_dsn}}"
  echo "==> Using external DB (default: host.docker.internal:5433, db telemetry, user telemetry)."
  echo "==> Match pgAdmin to this DSN (or set DATABASE_URL / COLLECTOR_DATABASE_URL)."
fi

TEST_UUID="$(uuidgen 2>/dev/null | tr '[:upper:]' '[:lower:]' || python3 -c 'import uuid; print(uuid.uuid4())')"
if command -v python3 >/dev/null 2>&1; then
  PROCESSED_NANO="$(python3 -c 'import time; print(time.time_ns())')"
else
  PROCESSED_NANO="$(($(date +%s) * 1000000000))"
fi
PAYLOAD_FILE="$(mktemp)"
COLLECTOR_LOG="$(mktemp)"
trap 'rm -f "$PAYLOAD_FILE" "$COLLECTOR_LOG"' EXIT

echo "==> E2E payload uuid=${TEST_UUID} processed_at_unix_nano=${PROCESSED_NANO}"

cat >"$PAYLOAD_FILE" <<EOF
{"metric_name":"gpu.temperature","gpu_id":"gpu-e2e-${TEST_UUID%%-*}","device":"nvidia0","uuid":"${TEST_UUID}","model_name":"A100","host_name":"e2e-${TEST_UUID%%-*}","value":72.5,"labels_raw":"{}","processed_at_unix_nano":${PROCESSED_NANO}}
EOF

export MOCK_QUEUE_PAYLOADS_FILE="$PAYLOAD_FILE"
export POLL_INTERVAL="${POLL_INTERVAL:-500ms}"
export WORKER_COUNT="${WORKER_COUNT:-4}"

VERIFY_DSN="$DATABASE_URL"
export DATABASE_URL="${COLLECTOR_DATABASE_URL:-$VERIFY_DSN}"

echo "==> Collector DATABASE_URL (insert target): ${DATABASE_URL}"
echo "==> Verify / psql uses after collector stops: ${VERIFY_DSN}"
if [[ -n "${COLLECTOR_DATABASE_URL:-}" ]]; then
  echo "==> (COLLECTOR_DATABASE_URL overrides collector only; verification uses VERIFY_DSN above.)"
fi

echo "==> Running collector in background (stops after processing mock payloads)"
echo "    Log: ${COLLECTOR_LOG}"
echo "    Stop any other collector using the same DB; otherwise you may see rows for a different uuid."
go run ./cmd/collector >"$COLLECTOR_LOG" 2>&1 &
collector_pid=$!
sleep 8
kill "$collector_pid" 2>/dev/null || true
wait "$collector_pid" 2>/dev/null || true

export DATABASE_URL="$VERIFY_DSN"

max_verify="${E2E_VERIFY_RETRIES:-30}"
echo "==> Verifying row for uuid=${TEST_UUID} (poll 1s / up to ${max_verify}s)"
row_count="0"
for i in $(seq 1 "$max_verify"); do
  row_count="$(query_db "SELECT COUNT(*) FROM telemetry WHERE uuid='${TEST_UUID}' AND metric_name='gpu.temperature';" | tr -d '[:space:]')"
  if [[ "${row_count}" == "1" ]]; then
    echo "    Row found after ${i}s."
    break
  fi
  if (( i == 1 || i % 5 == 0 )); then
    echo "    ... waiting (${i}/${max_verify}s, count=${row_count})"
  fi
  sleep 1
done

if [[ "${row_count}" != "1" ]]; then
  echo "FAIL: expected 1 row for THIS run's uuid=${TEST_UUID}, got count=${row_count}"
  set +e
  latest_uuid="$(query_db "SELECT uuid FROM telemetry ORDER BY id DESC LIMIT 1;" | tr -d '[:space:]')"
  set -e
  if [[ -n "${latest_uuid}" && "${latest_uuid}" != "${TEST_UUID}" ]]; then
    echo "Hint: latest row in table has uuid=${latest_uuid} (different from this run)."
    echo "      That usually means another collector inserted it, or this run's collector never reached the DB."
  fi
  if grep -q 'no such host' "$COLLECTOR_LOG" 2>/dev/null; then
    echo "Hint: collector could not resolve the DB hostname (expected: host.docker.internal)."
    echo "      Run ./scripts/test_collector_e2e.sh inside the devcontainer, or use a shell where host.docker.internal resolves."
    echo "      Devcontainer compose maps host.docker.internal via extra_hosts (Linux Docker)."
  fi
  if grep -q 'telemetry message processing failed' "$COLLECTOR_LOG" 2>/dev/null; then
    echo "==> Collector errors (last 15 lines):"
    tail -15 "$COLLECTOR_LOG" || true
  fi
  explain_db
  exit 1
fi

echo "OK: telemetry persisted (uuid=${TEST_UUID})"

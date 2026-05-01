# telemetry-collector
Telemetry Collector consumes telemetry from the custom queue and persists it. The implementation should support the ability to dynamically scale up/down the number of collectors.

## Architecture

- Clean Architecture + DDD separation:
  - `internal/domain`: telemetry aggregate and validation rules
  - `internal/application`: use case orchestration + repository ports
  - `internal/adapters/inbound`: queue consumer and protobuf payload processor
  - `internal/adapters/outbound`: PostgreSQL repository (Bun) and DLQ publisher
  - `internal/infrastructure`: Fx module, config, worker pool, retry policy

## Flow

1. Consumer pulls message batches from queue.
2. Worker pool processes messages concurrently.
3. Payload is mapped to domain input.
4. Domain validation enforces invariants (for example utilization 0-100).
5. Repository persists into PostgreSQL.
6. Error categories drive queue action:
   - Validation => reject/DLQ
   - Transient/system => retry with backoff
   - Success => ack

## Local Development

- Build: `make build`
- Run with containers: `make run`
- Stop containers: `make stop`
- Unit tests: `make test`
- Test coverage: `make test-coverage`

## Protobuf

- Proto schema lives in `api/telemetry/v1/telemetry.proto`.
- `buf.gen.yaml` and `buf.yaml` are included for protobuf generation.

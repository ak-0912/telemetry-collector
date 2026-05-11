# Future Improvements

## 1. DLQ — Publish to Custom MQ
Replace the log-only stub with a real producer that publishes rejected messages to a dedicated MQ topic (e.g. `gpu-telemetry-dlq`) with the failure reason and original payload.

## 2. Batch Inserts
Buffer validated messages and flush as a single bulk `INSERT` to reduce Postgres round-trips under load.

## 3. Health & Readiness Probes
Add `/healthz` and `/readyz` HTTP endpoints checking Postgres and MQ connectivity. Wire into Helm liveness/readiness probes.

## 4. Prometheus Metrics
Expose `/metrics` with counters (received, processed, failed, duplicates) and latency histograms for message processing and DB writes.

## 5. Structured Logging
Migrate from `log.Printf` to `log/slog` with JSON output and consistent trace fields (`metric_name`, `uuid`, `partition`).

## 6. Graceful Shutdown
On SIGTERM, stop polling but drain in-flight workers before exiting. Match `terminationGracePeriodSeconds` to the drain deadline.

## 7. Per-Message Retry Tracking
Track retry count per message and apply true exponential backoff. Route to DLQ after max retries.

## 8. DB Connection Pooling
Tune `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`. Add circuit breaker to prevent goroutine buildup during Postgres outages.

## 9. Consumer Rebalance Stability
Add backoff/cooldown on heartbeat-triggered rebalances so partition assignments settle instead of looping every 3 seconds.

## 10. Queue Lag-Based HPA
Expose consumer lag as a Prometheus metric and use KEDA or Prometheus Adapter for autoscaling instead of CPU-only HPA.

CREATE TABLE IF NOT EXISTS telemetry (
    id BIGSERIAL PRIMARY KEY,
    metric_name TEXT NOT NULL,
    gpu_id TEXT NOT NULL,
    device TEXT NOT NULL,
    uuid TEXT NOT NULL,
    model_name TEXT NOT NULL,
    host_name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    labels_raw TEXT NOT NULL,
    processed_at_unix_nano BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_uuid ON telemetry(uuid);
CREATE INDEX IF NOT EXISTS idx_telemetry_gpu_id ON telemetry(gpu_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_host_name ON telemetry(host_name);

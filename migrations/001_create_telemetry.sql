CREATE TABLE IF NOT EXISTS telemetry (
    id BIGSERIAL PRIMARY KEY,
    gpu_id TEXT NOT NULL,
    host_id TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    gpu_utilization DOUBLE PRECISION NOT NULL,
    memory_used_mb BIGINT NOT NULL,
    temperature_c DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_timestamp ON telemetry(timestamp);
CREATE INDEX IF NOT EXISTS idx_telemetry_gpu_id ON telemetry(gpu_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_host_id ON telemetry(host_id);

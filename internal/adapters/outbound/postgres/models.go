package postgres

import (
	"time"

	"github.com/uptrace/bun"
)

type TelemetryModel struct {
	bun.BaseModel `bun:"table:telemetry"`

	ID             int64     `bun:",pk,autoincrement"`
	GPUID          string    `bun:"gpu_id,notnull"`
	HostID         string    `bun:"host_id,notnull"`
	Timestamp      time.Time `bun:"timestamp,notnull"`
	GPUUtilization float64   `bun:"gpu_utilization,notnull"`
	MemoryUsedMB   int64     `bun:"memory_used_mb,notnull"`
	TemperatureC   float64   `bun:"temperature_c,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

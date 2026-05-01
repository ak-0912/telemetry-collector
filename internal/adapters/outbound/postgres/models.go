package postgres

import (
	"time"

	"github.com/uptrace/bun"
)

type TelemetryModel struct {
	bun.BaseModel `bun:"table:telemetry"`

	ID                  int64     `bun:",pk,autoincrement"`
	MetricName          string    `bun:"metric_name,notnull"`
	GPUID               string    `bun:"gpu_id,notnull"`
	Device              string    `bun:"device,notnull"`
	UUID                string    `bun:"uuid,notnull"`
	ModelName           string    `bun:"model_name,notnull"`
	HostName            string    `bun:"host_name,notnull"`
	Value               float64   `bun:"value,notnull"`
	LabelsRaw           string    `bun:"labels_raw,notnull"`
	ProcessedAtUnixNano int64     `bun:"processed_at_unix_nano,notnull"`
	CreatedAt           time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

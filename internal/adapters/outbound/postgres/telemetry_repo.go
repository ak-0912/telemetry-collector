// Package postgres implements the outbound adapter for persisting telemetry
// data to PostgreSQL using the Bun ORM.
package postgres

import (
	"context"
	"fmt"
	"log"
	"sync"

	domain "telemetry-collector/internal/domain/telemetry"

	"github.com/uptrace/bun"
)

// TelemetryRepository persists validated [domain.Telemetry] aggregates.
// It performs best-effort schema bootstrap on first write so the collector
// works against a fresh database without a separate migration step.
type TelemetryRepository struct {
	db               *bun.DB
	ensureSchemaOnce sync.Once
}

// NewTelemetryRepository constructs the repository with the given Bun database.
func NewTelemetryRepository(db *bun.DB) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

// Save inserts a telemetry record using an idempotent upsert
// (ON CONFLICT DO NOTHING on the natural key).
func (r *TelemetryRepository) Save(ctx context.Context, t domain.Telemetry) error {
	r.ensureSchemaOnce.Do(func() {
		if _, err := r.db.ExecContext(
			context.Background(),
			`CREATE UNIQUE INDEX IF NOT EXISTS telemetry_idempotency_key ON telemetry (metric_name, uuid, processed_at_unix_nano)`,
		); err != nil {
			log.Printf("telemetry repo: bootstrap index: %v", err)
		}
	})

	model := TelemetryModel{
		MetricName:          t.MetricName,
		GPUID:               t.GPUID,
		Device:              t.Device,
		UUID:                t.UUID,
		ModelName:           t.ModelName,
		HostName:            t.HostName,
		Value:               t.Value,
		LabelsRaw:           t.LabelsRaw,
		ProcessedAtUnixNano: t.ProcessedAtUnixNano,
	}

	if _, err := r.db.NewInsert().
		Model(&model).
		On("CONFLICT (metric_name, uuid, processed_at_unix_nano) DO NOTHING").
		Exec(ctx); err != nil {
		return fmt.Errorf("%w: insert telemetry: %v", domain.ErrTransient, err)
	}
	return nil
}

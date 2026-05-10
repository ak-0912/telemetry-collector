package postgres

import (
	"context"
	"fmt"
	"sync"

	domain "telemetry-collector/internal/domain/telemetry"

	"github.com/uptrace/bun"
)

type TelemetryRepository struct {
	db              *bun.DB
	ensureIndexOnce sync.Once
}

func NewTelemetryRepository(db *bun.DB) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

func (r *TelemetryRepository) Save(ctx context.Context, t domain.Telemetry) error {
	// Best-effort bootstrap so idempotent insert works without requiring a manual migration step.
	// If index creation fails (e.g. lacking privileges), keep prior behavior and attempt insert anyway.
	r.ensureIndexOnce.Do(func() {
		_, _ = r.db.ExecContext(
			context.Background(),
			`CREATE UNIQUE INDEX IF NOT EXISTS telemetry_idempotency_key ON telemetry (metric_name, uuid, processed_at_unix_nano)`,
		)
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

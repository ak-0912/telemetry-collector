package postgres

import (
	"context"
	"fmt"

	domain "telemetry-collector/internal/domain/telemetry"

	"github.com/uptrace/bun"
)

type TelemetryRepository struct {
	db *bun.DB
}

func NewTelemetryRepository(db *bun.DB) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

func (r *TelemetryRepository) Save(ctx context.Context, t domain.Telemetry) error {
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

	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return fmt.Errorf("%w: insert telemetry: %v", domain.ErrTransient, err)
	}
	return nil
}

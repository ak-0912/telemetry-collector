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
		GPUID:          t.GPUID,
		HostID:         t.HostID,
		Timestamp:      t.Timestamp,
		GPUUtilization: t.GPUUtilization,
		MemoryUsedMB:   t.MemoryUsedMB,
		TemperatureC:   t.TemperatureC,
	}

	if _, err := r.db.NewInsert().Model(&model).Exec(ctx); err != nil {
		return fmt.Errorf("%w: insert telemetry: %v", domain.ErrTransient, err)
	}
	return nil
}

package telemetry

import (
	"context"
	"fmt"

	domain "telemetry-collector/internal/domain/telemetry"
)

type ProcessUseCase struct {
	repo TelemetryRepository
}

func NewProcessUseCase(repo TelemetryRepository) *ProcessUseCase {
	return &ProcessUseCase{repo: repo}
}

func (u *ProcessUseCase) Execute(ctx context.Context, in Input) error {
	entity := domain.Telemetry{
		GPUID:          in.GPUID,
		HostID:         in.HostID,
		Timestamp:      in.Timestamp,
		GPUUtilization: in.GPUUtilization,
		MemoryUsedMB:   in.MemoryUsedMB,
		TemperatureC:   in.TemperatureC,
	}

	if err := entity.Validate(); err != nil {
		return err
	}
	if err := u.repo.Save(ctx, entity); err != nil {
		return fmt.Errorf("save telemetry: %w", err)
	}
	return nil
}

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
		MetricName:          in.MetricName,
		GPUID:               in.GPUID,
		Device:              in.Device,
		UUID:                in.UUID,
		ModelName:           in.ModelName,
		HostName:            in.HostName,
		Value:               in.Value,
		LabelsRaw:           in.LabelsRaw,
		ProcessedAtUnixNano: in.ProcessedAtUnixNano,
	}

	if err := entity.Validate(); err != nil {
		return err
	}
	if err := u.repo.Save(ctx, entity); err != nil {
		return fmt.Errorf("save telemetry: %w", err)
	}
	return nil
}

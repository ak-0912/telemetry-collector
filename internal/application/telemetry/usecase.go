package telemetry

import (
	"context"
	"fmt"

	domain "telemetry-collector/internal/domain/telemetry"
)

// ProcessUseCase orchestrates telemetry ingestion: maps the inbound DTO to the
// domain aggregate, runs validation, and delegates persistence to the repository.
type ProcessUseCase struct {
	repo TelemetryRepository
}

// NewProcessUseCase constructs the use case with the given repository port.
func NewProcessUseCase(repo TelemetryRepository) *ProcessUseCase {
	return &ProcessUseCase{repo: repo}
}

// Execute validates and persists a single telemetry data point.
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

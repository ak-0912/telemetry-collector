package telemetry

import (
	"context"

	domain "telemetry-collector/internal/domain/telemetry"
)

type TelemetryRepository interface {
	Save(ctx context.Context, t domain.Telemetry) error
}

type DLQPublisher interface {
	Publish(ctx context.Context, payload []byte, reason string) error
}

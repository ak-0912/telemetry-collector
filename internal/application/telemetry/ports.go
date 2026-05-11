// Package telemetry contains the application use-case orchestration for
// processing GPU telemetry data.
package telemetry

import (
	"context"

	domain "telemetry-collector/internal/domain/telemetry"
)

// TelemetryRepository is the outbound port for persisting validated telemetry.
type TelemetryRepository interface {
	Save(ctx context.Context, t domain.Telemetry) error
}

// DLQPublisher is the outbound port for publishing unprocessable messages to
// a dead-letter queue.
type DLQPublisher interface {
	Publish(ctx context.Context, payload []byte, reason string) error
}

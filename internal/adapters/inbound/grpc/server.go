package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	pb "telemetry-collector/gen/telemetry/v1"
	app "telemetry-collector/internal/application/telemetry"
)

type Processor struct {
	useCase *app.ProcessUseCase
}

func NewProcessor(useCase *app.ProcessUseCase) *Processor {
	return &Processor{useCase: useCase}
}

// Process deserializes protobuf payload and executes the use case.
func (p *Processor) Process(ctx context.Context, payload []byte) error {
	var msg pb.TelemetryMessage
	// Payload comes from protobuf-generated contracts. For local development, this
	// decoder accepts JSON-compatible payloads produced by mock queue services.
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("unmarshal telemetry protobuf: %w", err)
	}
	return p.useCase.Execute(ctx, ToInput(&msg))
}

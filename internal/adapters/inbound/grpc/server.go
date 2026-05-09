package grpc

import (
	"context"
	"fmt"

	app "telemetry-collector/internal/application/telemetry"
)

type Processor struct {
	useCase *app.ProcessUseCase
}

func NewProcessor(useCase *app.ProcessUseCase) *Processor {
	return &Processor{useCase: useCase}
}

// Process decodes JSON or CSV queue payloads and executes the use case.
func (p *Processor) Process(ctx context.Context, payload []byte) error {
	msg, err := parseTelemetryPayload(payload)
	if err != nil {
		return fmt.Errorf("decode telemetry payload: %w", err)
	}
	return p.useCase.Execute(ctx, ToInput(msg))
}

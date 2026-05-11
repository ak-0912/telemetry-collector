package queue

import (
	"context"
	"fmt"

	telemetryv1 "telemetry-collector/api/telemetry/v1"
	app "telemetry-collector/internal/application/telemetry"

	"google.golang.org/protobuf/proto"
)

// ProtoProcessor implements [Processor] for wire-encoded
// telemetry.v1.TelemetryMessage protobuf payloads.
type ProtoProcessor struct {
	useCase *app.ProcessUseCase
}

// NewProtoProcessor constructs a processor backed by the given use case.
func NewProtoProcessor(useCase *app.ProcessUseCase) *ProtoProcessor {
	return &ProtoProcessor{useCase: useCase}
}

// Process unmarshals a protobuf TelemetryMessage and delegates to the use case.
func (p *ProtoProcessor) Process(ctx context.Context, payload []byte) error {
	var msg telemetryv1.TelemetryMessage
	if err := proto.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("decode telemetry payload: protobuf: %w", err)
	}
	if msg.GetMetricName() == "" {
		return fmt.Errorf("decode telemetry payload: protobuf: missing metric_name")
	}
	return p.useCase.Execute(ctx, app.Input{
		MetricName:          msg.GetMetricName(),
		GPUID:               msg.GetGpuId(),
		Device:              msg.GetDevice(),
		UUID:                msg.GetUuid(),
		ModelName:           msg.GetModelName(),
		HostName:            msg.GetHostName(),
		Value:               msg.GetValue(),
		LabelsRaw:           msg.GetLabelsRaw(),
		ProcessedAtUnixNano: msg.GetProcessedAtUnixNano(),
	})
}

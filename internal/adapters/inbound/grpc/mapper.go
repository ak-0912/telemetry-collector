package grpc

import (
	pb "telemetry-collector/api/telemetry/v1"
	app "telemetry-collector/internal/application/telemetry"
)

func ToInput(msg *pb.TelemetryMessage) app.Input {
	return app.Input{
		MetricName:          msg.MetricName,
		GPUID:               msg.GpuId,
		Device:              msg.Device,
		UUID:                msg.Uuid,
		ModelName:           msg.ModelName,
		HostName:            msg.HostName,
		Value:               msg.Value,
		LabelsRaw:           msg.LabelsRaw,
		ProcessedAtUnixNano: msg.ProcessedAtUnixNano,
	}
}

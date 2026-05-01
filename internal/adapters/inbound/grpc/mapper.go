package grpc

import (
	pb "telemetry-collector/gen/telemetry/v1"
	app "telemetry-collector/internal/application/telemetry"
)

func ToInput(msg *pb.TelemetryMessage) app.Input {
	return app.Input{
		GPUID:          msg.GpuId,
		HostID:         msg.HostId,
		Timestamp:      msg.Timestamp.AsTime(),
		GPUUtilization: msg.GpuUtilization,
		MemoryUsedMB:   msg.MemoryUsedMb,
		TemperatureC:   msg.TemperatureC,
	}
}

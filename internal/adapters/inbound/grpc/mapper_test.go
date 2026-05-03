package grpc

import (
	"testing"
	pb "telemetry-collector/api/telemetry/v1"
)

func TestToInput(t *testing.T) {
	msg := &pb.TelemetryMessage{
		MetricName:          "gpu.temperature",
		GpuId:               "gpu-1",
		Device:              "nvidia0",
		Uuid:                "6a87a232-6556-4386-a3c0-0db1fd9ee579",
		ModelName:           "A100",
		HostName:            "node-1",
		Value:               70.5,
		LabelsRaw:           "{\"cluster\":\"prod\"}",
		ProcessedAtUnixNano: 1735689600000000000,
	}

	in := ToInput(msg)
	if in.MetricName != "gpu.temperature" || in.GPUID != "gpu-1" || in.UUID != "6a87a232-6556-4386-a3c0-0db1fd9ee579" {
		t.Fatalf("unexpected core mapping: %+v", in)
	}
	if in.Device != "nvidia0" || in.ModelName != "A100" || in.HostName != "node-1" {
		t.Fatalf("unexpected identity mapping: %+v", in)
	}
	if in.Value != 70.5 || in.LabelsRaw != "{\"cluster\":\"prod\"}" || in.ProcessedAtUnixNano != 1735689600000000000 {
		t.Fatalf("unexpected value mapping: %+v", in)
	}
}

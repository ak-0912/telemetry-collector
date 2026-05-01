package grpc

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "telemetry-collector/gen/telemetry/v1"
)

func TestToInput(t *testing.T) {
	ts := time.Now().UTC()
	msg := &pb.TelemetryMessage{
		GpuId:          "gpu-1",
		HostId:         "host-1",
		Timestamp:      timestamppb.New(ts),
		GpuUtilization: 51.2,
		MemoryUsedMb:   1024,
		TemperatureC:   70.5,
	}

	in := ToInput(msg)
	if in.GPUID != "gpu-1" || in.HostID != "host-1" {
		t.Fatalf("unexpected id mapping: %+v", in)
	}
	if in.Timestamp.Unix() != ts.Unix() {
		t.Fatalf("unexpected timestamp mapping: got %v want unix %v", in.Timestamp, ts.Unix())
	}
	if in.GPUUtilization != 51.2 || in.MemoryUsedMB != 1024 || in.TemperatureC != 70.5 {
		t.Fatalf("unexpected metric mapping: %+v", in)
	}
}

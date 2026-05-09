package queue

import (
	"context"
	"testing"

	telemetryv1 "telemetry-collector/api/telemetry/v1"
	app "telemetry-collector/internal/application/telemetry"
	domain "telemetry-collector/internal/domain/telemetry"

	"google.golang.org/protobuf/proto"
)

type protoRepo struct {
	last domain.Telemetry
}

func (r *protoRepo) Save(_ context.Context, t domain.Telemetry) error {
	r.last = t
	return nil
}

func TestProtoProcessorProcessSuccess(t *testing.T) {
	repo := &protoRepo{}
	processor := NewProtoProcessor(app.NewProcessUseCase(repo))
	wire, err := proto.Marshal(&telemetryv1.TelemetryMessage{
		MetricName:          "gpu.temperature",
		GpuId:               "0",
		Device:              "nvidia0",
		Uuid:                "6a87a232-6556-4386-a3c0-0db1fd9ee579",
		ModelName:           "H100",
		HostName:            "host-1",
		Value:               65.5,
		LabelsRaw:           `instance="host-1:9400",job="dcgm"`,
		ProcessedAtUnixNano: 1735689600000000000,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := processor.Process(context.Background(), wire); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.last.Value != 65.5 {
		t.Fatalf("value %v", repo.last.Value)
	}
	if repo.last.LabelsRaw != `instance="host-1:9400",job="dcgm"` {
		t.Fatalf("labels_raw %q", repo.last.LabelsRaw)
	}
}

func TestProtoProcessorProcessRejectsNonProtoPayload(t *testing.T) {
	repo := &protoRepo{}
	processor := NewProtoProcessor(app.NewProcessUseCase(repo))
	err := processor.Process(context.Background(), []byte(`instance="host:9400"`))
	if err == nil {
		t.Fatal("expected decode error")
	}
}

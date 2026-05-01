package grpc

import (
	"context"
	"testing"

	app "telemetry-collector/internal/application/telemetry"
	domain "telemetry-collector/internal/domain/telemetry"
)

type fakeRepo struct {
	saveErr error
}

func (f fakeRepo) Save(context.Context, domain.Telemetry) error { return f.saveErr }

func TestProcessorProcessSuccess(t *testing.T) {
	repo := fakeRepo{}
	useCase := app.NewProcessUseCase(repo)
	processor := NewProcessor(useCase)

	payload := []byte(`{"metric_name":"gpu.temperature","gpu_id":"gpu-1","device":"nvidia0","uuid":"d083db3f-88d3-4714-bcff-e0a4e95d709f","model_name":"A100","host_name":"host-1","value":65.5,"labels_raw":"{}","processed_at_unix_nano":1735689600000000000}`)

	if err := processor.Process(context.Background(), payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessorProcessInvalidPayload(t *testing.T) {
	repo := fakeRepo{}
	useCase := app.NewProcessUseCase(repo)
	processor := NewProcessor(useCase)

	if err := processor.Process(context.Background(), []byte("{bad-json")); err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestProcessorProcessValidationError(t *testing.T) {
	repo := fakeRepo{}
	useCase := app.NewProcessUseCase(repo)
	processor := NewProcessor(useCase)

	payload := []byte(`{"metric_name":"gpu.temperature","gpu_id":"","device":"nvidia0","uuid":"d083db3f-88d3-4714-bcff-e0a4e95d709f","model_name":"A100","host_name":"host-1","value":50,"labels_raw":"{}","processed_at_unix_nano":1735689600000000000}`)
	if err := processor.Process(context.Background(), payload); err == nil {
		t.Fatal("expected validation error")
	}
}

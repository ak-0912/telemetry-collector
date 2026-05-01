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

	payload := []byte(`{"gpu_id":"gpu-1","host_id":"host-1","timestamp":{"unix_seconds":1735689600},"gpu_utilization":60,"memory_used_mb":2048,"temperature_c":65.5}`)

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

	payload := []byte(`{"gpu_id":"","host_id":"host-1","timestamp":{"unix_seconds":1735689600},"gpu_utilization":42,"memory_used_mb":1234,"temperature_c":50}`)
	if err := processor.Process(context.Background(), payload); err == nil {
		t.Fatal("expected validation error")
	}
}

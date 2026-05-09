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

	payload := []byte(`{"metric_name":"gpu.temperature","gpu_id":"gpu-1","device":"nvidia0","uuid":"6a87a232-6556-4386-a3c0-0db1fd9ee579","model_name":"A100","host_name":"host-1","value":65.5,"labels_raw":"{}","processed_at_unix_nano":1735689600000000000}`)

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

func TestProcessorProcessCSVPayload(t *testing.T) {
	repo := fakeRepo{}
	useCase := app.NewProcessUseCase(repo)
	processor := NewProcessor(useCase)

	payload := []byte("metric_name,gpu_id,device,uuid,model_name,host_name,value,labels_raw,processed_at_unix_nano\n" +
		"gpu.temperature,gpu-1,nvidia0,6a87a232-6556-4386-a3c0-0db1fd9ee579,A100,host-csv,70.0,{},1735689600000000000\n")

	if err := processor.Process(context.Background(), payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestProcessorProcessValidationError(t *testing.T) {
	repo := fakeRepo{}
	useCase := app.NewProcessUseCase(repo)
	processor := NewProcessor(useCase)

	payload := []byte(`{"metric_name":"gpu.temperature","gpu_id":"","device":"nvidia0","uuid":"6a87a232-6556-4386-a3c0-0db1fd9ee579","model_name":"A100","host_name":"host-1","value":50,"labels_raw":"{}","processed_at_unix_nano":1735689600000000000}`)
	if err := processor.Process(context.Background(), payload); err == nil {
		t.Fatal("expected validation error")
	}
}

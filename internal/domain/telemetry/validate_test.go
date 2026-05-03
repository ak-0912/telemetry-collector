package telemetry

import (
	"errors"
	"testing"
)

func TestTelemetryValidate(t *testing.T) {
	valid := Telemetry{
		MetricName:          "gpu.temperature",
		GPUID:               "gpu-1",
		Device:              "nvidia0",
		UUID:                "6a87a232-6556-4386-a3c0-0db1fd9ee579",
		ModelName:           "A100",
		HostName:            "host-1",
		Value:               72.1,
		LabelsRaw:           "{}",
		ProcessedAtUnixNano: 1735689600000000000,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	invalid := valid
	invalid.ProcessedAtUnixNano = 0
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTelemetryValidateMissingGPUID(t *testing.T) {
	err := Telemetry{
		MetricName:          "gpu.temperature",
		UUID:                "6a87a232-6556-4386-a3c0-0db1fd9ee579",
		HostName:            "host-1",
		ProcessedAtUnixNano: 1735689600000000000,
	}.Validate()
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTelemetryValidateMissingUUID(t *testing.T) {
	err := Telemetry{
		MetricName:          "gpu.temperature",
		GPUID:               "gpu-1",
		HostName:            "host-1",
		ProcessedAtUnixNano: 1735689600000000000,
	}.Validate()
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTelemetryValidateMissingMetricName(t *testing.T) {
	err := Telemetry{
		GPUID:               "gpu-1",
		UUID:                "6a87a232-6556-4386-a3c0-0db1fd9ee579",
		HostName:            "host-1",
		ProcessedAtUnixNano: 1735689600000000000,
	}.Validate()
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTelemetryValidateMissingProcessedAt(t *testing.T) {
	err := Telemetry{
		MetricName: "gpu.temperature",
		GPUID:      "gpu-1",
		UUID:       "6a87a232-6556-4386-a3c0-0db1fd9ee579",
		HostName:   "host-1",
	}.Validate()
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestErrorHelpers(t *testing.T) {
	if !IsValidationError(ErrValidation) {
		t.Fatal("expected IsValidationError true")
	}
	if IsValidationError(ErrTransient) {
		t.Fatal("expected IsValidationError false")
	}
	if !IsTransientError(ErrTransient) {
		t.Fatal("expected IsTransientError true")
	}
	if IsTransientError(ErrSystem) {
		t.Fatal("expected IsTransientError false")
	}
}

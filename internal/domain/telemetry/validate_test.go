package telemetry

import (
	"errors"
	"testing"
	"time"
)

func TestTelemetryValidate(t *testing.T) {
	valid := Telemetry{
		GPUID:          "gpu-1",
		HostID:         "host-1",
		Timestamp:      time.Now().UTC(),
		GPUUtilization: 64.5,
		MemoryUsedMB:   8192,
		TemperatureC:   72.1,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	invalid := valid
	invalid.GPUUtilization = 120
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTelemetryValidateMissingGPUID(t *testing.T) {
	err := Telemetry{Timestamp: time.Now().UTC(), GPUUtilization: 10}.Validate()
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTelemetryValidateMissingTimestamp(t *testing.T) {
	err := Telemetry{GPUID: "gpu-1", GPUUtilization: 10}.Validate()
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTelemetryValidateNegativeMemory(t *testing.T) {
	err := Telemetry{GPUID: "gpu-1", Timestamp: time.Now().UTC(), GPUUtilization: 10, MemoryUsedMB: -1}.Validate()
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTelemetryValidateTemperatureOutOfRange(t *testing.T) {
	err := Telemetry{GPUID: "gpu-1", Timestamp: time.Now().UTC(), GPUUtilization: 10, TemperatureC: -500}.Validate()
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

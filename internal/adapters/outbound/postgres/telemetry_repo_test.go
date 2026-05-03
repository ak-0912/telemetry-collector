package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "telemetry-collector/internal/domain/telemetry"
)

func TestNewTelemetryRepository(t *testing.T) {
	repo := NewTelemetryRepository(nil)
	if repo == nil {
		t.Fatal("expected repository")
	}
}

func TestTelemetryRepositorySaveWrapsTransientError(t *testing.T) {
	db := NewBunDB("postgres://telemetry:telemetry@localhost:5432/telemetry?sslmode=disable")
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	err := repo.Save(ctx, domain.Telemetry{
		MetricName:          "gpu.temperature",
		GPUID:               "gpu-1",
		Device:              "nvidia0",
		UUID:                "6a87a232-6556-4386-a3c0-0db1fd9ee579",
		ModelName:           "A100",
		HostName:            "host-1",
		Value:               40,
		LabelsRaw:           "{}",
		ProcessedAtUnixNano: 1735689600000000000,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrTransient) {
		t.Fatalf("expected transient wrapped error, got %v", err)
	}
}

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
	db := NewBunDB("postgres://postgres:postgres@localhost:5432/telemetry?sslmode=disable")
	defer db.Close()

	repo := NewTelemetryRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := repo.Save(ctx, domain.Telemetry{
		GPUID:          "gpu-1",
		HostID:         "host-1",
		Timestamp:      time.Now().UTC(),
		GPUUtilization: 30,
		MemoryUsedMB:   10,
		TemperatureC:   40,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrTransient) {
		t.Fatalf("expected transient wrapped error, got %v", err)
	}
}

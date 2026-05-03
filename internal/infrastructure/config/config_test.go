package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadReadsEnvAndFallbacks(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_DSN", "postgres://x")
	t.Setenv("WORKER_COUNT", "3")
	t.Setenv("QUEUE_BATCH_SIZE", "10")
	t.Setenv("POLL_INTERVAL", "5s")

	cfg := Load()
	if cfg.PostgresDSN != "postgres://x" || cfg.WorkerCount != 3 || cfg.QueueBatchSize != 10 || cfg.PollInterval != 5*time.Second {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadPrefersDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://telemetry:telemetry@host.docker.internal:5433/telemetry?sslmode=disable")
	t.Setenv("POSTGRES_DSN", "postgres://ignored")
	t.Setenv("WORKER_COUNT", "1")
	t.Setenv("QUEUE_BATCH_SIZE", "1")
	t.Setenv("POLL_INTERVAL", "1s")

	cfg := Load()
	want := "postgres://telemetry:telemetry@host.docker.internal:5433/telemetry?sslmode=disable"
	if cfg.PostgresDSN != want {
		t.Fatalf("expected DATABASE_URL, got %q", cfg.PostgresDSN)
	}
}

func TestGetenvIntFallbackOnInvalid(t *testing.T) {
	t.Setenv("WORKER_COUNT", "bad")
	if got := getenvInt("WORKER_COUNT", 7); got != 7 {
		t.Fatalf("expected fallback, got %d", got)
	}
}

func TestGetenvDurationFallbackOnInvalid(t *testing.T) {
	t.Setenv("POLL_INTERVAL", "bad")
	if got := getenvDuration("POLL_INTERVAL", "2s"); got != 2*time.Second {
		t.Fatalf("expected fallback duration, got %v", got)
	}
}

func TestGetenvFallbackWhenMissing(t *testing.T) {
	_ = os.Unsetenv("MISSING_KEY")
	if got := getenv("MISSING_KEY", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

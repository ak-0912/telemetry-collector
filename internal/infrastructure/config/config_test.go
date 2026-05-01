package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadReadsEnvAndFallbacks(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://x")
	t.Setenv("WORKER_COUNT", "3")
	t.Setenv("QUEUE_BATCH_SIZE", "10")
	t.Setenv("POLL_INTERVAL", "5s")

	cfg := Load()
	if cfg.PostgresDSN != "postgres://x" || cfg.WorkerCount != 3 || cfg.QueueBatchSize != 10 || cfg.PollInterval != 5*time.Second {
		t.Fatalf("unexpected config: %+v", cfg)
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

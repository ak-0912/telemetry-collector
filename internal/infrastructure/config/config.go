package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	PostgresDSN    string
	WorkerCount    int
	QueueBatchSize int
	PollInterval   time.Duration
}

func Load() Config {
	return Config{
		PostgresDSN:    getenv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/telemetry?sslmode=disable"),
		WorkerCount:    getenvInt("WORKER_COUNT", 8),
		QueueBatchSize: getenvInt("QUEUE_BATCH_SIZE", 50),
		PollInterval:   getenvDuration("POLL_INTERVAL", "2s"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getenvDuration(key, fallback string) time.Duration {
	raw := getenv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		d, _ = time.ParseDuration(fallback)
	}
	return d
}

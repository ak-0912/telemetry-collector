// Package config loads runtime configuration from environment variables with
// sensible defaults for local development.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime settings for the telemetry collector.
type Config struct {
	PostgresDSN    string
	WorkerCount    int
	QueueBatchSize int
	PollInterval   time.Duration

	// QueueBackend selects the queue adapter: "mock", "http", or "grpc".
	QueueBackend string

	// HTTP queue backend.
	MQHTTPBase     string
	MQHTTPPullPath string
	MQHTTPAckPath  string

	// gRPC queue backend (mq.v1.MessageQueueService).
	MQGRPCAddr          string
	MQTopic             string
	MQGroup             string
	MQMemberID          string
	MQHeartbeatInterval time.Duration

	// MQGRPCPreferIPv4 forces tcp4 when dialing MQ_GRPC_ADDR to avoid
	// broken IPv6 for host.docker.internal on some Docker networks.
	MQGRPCPreferIPv4 bool
}

func resolvePostgresDSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return getenv("POSTGRES_DSN", "postgres://telemetry:telemetry@host.docker.internal:5433/telemetry?sslmode=disable")
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		PostgresDSN:         resolvePostgresDSN(),
		WorkerCount:         getenvInt("WORKER_COUNT", 8),
		QueueBatchSize:      getenvInt("QUEUE_BATCH_SIZE", 50),
		PollInterval:        getenvDuration("POLL_INTERVAL", "2s"),
		QueueBackend:        getenv("QUEUE_BACKEND", "mock"),
		MQHTTPBase:          getenv("MQ_HTTP_BASE", ""),
		MQHTTPPullPath:      getenv("MQ_HTTP_PULL_PATH", "/pull"),
		MQHTTPAckPath:       getenv("MQ_HTTP_ACK_PATH", "/ack"),
		MQGRPCAddr:          getenv("MQ_GRPC_ADDR", ""),
		MQTopic:             getenv("MQ_TOPIC", "gpu-telemetry"),
		MQGroup:             getenv("MQ_GROUP", "telemetry-collector"),
		MQMemberID:          getenv("MQ_MEMBER_ID", ""),
		MQHeartbeatInterval: getenvDuration("MQ_HEARTBEAT_INTERVAL", "3s"),
		MQGRPCPreferIPv4:    getenvBool("MQ_GRPC_PREFER_IPV4", true),
	}
}

func getenvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
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

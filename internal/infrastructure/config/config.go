package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	PostgresDSN    string
	WorkerCount    int
	QueueBatchSize int
	PollInterval   time.Duration
	// QueueBackend: "mock" (MOCK_QUEUE_PAYLOADS_FILE), "http" (MQ_HTTP_BASE), or "grpc" (MQ_GRPC_ADDR).
	QueueBackend   string
	MQHTTPBase     string
	MQHTTPPullPath string
	MQHTTPAckPath  string
	// gRPC message queue (mq.v1.MessageQueueService).
	MQGRPCAddr            string
	MQTopic               string
	MQGroup               string
	MQMemberID            string
	MQHeartbeatInterval   time.Duration
	// MQGRPCPreferIPv4 uses tcp4 when dialing MQ_GRPC_ADDR (avoids broken IPv6 for host.docker.internal in some Docker setups).
	MQGRPCPreferIPv4 bool
}

func resolvePostgresDSN() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return getenv("POSTGRES_DSN", "postgres://telemetry:telemetry@host.docker.internal:5433/telemetry?sslmode=disable")
}

func Load() Config {
	return Config{
		PostgresDSN:    resolvePostgresDSN(),
		WorkerCount:    getenvInt("WORKER_COUNT", 8),
		QueueBatchSize: getenvInt("QUEUE_BATCH_SIZE", 50),
		PollInterval:   getenvDuration("POLL_INTERVAL", "2s"),
		QueueBackend:   getenv("QUEUE_BACKEND", "mock"),
		MQHTTPBase:     getenv("MQ_HTTP_BASE", ""),
		MQHTTPPullPath: getenv("MQ_HTTP_PULL_PATH", "/pull"),
		MQHTTPAckPath:  getenv("MQ_HTTP_ACK_PATH", "/ack"),
		MQGRPCAddr:     getenv("MQ_GRPC_ADDR", ""),
		MQTopic:        getenv("MQ_TOPIC", "gpu-telemetry"),
		MQGroup:        getenv("MQ_GROUP", "telemetry-collector"),
		MQMemberID:     getenv("MQ_MEMBER_ID", ""),
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

// Package queue provides inbound adapters for consuming telemetry messages
// from the custom message queue (gRPC, HTTP, or mock backends).
package queue

import (
	"context"
	"time"
)

// Message represents a single message pulled from the queue.
// Implementations must support acknowledge, retry, and reject semantics.
type Message interface {
	Body() []byte
	Ack(ctx context.Context) error
	Retry(ctx context.Context, delay time.Duration) error
	Reject(ctx context.Context) error
}

// Client is the inbound port for pulling message batches from a queue backend.
type Client interface {
	Pull(ctx context.Context, batchSize int) ([]Message, error)
}

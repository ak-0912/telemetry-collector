// Package dlq provides a stub dead-letter queue publisher.
// Replace with a real DLQ backend (Kafka, SQS, etc.) in production.
package dlq

import (
	"context"
	"log"
)

// Producer is a no-op DLQ publisher that logs rejected payloads.
type Producer struct{}

// NewProducer constructs the stub producer.
func NewProducer() *Producer {
	return &Producer{}
}

// Publish logs the rejected payload metadata. A production implementation
// should forward the payload to a durable store for later inspection.
func (p *Producer) Publish(_ context.Context, payload []byte, reason string) error {
	log.Printf("dlq: reason=%q payload_bytes=%d", reason, len(payload))
	return nil
}

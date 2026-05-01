package dlq

import (
	"context"
	"log"
)

type Producer struct{}

func NewProducer() *Producer {
	return &Producer{}
}

func (p *Producer) Publish(_ context.Context, payload []byte, reason string) error {
	log.Printf("dlq publish reason=%s payload_size=%d", reason, len(payload))
	return nil
}

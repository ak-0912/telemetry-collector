package dlq

import (
	"context"
	"testing"
)

func TestProducerPublish(t *testing.T) {
	p := NewProducer()
	if err := p.Publish(context.Background(), []byte("payload"), "test-reason"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

package queue

import (
	"context"
	"time"
)

type Message interface {
	Body() []byte
	Ack(ctx context.Context) error
	Retry(ctx context.Context, delay time.Duration) error
	Reject(ctx context.Context) error
}

type Client interface {
	Pull(ctx context.Context, batchSize int) ([]Message, error)
}

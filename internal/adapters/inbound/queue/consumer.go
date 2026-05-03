package queue

import (
	"context"
	"errors"
	"log"
	"time"

	app "telemetry-collector/internal/application/telemetry"
	domain "telemetry-collector/internal/domain/telemetry"
	"telemetry-collector/internal/infrastructure/retry"
	"telemetry-collector/internal/infrastructure/workerpool"
)

type Processor interface {
	Process(ctx context.Context, payload []byte) error
}

type Consumer struct {
	client       Client
	processor    Processor
	dlq          app.DLQPublisher
	batchSize    int
	pollInterval time.Duration
	workers      *workerpool.Pool
	retryPolicy  retry.Policy
}

func NewConsumer(
	client Client,
	processor Processor,
	dlq app.DLQPublisher,
	workers *workerpool.Pool,
	batchSize int,
	pollInterval time.Duration,
	retryPolicy retry.Policy,
) *Consumer {
	return &Consumer{
		client:       client,
		processor:    processor,
		dlq:          dlq,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		workers:      workers,
		retryPolicy:  retryPolicy,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.pollOnce(ctx)
		}
	}
}

func (c *Consumer) pollOnce(ctx context.Context) {
	msgs, err := c.client.Pull(ctx, c.batchSize)
	if err != nil {
		log.Printf("queue pull failed: %v", err)
		return
	}

	for _, msg := range msgs {
		message := msg
		c.workers.Submit(func() { c.handleMessage(message) })
	}
}

// handleMessage uses [context.Background] for persistence and ack/retry so a
// stopped consumer (lifecycle cancel) does not cancel in-flight DB writes.
func (c *Consumer) handleMessage(msg Message) {
	workCtx := context.Background()
	err := c.processor.Process(workCtx, msg.Body())
	if err == nil {
		_ = msg.Ack(workCtx)
		return
	}

	log.Printf("telemetry message processing failed: %v", err)

	switch {
	case domain.IsValidationError(err):
		_ = c.dlq.Publish(workCtx, msg.Body(), err.Error())
		_ = msg.Reject(workCtx)
	case domain.IsTransientError(err) || errors.Is(err, domain.ErrSystem):
		delay := c.retryPolicy.NextDelay(1)
		_ = msg.Retry(workCtx, delay)
	default:
		_ = msg.Retry(workCtx, c.retryPolicy.NextDelay(1))
	}
}

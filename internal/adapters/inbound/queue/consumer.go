package queue

import (
	"context"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	telemetryv1 "telemetry-collector/api/telemetry/v1"
	app "telemetry-collector/internal/application/telemetry"
	domain "telemetry-collector/internal/domain/telemetry"
	"telemetry-collector/internal/infrastructure/retry"
	"telemetry-collector/internal/infrastructure/workerpool"

	"google.golang.org/protobuf/proto"
)

// Processor decodes a raw queue payload and persists the result.
type Processor interface {
	Process(ctx context.Context, payload []byte) error
}

// Consumer polls a queue Client on a fixed interval, fans messages out to a
// worker pool, and drives the ack / retry / DLQ lifecycle for each message.
type Consumer struct {
	client       Client
	processor    Processor
	dlq          app.DLQPublisher
	batchSize    int
	pollInterval time.Duration
	workers      *workerpool.Pool
	retryPolicy  retry.Policy
	deduper      *messageDeduper
}

// NewConsumer wires together the queue consumer with all its dependencies.
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
		deduper:      newMessageDeduper(10 * time.Minute),
	}
}

// Start blocks until ctx is cancelled, polling the queue on each tick.
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
	if len(msgs) > 0 {
		log.Printf("queue pull succeeded: fetched=%d", len(msgs))
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
	body := msg.Body()
	metricName, uuid, processedAtUnixNano := traceFieldsFromPayload(body)

	log.Printf("queue message received: payload_bytes=%d metric_name=%q uuid=%q", len(body), metricName, uuid)

	if key := idempotencyKey(metricName, uuid, processedAtUnixNano); key != "" && c.deduper.Seen(key) {
		log.Printf("queue duplicate skipped: metric_name=%q uuid=%q processed_at_unix_nano=%d", metricName, uuid, processedAtUnixNano)
		if ackErr := msg.Ack(workCtx); ackErr != nil {
			log.Printf("queue ack failed (duplicate): metric_name=%q uuid=%q err=%v", metricName, uuid, ackErr)
		}
		return
	}

	err := c.processor.Process(workCtx, body)
	if err == nil {
		if key := idempotencyKey(metricName, uuid, processedAtUnixNano); key != "" {
			c.deduper.Add(key)
		}
		log.Printf("queue message saved: metric_name=%q uuid=%q", metricName, uuid)
		if ackErr := msg.Ack(workCtx); ackErr != nil {
			log.Printf("queue ack failed: metric_name=%q uuid=%q err=%v", metricName, uuid, ackErr)
		}
		return
	}

	log.Printf("queue message processing failed: metric_name=%q uuid=%q err=%v", metricName, uuid, err)

	switch {
	case domain.IsValidationError(err):
		if pubErr := c.dlq.Publish(workCtx, msg.Body(), err.Error()); pubErr != nil {
			log.Printf("dlq publish failed: metric_name=%q uuid=%q err=%v", metricName, uuid, pubErr)
		}
		if rejErr := msg.Reject(workCtx); rejErr != nil {
			log.Printf("queue reject failed: metric_name=%q uuid=%q err=%v", metricName, uuid, rejErr)
		}
	case domain.IsTransientError(err) || errors.Is(err, domain.ErrSystem):
		delay := c.retryPolicy.NextDelay(1)
		if retryErr := msg.Retry(workCtx, delay); retryErr != nil {
			log.Printf("queue retry failed: metric_name=%q uuid=%q err=%v", metricName, uuid, retryErr)
		}
	default:
		delay := c.retryPolicy.NextDelay(1)
		if retryErr := msg.Retry(workCtx, delay); retryErr != nil {
			log.Printf("queue retry failed: metric_name=%q uuid=%q err=%v", metricName, uuid, retryErr)
		}
	}
}

func traceFieldsFromPayload(payload []byte) (metricName, uuid string, processedAtUnixNano int64) {
	var tm telemetryv1.TelemetryMessage
	if err := proto.Unmarshal(payload, &tm); err != nil {
		return "unknown", "unknown", 0
	}
	metricName = tm.GetMetricName()
	uuid = tm.GetUuid()
	processedAtUnixNano = tm.GetProcessedAtUnixNano()
	if metricName == "" {
		metricName = "unknown"
	}
	if uuid == "" {
		uuid = "unknown"
	}
	return metricName, uuid, processedAtUnixNano
}

func idempotencyKey(metricName, uuid string, processedAtUnixNano int64) string {
	if metricName == "" || metricName == "unknown" || uuid == "" || uuid == "unknown" || processedAtUnixNano <= 0 {
		return ""
	}
	return metricName + "\x00" + uuid + "\x00" + strconv.FormatInt(processedAtUnixNano, 10)
}

// messageDeduper is a TTL-based in-memory set that prevents re-processing of
// recently seen messages (keyed on metric_name + uuid + processed_at_unix_nano).
type messageDeduper struct {
	mu    sync.Mutex
	ttl   time.Duration
	seen  map[string]time.Time
	sweep int
}

func newMessageDeduper(ttl time.Duration) *messageDeduper {
	return &messageDeduper{
		ttl:  ttl,
		seen: make(map[string]time.Time),
	}
}

// Seen reports whether key has been recorded within the TTL window.
func (d *messageDeduper) Seen(key string) bool {
	if key == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maybeSweepLocked(time.Now())
	exp, ok := d.seen[key]
	return ok && exp.After(time.Now())
}

// Add records key with the configured TTL.
func (d *messageDeduper) Add(key string) {
	if key == "" {
		return
	}
	now := time.Now()
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen[key] = now.Add(d.ttl)
	d.maybeSweepLocked(now)
}

func (d *messageDeduper) maybeSweepLocked(now time.Time) {
	d.sweep++
	if d.sweep%256 != 0 {
		return
	}
	for k, exp := range d.seen {
		if !exp.After(now) {
			delete(d.seen, k)
		}
	}
}

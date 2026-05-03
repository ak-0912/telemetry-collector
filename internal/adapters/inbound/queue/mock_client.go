package queue

import (
	"bytes"
	"context"
	"log"
	"os"
	"sync"
	"time"
)

// MockClient is a placeholder queue adapter. When MOCK_QUEUE_PAYLOADS_FILE is set,
// each non-empty line in that file is treated as one message body (JSON matching
// TelemetryMessage) and returned from Pull until the in-memory queue is drained.
type MockClient struct {
	mu    sync.Mutex
	queue [][]byte
}

func NewMockClient() *MockClient {
	c := &MockClient{}
	path := os.Getenv("MOCK_QUEUE_PAYLOADS_FILE")
	if path == "" {
		return c
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("mock queue: read %s: %v", path, err)
		return c
	}
	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		c.queue = append(c.queue, append([]byte(nil), line...))
	}
	log.Printf("mock queue: loaded %d payload(s) from %s", len(c.queue), path)
	return c
}

func (c *MockClient) Pull(_ context.Context, batchSize int) ([]Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.queue) == 0 {
		return nil, nil
	}
	if batchSize <= 0 {
		batchSize = 1
	}
	n := batchSize
	if n > len(c.queue) {
		n = len(c.queue)
	}
	out := make([]Message, 0, n)
	for i := 0; i < n; i++ {
		payload := c.queue[0]
		c.queue = c.queue[1:]
		out = append(out, NewMockMessage(payload))
	}
	return out, nil
}

type MockMessage struct {
	payload []byte
}

func NewMockMessage(payload []byte) *MockMessage {
	return &MockMessage{payload: payload}
}

func (m *MockMessage) Body() []byte { return m.payload }
func (m *MockMessage) Ack(context.Context) error {
	return nil
}
func (m *MockMessage) Retry(context.Context, time.Duration) error {
	return nil
}
func (m *MockMessage) Reject(context.Context) error {
	return nil
}

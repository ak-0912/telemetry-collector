package queue

import (
	"context"
	"time"
)

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) Pull(_ context.Context, _ int) ([]Message, error) {
	// Placeholder queue adapter until custom queue client is integrated.
	return nil, nil
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

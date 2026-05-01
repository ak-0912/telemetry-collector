package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	app "telemetry-collector/internal/application/telemetry"
	domain "telemetry-collector/internal/domain/telemetry"
	"telemetry-collector/internal/infrastructure/retry"
	"telemetry-collector/internal/infrastructure/workerpool"

	"github.com/golang/mock/gomock"
)

type fakeDLQ struct{}

func (f fakeDLQ) Publish(context.Context, []byte, string) error { return nil }

func newTestConsumer(t *testing.T, client Client, processor Processor, dlq app.DLQPublisher) *Consumer {
	t.Helper()
	return NewConsumer(client, processor, dlq, workerpool.New(1), 2, time.Millisecond, retry.NewPolicy())
}

func TestConsumerHandleMessageAckOnSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockQueueClient(ctrl)
	processor := NewMockProcessor(ctrl)
	dlq := fakeDLQ{}
	msg := NewMockQueueMessage(ctrl)
	c := newTestConsumer(t, client, processor, dlq)
	defer c.workers.Close()

	payload := []byte("payload")
	msg.EXPECT().Body().Return(payload).Times(1)
	processor.EXPECT().Process(gomock.Any(), payload).Return(nil)
	msg.EXPECT().Ack(gomock.Any()).Return(nil)

	c.handleMessage(context.Background(), msg)
}

func TestConsumerHandleMessageValidationErrorToDLQ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockQueueClient(ctrl)
	processor := NewMockProcessor(ctrl)
	dlq := fakeDLQ{}
	msg := NewMockQueueMessage(ctrl)
	c := newTestConsumer(t, client, processor, dlq)
	defer c.workers.Close()

	payload := []byte("invalid")
	err := errors.New(domain.ErrValidation.Error())
	err = errors.Join(domain.ErrValidation, err)

	msg.EXPECT().Body().Return(payload).Times(2)
	processor.EXPECT().Process(gomock.Any(), payload).Return(err)
	msg.EXPECT().Reject(gomock.Any()).Return(nil)

	c.handleMessage(context.Background(), msg)
}

func TestConsumerHandleMessageTransientRetries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockQueueClient(ctrl)
	processor := NewMockProcessor(ctrl)
	dlq := fakeDLQ{}
	msg := NewMockQueueMessage(ctrl)
	c := newTestConsumer(t, client, processor, dlq)
	defer c.workers.Close()

	payload := []byte("retry")
	msg.EXPECT().Body().Return(payload).Times(1)
	processor.EXPECT().Process(gomock.Any(), payload).Return(domain.ErrTransient)
	msg.EXPECT().Retry(gomock.Any(), retry.NewPolicy().BaseDelay).Return(nil)

	c.handleMessage(context.Background(), msg)
}

func TestConsumerPollOnceSubmitsMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockQueueClient(ctrl)
	processor := NewMockProcessor(ctrl)
	dlq := fakeDLQ{}
	msg := NewMockQueueMessage(ctrl)
	c := newTestConsumer(t, client, processor, dlq)
	defer c.workers.Close()

	payload := []byte("ok")
	client.EXPECT().Pull(gomock.Any(), 2).Return([]Message{msg}, nil)
	msg.EXPECT().Body().Return(payload).Times(1)
	processor.EXPECT().Process(gomock.Any(), payload).Return(nil)
	msg.EXPECT().Ack(gomock.Any()).Return(nil)

	c.pollOnce(context.Background())
	time.Sleep(10 * time.Millisecond)
}

func TestConsumerPollOncePullError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockQueueClient(ctrl)
	processor := NewMockProcessor(ctrl)
	c := newTestConsumer(t, client, processor, fakeDLQ{})
	defer c.workers.Close()

	client.EXPECT().Pull(gomock.Any(), 2).Return(nil, errors.New("pull failed"))
	c.pollOnce(context.Background())
}

func TestConsumerHandleMessageSystemErrorRetries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockQueueClient(ctrl)
	processor := NewMockProcessor(ctrl)
	msg := NewMockQueueMessage(ctrl)
	c := newTestConsumer(t, client, processor, fakeDLQ{})
	defer c.workers.Close()

	payload := []byte("retry")
	msg.EXPECT().Body().Return(payload).Times(1)
	processor.EXPECT().Process(gomock.Any(), payload).Return(domain.ErrSystem)
	msg.EXPECT().Retry(gomock.Any(), retry.NewPolicy().BaseDelay).Return(nil)

	c.handleMessage(context.Background(), msg)
}

func TestConsumerHandleMessageUnknownErrorRetries(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := NewMockQueueClient(ctrl)
	processor := NewMockProcessor(ctrl)
	msg := NewMockQueueMessage(ctrl)
	c := newTestConsumer(t, client, processor, fakeDLQ{})
	defer c.workers.Close()

	payload := []byte("retry")
	msg.EXPECT().Body().Return(payload).Times(1)
	processor.EXPECT().Process(gomock.Any(), payload).Return(errors.New("unexpected"))
	msg.EXPECT().Retry(gomock.Any(), retry.NewPolicy().BaseDelay).Return(nil)

	c.handleMessage(context.Background(), msg)
}

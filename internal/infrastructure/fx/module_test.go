package fxmodule

import (
	"context"
	"testing"
	"time"

	inboundgrpc "telemetry-collector/internal/adapters/inbound/grpc"
	"telemetry-collector/internal/adapters/inbound/queue"
	"telemetry-collector/internal/adapters/outbound/dlq"
	app "telemetry-collector/internal/application/telemetry"
	domain "telemetry-collector/internal/domain/telemetry"
	"telemetry-collector/internal/infrastructure/config"
	"telemetry-collector/internal/infrastructure/retry"

	"go.uber.org/fx"
)

type fakeLifecycle struct {
	hook fx.Hook
}

func (f *fakeLifecycle) Append(h fx.Hook) {
	f.hook = h
}

type fakeRepo struct{}

func (fakeRepo) Save(context.Context, domain.Telemetry) error { return nil }

func TestProvideWorkerPool(t *testing.T) {
	pool := provideWorkerPool(config.Config{WorkerCount: 1})
	if pool == nil {
		t.Fatal("expected worker pool")
	}
	pool.Close()
}

func TestProvideConsumer(t *testing.T) {
	cfg := config.Config{QueueBatchSize: 3, PollInterval: time.Second}
	c := provideConsumer(cfg, nil, nil, nil, nil, retry.NewPolicy())
	if c == nil {
		t.Fatal("expected consumer")
	}
}

func TestRunConsumerLifecycleHooks(t *testing.T) {
	cfg := config.Config{QueueBatchSize: 1, PollInterval: time.Millisecond}
	workers := provideWorkerPool(config.Config{WorkerCount: 1})

	useCase := app.NewProcessUseCase(fakeRepo{})
	processor := inboundgrpc.NewProcessor(useCase)
	consumer := provideConsumer(cfg, queue.NewMockClient(), processor, dlq.NewProducer(), workers, retry.NewPolicy())

	lc := &fakeLifecycle{}
	runConsumer(lc, consumer, workers)

	if lc.hook.OnStart == nil || lc.hook.OnStop == nil {
		t.Fatal("expected lifecycle hooks to be registered")
	}
	if err := lc.hook.OnStart(context.Background()); err != nil {
		t.Fatalf("expected OnStart success, got %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := lc.hook.OnStop(context.Background()); err != nil {
		t.Fatalf("expected OnStop success, got %v", err)
	}
}

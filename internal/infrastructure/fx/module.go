// Package fxmodule wires the dependency injection graph using Uber Fx.
package fxmodule

import (
	"context"
	"fmt"
	"log"
	"strings"

	"telemetry-collector/internal/adapters/inbound/queue"
	"telemetry-collector/internal/adapters/outbound/dlq"
	"telemetry-collector/internal/adapters/outbound/postgres"
	app "telemetry-collector/internal/application/telemetry"
	"telemetry-collector/internal/infrastructure/config"
	"telemetry-collector/internal/infrastructure/retry"
	"telemetry-collector/internal/infrastructure/workerpool"

	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

// Module returns the root Fx option that provides all collector dependencies.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			config.Load,
			retry.NewPolicy,
			provideWorkerPool,
			provideBunDB,
			fx.Annotate(
				postgres.NewTelemetryRepository,
				fx.As(new(app.TelemetryRepository)),
			),
			app.NewProcessUseCase,
			dlq.NewProducer,
			provideQueueClient,
			queue.NewProtoProcessor,
			provideConsumer,
		),
		fx.Invoke(registerQueueClientClose),
		fx.Invoke(runConsumer),
	)
}

func provideWorkerPool(cfg config.Config) *workerpool.Pool {
	return workerpool.New(cfg.WorkerCount)
}

func provideBunDB(cfg config.Config) *bun.DB {
	return postgres.NewBunDB(cfg.PostgresDSN)
}

func provideQueueClient(cfg config.Config) (queue.Client, error) {
	backend := strings.TrimSpace(cfg.QueueBackend)

	switch {
	case strings.EqualFold(backend, "grpc"):
		if strings.TrimSpace(cfg.MQGRPCAddr) == "" {
			return nil, fmt.Errorf("QUEUE_BACKEND=grpc requires MQ_GRPC_ADDR")
		}
		c, err := queue.NewGRPCMQClient(cfg)
		if err != nil {
			return nil, err
		}
		log.Printf("queue: backend=grpc addr=%s topic=%q group=%q", cfg.MQGRPCAddr, cfg.MQTopic, cfg.MQGroup)
		return c, nil

	case strings.EqualFold(backend, "http"):
		if strings.TrimSpace(cfg.MQHTTPBase) == "" {
			log.Printf("queue: backend=http but MQ_HTTP_BASE empty; falling back to mock")
			return queue.NewMockClient(), nil
		}
		c, err := queue.NewHTTPClient(cfg)
		if err != nil {
			return nil, err
		}
		log.Printf("queue: backend=http base=%s pull=%s ack=%s", cfg.MQHTTPBase, cfg.MQHTTPPullPath, cfg.MQHTTPAckPath)
		return c, nil

	default:
		log.Printf("queue: backend=mock")
		return queue.NewMockClient(), nil
	}
}

// registerQueueClientClose ensures the gRPC/HTTP client is closed on shutdown.
func registerQueueClientClose(lc fx.Lifecycle, c queue.Client) {
	type closer interface {
		Close() error
	}
	gc, ok := c.(closer)
	if !ok {
		return
	}
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return gc.Close()
		},
	})
}

func provideConsumer(
	cfg config.Config,
	client queue.Client,
	processor *queue.ProtoProcessor,
	dlq *dlq.Producer,
	workers *workerpool.Pool,
	policy retry.Policy,
) *queue.Consumer {
	return queue.NewConsumer(client, processor, dlq, workers, cfg.QueueBatchSize, cfg.PollInterval, policy)
}

func runConsumer(lc fx.Lifecycle, c *queue.Consumer, workers *workerpool.Pool) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Println("consumer: starting telemetry consumer loop")
			runCtx, stop := context.WithCancel(context.Background())
			cancel = stop
			go c.Start(runCtx)
			return nil
		},
		OnStop: func(context.Context) error {
			log.Println("consumer: shutting down")
			if cancel != nil {
				cancel()
			}
			workers.Close()
			return nil
		},
	})
}

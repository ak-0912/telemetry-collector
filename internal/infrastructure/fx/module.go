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
		// OnStop runs in reverse order: consumer stops before queue Close().
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
	if strings.EqualFold(backend, "grpc") {
		if strings.TrimSpace(cfg.MQGRPCAddr) == "" {
			return nil, fmt.Errorf("QUEUE_BACKEND=grpc requires MQ_GRPC_ADDR")
		}
		c, err := queue.NewGRPCMQClient(cfg)
		if err != nil {
			return nil, err
		}
		log.Printf("queue client: grpc backend %s topic=%q group=%q", cfg.MQGRPCAddr, cfg.MQTopic, cfg.MQGroup)
		return c, nil
	}
	if strings.EqualFold(backend, "http") && strings.TrimSpace(cfg.MQHTTPBase) != "" {
		c, err := queue.NewHTTPClient(cfg)
		if err != nil {
			return nil, err
		}
		log.Printf("queue client: http backend %s (pull %s, ack %s)", cfg.MQHTTPBase, cfg.MQHTTPPullPath, cfg.MQHTTPAckPath)
		return c, nil
	}
	if strings.EqualFold(backend, "http") {
		log.Printf("queue client: QUEUE_BACKEND=http but MQ_HTTP_BASE empty; using mock queue")
	}
	return queue.NewMockClient(), nil
}

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
		OnStart: func(ctx context.Context) error {
			log.Println("starting telemetry consumer")
			runCtx, stop := context.WithCancel(context.Background())
			cancel = stop
			go c.Start(runCtx)
			return nil
		},
		OnStop: func(context.Context) error {
			if cancel != nil {
				cancel()
			}
			workers.Close()
			return nil
		},
	})
}

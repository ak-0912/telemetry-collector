package fxmodule

import (
	"context"
	"log"

	inboundgrpc "telemetry-collector/internal/adapters/inbound/grpc"
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
			inboundgrpc.NewProcessor,
			dlq.NewProducer,
			queue.NewMockClient,
			provideConsumer,
		),
		fx.Invoke(runConsumer),
	)
}

func provideWorkerPool(cfg config.Config) *workerpool.Pool {
	return workerpool.New(cfg.WorkerCount)
}

func provideBunDB(cfg config.Config) *bun.DB {
	return postgres.NewBunDB(cfg.PostgresDSN)
}

func provideConsumer(
	cfg config.Config,
	client *queue.MockClient,
	processor *inboundgrpc.Processor,
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

package main

import (
	"context"
	"log"
	"time"

	mqv1 "telemetry-collector/api/mq/v1"
	telemetryv1 "telemetry-collector/api/telemetry/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "host.docker.internal:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	payload, err := proto.Marshal(&telemetryv1.TelemetryMessage{
		MetricName:          "gpu.temperature",
		GpuId:               "gpu-test-1",
		Device:              "nvidia0",
		Uuid:                "11111111-1111-1111-1111-111111111111",
		ModelName:           "H100",
		HostName:            "manual-test",
		Value:               42.5,
		LabelsRaw:           `instance="manual-test:9400",job="manual"`,
		ProcessedAtUnixNano: time.Now().UnixNano(),
	})
	if err != nil {
		log.Fatal(err)
	}

	client := mqv1.NewMessageQueueServiceClient(conn)
	resp, err := client.Publish(ctx, &mqv1.PublishRequest{
		Topic:   "gpu-telemetry",
		Key:     "manual-test",
		Payload: payload,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("published partition=%d offset=%d", resp.GetPartition(), resp.GetOffset())
}

package grpc

import (
	pb "telemetry-collector/api/telemetry/v1"

	"google.golang.org/protobuf/proto"
)

// tryProtoTelemetryMessage decodes a protobuf wire-encoded TelemetryMessage.
func tryProtoTelemetryMessage(payload []byte) (*pb.TelemetryMessage, bool) {
	var msg pb.TelemetryMessage
	if err := proto.Unmarshal(payload, &msg); err != nil {
		return nil, false
	}
	if msg.GetMetricName() == "" {
		return nil, false
	}
	return &msg, true
}

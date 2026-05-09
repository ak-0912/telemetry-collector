package grpc

import (
	"testing"
)

func TestParseTelemetryPayload_JSONObject(t *testing.T) {
	raw := []byte(`{"metric_name":"gpu.temperature","gpu_id":"g1","device":"d0","uuid":"550e8400-e29b-41d4-a716-446655440000","model_name":"A100","host_name":"h1","value":1.5,"labels_raw":"{}","processed_at_unix_nano":99}`)
	msg, err := parseTelemetryPayload(raw)
	if err != nil || msg.GetMetricName() != "gpu.temperature" || msg.GetValue() != 1.5 {
		t.Fatalf("got %+v err=%v", msg, err)
	}
}

func TestParseTelemetryPayload_UTF8BOM(t *testing.T) {
	raw := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"metric_name":"gpu.temperature","gpu_id":"g1","device":"d0","uuid":"550e8400-e29b-41d4-a716-446655440000","model_name":"A100","host_name":"h1","value":2,"labels_raw":"{}","processed_at_unix_nano":1}`)...)
	msg, err := parseTelemetryPayload(raw)
	if err != nil || msg.GetMetricName() != "gpu.temperature" {
		t.Fatalf("got %+v err=%v", msg, err)
	}
}

func TestParseTelemetryPayload_JSONArray(t *testing.T) {
	raw := []byte(`[{"metric_name":"gpu.temperature","gpu_id":"g1","device":"d0","uuid":"550e8400-e29b-41d4-a716-446655440000","model_name":"A100","host_name":"h1","value":3,"labels_raw":"{}","processed_at_unix_nano":2}]`)
	msg, err := parseTelemetryPayload(raw)
	if err != nil || msg.GetMetricName() != "gpu.temperature" {
		t.Fatalf("got %+v err=%v", msg, err)
	}
}

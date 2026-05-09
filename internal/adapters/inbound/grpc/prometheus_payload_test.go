package grpc

import (
	"strconv"
	"testing"
)

func TestParsePrometheusSample_instanceLabel(t *testing.T) {
	line := `DCGM_FI_DEV_GPU_TEMP{instance="mtv5-dgx1-hgpu-031:9400",gpu="0",UUID="GPU-aaa"} 55.0 1710000000000`
	msg, err := parsePrometheusSample([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected message")
	}
	if msg.GetMetricName() != "DCGM_FI_DEV_GPU_TEMP" || msg.GetValue() != 55 {
		t.Fatalf("metric/value: %+v", msg)
	}
	if msg.GetHostName() != "mtv5-dgx1-hgpu-031" {
		t.Fatalf("host from instance: got %q", msg.GetHostName())
	}
	if msg.GetGpuId() != "0" {
		t.Fatalf("gpu from label: got %q", msg.GetGpuId())
	}
	if msg.GetUuid() != "GPU-aaa" {
		t.Fatalf("uuid: got %q", msg.GetUuid())
	}
	if msg.GetProcessedAtUnixNano() != 1710000000000*1_000_000 {
		t.Fatalf("ts nano: got %d", msg.GetProcessedAtUnixNano())
	}
}

func TestParsePrometheusSample_noLabels(t *testing.T) {
	msg, err := parsePrometheusSample([]byte("up 1"))
	if err != nil || msg == nil {
		t.Fatalf("msg=%v err=%v", msg, err)
	}
	if msg.GetMetricName() != "up" || msg.GetValue() != 1 {
		t.Fatalf("%+v", msg)
	}
}

func TestParseTelemetryPayload_PrometheusBeforeCSV(t *testing.T) {
	line := `metric_x{instance="host:1234"} 3.25`
	msg, err := parseTelemetryPayload([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if msg.GetMetricName() != "metric_x" || msg.GetValue() != 3.25 || msg.GetHostName() != "host" {
		t.Fatalf("%+v", msg)
	}
}

func TestPrometheusSampleText_multilineValue(t *testing.T) {
	block := "DCGM_FI_DEV_GPU_TEMP{instance=\"h:9400\"}\n55"
	line := prometheusSampleText([]byte(block))
	want := `DCGM_FI_DEV_GPU_TEMP{instance="h:9400"} 55`
	if line != want {
		t.Fatalf("got %q want %q", line, want)
	}
	msg, err := parsePrometheusSample([]byte(block))
	if err != nil || msg == nil || msg.GetValue() != 55 {
		t.Fatalf("msg=%v err=%v", msg, err)
	}
}

func TestParseTelemetryPayload_JSONEnvelopePayload(t *testing.T) {
	inner := `DCGM_FI_DEV_GPU_TEMP{instance="n:1"} 3`
	wrapped := `{"payload":` + strconv.Quote(inner) + `}`
	msg, err := parseTelemetryPayload([]byte(wrapped))
	if err != nil {
		t.Fatal(err)
	}
	if msg.GetMetricName() != "DCGM_FI_DEV_GPU_TEMP" || msg.GetValue() != 3 {
		t.Fatalf("%+v", msg)
	}
}

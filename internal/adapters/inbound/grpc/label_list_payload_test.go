package grpc

import "testing"

func TestParseLabelListTelemetry_DCGMExporterLine(t *testing.T) {
	line := `DCGM_FI_DRIVER_VERSION="535.129.03",Hostname="mtv5-dgx1-hgpu-006",UUID="GPU-1df3e374-2435-dc0a-53d1-52a93108d331",__name__="DCGM_FI_DEV_MEM_COPY_UTIL",device="nvidia1",gpu="1",instance="mtv5-dgx1-hgpu-006:9400",job="dgx_dcgm_exporter",modelName="NVIDIA H100 80GB HBM3"`
	msg, err := parseLabelListTelemetry([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected message")
	}
	if msg.GetMetricName() != "DCGM_FI_DEV_MEM_COPY_UTIL" {
		t.Fatalf("metric %q", msg.GetMetricName())
	}
	if msg.GetGpuId() != "1" || msg.GetUuid() != "GPU-1df3e374-2435-dc0a-53d1-52a93108d331" {
		t.Fatalf("gpu/uuid: %+v", msg)
	}
	if msg.GetHostName() != "mtv5-dgx1-hgpu-006" {
		t.Fatalf("host %q", msg.GetHostName())
	}
	if msg.GetModelName() != "NVIDIA H100 80GB HBM3" {
		t.Fatalf("model %q", msg.GetModelName())
	}
	if msg.GetDevice() != "nvidia1" {
		t.Fatalf("device %q", msg.GetDevice())
	}
}

func TestParseTelemetryPayload_labelList(t *testing.T) {
	line := `UUID="GPU-x",__name__="m",gpu="0",Hostname="h1"`
	msg, err := parseTelemetryPayload([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if msg.GetMetricName() != "m" || msg.GetGpuId() != "0" || msg.GetHostName() != "h1" {
		t.Fatalf("%+v", msg)
	}
}

func TestParseTelemetryPayload_labelListWithWhitespaceAroundName(t *testing.T) {
	line := `Hostname="h1", UUID="GPU-x", __name__ = "m", gpu="0", instance="h1:9400"`
	msg, err := parseTelemetryPayload([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected message")
	}
	if msg.GetMetricName() != "m" {
		t.Fatalf("metric %q", msg.GetMetricName())
	}
	if msg.GetHostName() != "h1" {
		t.Fatalf("host %q", msg.GetHostName())
	}
}

func TestParseTelemetryPayload_labelListWithMetricNameKey(t *testing.T) {
	line := `Hostname="h1", UUID="GPU-x", metric_name = "m", gpu="0", instance="h1:9400"`
	msg, err := parseTelemetryPayload([]byte(line))
	if err != nil {
		t.Fatal(err)
	}
	if msg == nil {
		t.Fatal("expected message")
	}
	if msg.GetMetricName() != "m" {
		t.Fatalf("metric %q", msg.GetMetricName())
	}
	if msg.GetHostName() != "h1" {
		t.Fatalf("host %q", msg.GetHostName())
	}
}

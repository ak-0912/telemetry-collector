package telemetry

// Input is the application-layer DTO for a single telemetry data point.
// It decouples inbound adapter formats (protobuf, CSV, Prometheus) from the
// domain model.
type Input struct {
	MetricName          string
	GPUID               string
	Device              string
	UUID                string
	ModelName           string
	HostName            string
	Value               float64
	LabelsRaw           string
	ProcessedAtUnixNano int64
}

// Package telemetry defines the core domain model for GPU telemetry data.
package telemetry

// Telemetry is the core domain aggregate representing a single GPU telemetry
// data point collected from a DCGM/Prometheus exporter.
type Telemetry struct {
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

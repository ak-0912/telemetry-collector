package telemetry

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

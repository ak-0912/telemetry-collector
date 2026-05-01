package telemetry

import "time"

type Input struct {
	GPUID          string
	HostID         string
	Timestamp      time.Time
	GPUUtilization float64
	MemoryUsedMB   int64
	TemperatureC   float64
}

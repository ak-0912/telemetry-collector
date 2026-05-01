package telemetry

import "time"

// Telemetry is the core domain aggregate.
type Telemetry struct {
	GPUID          string
	HostID         string
	Timestamp      time.Time
	GPUUtilization float64
	MemoryUsedMB   int64
	TemperatureC   float64
}

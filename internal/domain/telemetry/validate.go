package telemetry

import (
	"errors"
	"fmt"
)

func (t Telemetry) Validate() error {
	if t.GPUID == "" {
		return fmt.Errorf("%w: gpu_id is required", ErrValidation)
	}
	if t.Timestamp.IsZero() {
		return fmt.Errorf("%w: timestamp is required", ErrValidation)
	}
	if t.GPUUtilization < 0 || t.GPUUtilization > 100 {
		return fmt.Errorf("%w: gpu_utilization must be in range [0, 100]", ErrValidation)
	}
	if t.MemoryUsedMB < 0 {
		return fmt.Errorf("%w: memory_used_mb must be >= 0", ErrValidation)
	}
	if t.TemperatureC < -273.15 {
		return fmt.Errorf("%w: temperature_c is out of range", ErrValidation)
	}
	return nil
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

func IsTransientError(err error) bool {
	return errors.Is(err, ErrTransient)
}

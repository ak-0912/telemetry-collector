package telemetry

import (
	"errors"
	"fmt"
)

// Validate enforces domain invariants on a Telemetry aggregate.
// Returns an ErrValidation-wrapped error when any required field is missing.
func (t Telemetry) Validate() error {
	if t.MetricName == "" {
		return fmt.Errorf("%w: metric_name is required", ErrValidation)
	}
	if t.GPUID == "" {
		return fmt.Errorf("%w: gpu_id is required", ErrValidation)
	}
	if t.UUID == "" {
		return fmt.Errorf("%w: uuid is required", ErrValidation)
	}
	if t.HostName == "" {
		return fmt.Errorf("%w: host_name is required", ErrValidation)
	}
	if t.ProcessedAtUnixNano <= 0 {
		return fmt.Errorf("%w: processed_at_unix_nano must be > 0", ErrValidation)
	}
	return nil
}

// IsValidationError reports whether err (or any error in its chain) is a
// domain validation failure.
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsTransientError reports whether err (or any error in its chain) is a
// transient/retryable failure.
func IsTransientError(err error) bool {
	return errors.Is(err, ErrTransient)
}

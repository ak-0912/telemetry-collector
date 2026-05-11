package telemetry

import "errors"

var (
	// ErrValidation indicates a payload that fails business-rule checks and
	// should be routed to the dead-letter queue (not retried).
	ErrValidation = errors.New("validation error")

	// ErrTransient indicates a temporary failure (e.g. database timeout) that
	// is safe to retry with exponential back-off.
	ErrTransient = errors.New("transient error")

	// ErrSystem indicates an unrecoverable infrastructure fault.
	ErrSystem = errors.New("system error")
)

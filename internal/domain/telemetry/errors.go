package telemetry

import "errors"

var (
	ErrValidation = errors.New("validation error")
	ErrTransient  = errors.New("transient error")
	ErrSystem     = errors.New("system error")
)

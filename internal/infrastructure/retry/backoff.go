// Package retry provides an exponential back-off policy for transient failures.
package retry

import "time"

// Policy defines the parameters for exponential back-off.
type Policy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// NewPolicy returns the default retry policy (2 s base, 60 s cap).
func NewPolicy() Policy {
	return Policy{
		BaseDelay: 2 * time.Second,
		MaxDelay:  60 * time.Second,
	}
}

// NextDelay returns the back-off duration for the given attempt number
// (1-indexed). The delay doubles on each attempt, capped at MaxDelay.
func (p Policy) NextDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return p.BaseDelay
	}
	d := p.BaseDelay * time.Duration(1<<(attempt-1))
	if d > p.MaxDelay {
		return p.MaxDelay
	}
	return d
}

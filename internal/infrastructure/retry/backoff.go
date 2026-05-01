package retry

import "time"

type Policy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func NewPolicy() Policy {
	return Policy{
		BaseDelay: 2 * time.Second,
		MaxDelay:  60 * time.Second,
	}
}

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

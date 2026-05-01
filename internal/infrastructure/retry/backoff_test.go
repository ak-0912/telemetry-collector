package retry

import (
	"testing"
	"time"
)

func TestNewPolicyDefaults(t *testing.T) {
	p := NewPolicy()
	if p.BaseDelay != 2*time.Second || p.MaxDelay != 60*time.Second {
		t.Fatalf("unexpected defaults: %+v", p)
	}
}

func TestPolicyNextDelay(t *testing.T) {
	p := Policy{BaseDelay: time.Second, MaxDelay: 5 * time.Second}
	if got := p.NextDelay(1); got != time.Second {
		t.Fatalf("attempt 1 delay mismatch: %v", got)
	}
	if got := p.NextDelay(3); got != 4*time.Second {
		t.Fatalf("attempt 3 delay mismatch: %v", got)
	}
	if got := p.NextDelay(10); got != 5*time.Second {
		t.Fatalf("max cap mismatch: %v", got)
	}
}

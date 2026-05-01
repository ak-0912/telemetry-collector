package logging

import "testing"

func TestNewLogger(t *testing.T) {
	l := New()
	if l == nil {
		t.Fatal("expected logger instance")
	}
}

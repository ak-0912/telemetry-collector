package workerpool

import (
	"sync/atomic"
	"testing"
)

func TestPoolExecutesTasks(t *testing.T) {
	p := New(2)

	var counter int64
	for i := 0; i < 10; i++ {
		p.Submit(func() {
			atomic.AddInt64(&counter, 1)
		})
	}
	p.Close()

	if got := atomic.LoadInt64(&counter); got != 10 {
		t.Fatalf("expected 10 tasks executed, got %d", got)
	}
}

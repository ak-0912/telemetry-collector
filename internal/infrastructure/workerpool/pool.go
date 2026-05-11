// Package workerpool provides a fixed-size goroutine pool for concurrent
// message processing.
package workerpool

import "sync"

// Pool is a bounded worker pool backed by a buffered channel of tasks.
type Pool struct {
	tasks chan func()
	wg    sync.WaitGroup
}

// New spawns workerCount goroutines that pull tasks from a shared channel.
// The channel is buffered at 4× the worker count to absorb short bursts.
func New(workerCount int) *Pool {
	if workerCount <= 0 {
		workerCount = 1
	}
	p := &Pool{
		tasks: make(chan func(), workerCount*4),
	}
	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for task := range p.tasks {
				task()
			}
		}()
	}
	return p
}

// Submit enqueues a task for execution. It blocks if all workers are busy and
// the internal buffer is full, providing natural back-pressure.
func (p *Pool) Submit(task func()) {
	p.tasks <- task
}

// Close signals all workers to drain remaining tasks and waits for completion.
func (p *Pool) Close() {
	close(p.tasks)
	p.wg.Wait()
}

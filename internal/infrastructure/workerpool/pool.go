package workerpool

import "sync"

type Pool struct {
	tasks chan func()
	wg    sync.WaitGroup
}

func New(workerCount int) *Pool {
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

func (p *Pool) Submit(task func()) {
	p.tasks <- task
}

func (p *Pool) Close() {
	close(p.tasks)
	p.wg.Wait()
}

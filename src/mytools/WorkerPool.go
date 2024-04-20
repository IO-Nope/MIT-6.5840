package mytools

import (
	"sync"
)

type Job func()

type WorkerPool struct {
	jobs chan Job
	wg   sync.WaitGroup
}

// 创建一个具有指定容量的 channel
func NewWorkerPool(size int) *WorkerPool {
	p := &WorkerPool{
		jobs: make(chan Job, size), // 创建一个具有指定容量的 channel
	}

	p.wg.Add(size)
	for i := 0; i < size; i++ {
		go p.worker()
	}

	return p
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for job := range p.jobs {
		job()
	}
}

func (p *WorkerPool) Submit(job Job) {
	p.jobs <- job
}

func (p *WorkerPool) Shutdown() {
	close(p.jobs)
	p.wg.Wait()
}

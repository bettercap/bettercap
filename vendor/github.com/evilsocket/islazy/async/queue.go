package async

import (
	"runtime"
	"sync"
)

// Job is the generic interface object representing the data being
// pushed to the jobs queue and being passed to the workers.
type Job interface{}

// Logic is the implementation of the logic each worker will execute.
type Logic func(arg Job)

// WorkQueue is the object representing an async jobs queue with
// a given number of concurrent workers.
type WorkQueue struct {
	workers  int
	jobChan  chan Job
	stopChan chan struct{}
	jobs     sync.WaitGroup
	done     sync.WaitGroup
	logic    Logic
}

// NewQueue creates a new job queue with a specific worker logic.
// If workers is greater or equal than zero, it will be auto
// scaled to the number of logical CPUs usable by the current
// process.
func NewQueue(workers int, logic Logic) *WorkQueue {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	wq := &WorkQueue{
		workers:  workers,
		jobChan:  make(chan Job),
		stopChan: make(chan struct{}),
		jobs:     sync.WaitGroup{},
		done:     sync.WaitGroup{},
		logic:    logic,
	}

	for i := 0; i < workers; i++ {
		wq.done.Add(1)
		go wq.worker(i)
	}

	return wq
}

func (wq *WorkQueue) worker(id int) {
	defer wq.done.Done()
	for {
		select {
		case <-wq.stopChan:
			return

		case job := <-wq.jobChan:
			wq.logic(job)
			wq.jobs.Done()
		}
	}
}

// Add pushes a new job to the queue.
func (wq *WorkQueue) Add(job Job) {
	wq.jobs.Add(1)
	wq.jobChan <- job
}

// Wait stops until all the workers stopped.
func (wq *WorkQueue) Wait() {
	wq.done.Wait()
}

// WaitDone stops until all jobs on the queue have been processed.
func (wq *WorkQueue) WaitDone() {
	wq.jobs.Wait()
}

// Stop stops the job queue and the workers.
func (wq *WorkQueue) Stop() {
	close(wq.stopChan)
	wq.jobs.Wait()
	wq.done.Wait()
	close(wq.jobChan)
}

package concurrency

import (
	"context"
	"runtime"
)

// Semaphore provides controlled concurrency
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a new semaphore with the specified capacity
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		ch: make(chan struct{}, capacity),
	}
}

// NewDefaultSemaphore creates a semaphore with sensible defaults
func NewDefaultSemaphore() *Semaphore {
	// Use a conservative approach: min(NumCPU * 2, 10)
	capacity := min(runtime.NumCPU()*2, 10)
	return NewSemaphore(capacity)
}

// NewAPIClientSemaphore creates a semaphore optimized for API clients
func NewAPIClientSemaphore() *Semaphore {
	// For GitHub API, be more conservative to avoid rate limiting
	capacity := min(runtime.NumCPU(), 5)
	return NewSemaphore(capacity)
}

// Acquire acquires a slot in the semaphore, blocking if necessary
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire attempts to acquire a slot without blocking
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release releases a slot in the semaphore
func (s *Semaphore) Release() {
	<-s.ch
}

// Capacity returns the total capacity of the semaphore
func (s *Semaphore) Capacity() int {
	return cap(s.ch)
}

// Available returns the number of available slots
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}

// WorkerPool provides a controlled worker pool for concurrent tasks
type WorkerPool struct {
	semaphore *Semaphore
	workers   int
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		semaphore: NewSemaphore(workers),
		workers:   workers,
	}
}

// NewDefaultWorkerPool creates a worker pool with sensible defaults
func NewDefaultWorkerPool() *WorkerPool {
	workers := min(runtime.NumCPU()*2, 10)
	return NewWorkerPool(workers)
}

// Submit submits a task to the worker pool
func (p *WorkerPool) Submit(ctx context.Context, task func() error) error {
	if err := p.semaphore.Acquire(ctx); err != nil {
		return err
	}
	
	go func() {
		defer p.semaphore.Release()
		_ = task()
	}()
	
	return nil
}

// Wait waits for all workers to become available (tasks to complete)
func (p *WorkerPool) Wait(ctx context.Context) error {
	// Acquire all slots to ensure all workers are idle
	for i := 0; i < p.workers; i++ {
		if err := p.semaphore.Acquire(ctx); err != nil {
			// Release acquired slots and return error
			for j := 0; j < i; j++ {
				p.semaphore.Release()
			}
			return err
		}
	}
	
	// Release all slots
	for i := 0; i < p.workers; i++ {
		p.semaphore.Release()
	}
	
	return nil
}

// Workers returns the number of workers in the pool
func (p *WorkerPool) Workers() int {
	return p.workers
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
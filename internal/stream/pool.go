package stream

import (
	"context"
	"sync"
)

// Pool represents a worker pool that processes tasks concurrently.
type Pool struct {
	size      int                   // Number of worker goroutines
	processor func(Task) TaskResult // Function used to process each Task
}

// NewPool creates a new Pool with the given size and task processor.
func NewPool(size int, processor func(Task) TaskResult) *Pool {
	return &Pool{
		size:      size,
		processor: processor,
	}
}

// Process starts the worker pool to process tasks from the input channel.
// Each task is processed using the processor function and the result is sent to the results channel.
// It blocks until all tasks are processed or the context is cancelled.
func (p *Pool) Process(ctx context.Context, tasks <-chan Task, results chan<- TaskResult) error {
	var wg sync.WaitGroup

	// Start 'size' number of worker goroutines
	for i := 0; i < p.size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.pool(ctx, tasks, results)
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	return nil
}

// pool defines the logic each worker runs: continuously read tasks and process them.
func (p *Pool) pool(ctx context.Context, tasks <-chan Task, results chan<- TaskResult) {
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				// Input channel closed: no more tasks
				return
			}

			// Process the task
			result := p.processor(task)

			// Send the result unless the context is cancelled
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			// Exit early if context is cancelled
			return
		}
	}
}

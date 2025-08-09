package streaming

import (
	"context"
	"sync"

	"github.com/hambosto/hexwarden/internal/constants"
)

// Pool represents a worker pool that processes tasks concurrently
type Pool struct {
	size      int                                       // Number of worker goroutines
	processor func(constants.Task) constants.TaskResult // Function used to process each Task
}

// NewPool creates a new Pool with the given size and task processor
func NewPool(size int, processor func(constants.Task) constants.TaskResult) *Pool {
	if size <= 0 {
		size = constants.MaxConcurrency
	}

	return &Pool{
		size:      size,
		processor: processor,
	}
}

// Process starts the worker pool to process tasks from the input channel
// Each task is processed using the processor function and the result is sent to the results channel
// It blocks until all tasks are processed or the context is cancelled
func (p *Pool) Process(ctx context.Context, tasks <-chan constants.Task, results chan<- constants.TaskResult) error {
	var wg sync.WaitGroup

	// Start 'size' number of worker goroutines
	for i := 0; i < p.size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.worker(ctx, tasks, results)
		}()
	}

	// Wait for all workers to finish
	wg.Wait()
	return nil
}

// worker defines the logic each worker runs: continuously read tasks and process them
func (p *Pool) worker(ctx context.Context, tasks <-chan constants.Task, results chan<- constants.TaskResult) {
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

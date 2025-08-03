package worker

import (
	"context"
	"sync"
)

type WorkerPool struct {
	size      int
	processor func(Task) TaskResult
}

func NewWorkerPool(size int, processor func(Task) TaskResult) *WorkerPool {
	return &WorkerPool{
		size:      size,
		processor: processor,
	}
}

func (wp *WorkerPool) Process(ctx context.Context, tasks <-chan Task, results chan<- TaskResult) error {
	var wg sync.WaitGroup

	for i := 0; i < wp.size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wp.worker(ctx, tasks, results)
		}()
	}

	wg.Wait()
	return nil
}

func (wp *WorkerPool) worker(ctx context.Context, tasks <-chan Task, results chan<- TaskResult) {
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				return // Channel closed
			}

			result := wp.processor(task)

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

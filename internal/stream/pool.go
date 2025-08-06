package stream

import (
	"context"
	"sync"
)

type Pool struct {
	size      int
	processor func(Task) TaskResult
}

func NewPool(size int, processor func(Task) TaskResult) *Pool {
	return &Pool{
		size:      size,
		processor: processor,
	}
}

func (p *Pool) Process(ctx context.Context, tasks <-chan Task, results chan<- TaskResult) error {
	var wg sync.WaitGroup

	for i := 0; i < p.size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.pool(ctx, tasks, results)
		}()
	}

	wg.Wait()
	return nil
}

func (p *Pool) pool(ctx context.Context, tasks <-chan Task, results chan<- TaskResult) {
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				return // Channel closed
			}

			result := p.processor(task)

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

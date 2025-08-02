package worker

import (
	"fmt"
	"io"
	"sync"
)

func (w *Worker) runPipeline(input io.Reader, output io.Writer) error {
	// Initialize channels with buffering for better throughput
	w.taskChan = make(chan Task, w.config.QueueSize)
	w.resultChan = make(chan TaskResult, w.config.QueueSize)

	errChan := make(chan error, 3) // Reader, workers, writer can each send one error

	var wg sync.WaitGroup

	// Start reader goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(w.taskChan)

		if err := w.readTasks(input); err != nil {
			select {
			case errChan <- fmt.Errorf("reader error: %w", err):
			case <-w.ctx.Done():
			}
		}
	}()

	// Start worker pool
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(w.resultChan)

		if err := w.workerPool.Process(w.ctx, w.taskChan, w.resultChan); err != nil {
			select {
			case errChan <- fmt.Errorf("worker error: %w", err):
			case <-w.ctx.Done():
			}
		}
	}()

	// Start writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := w.writeResults(output); err != nil {
			select {
			case errChan <- fmt.Errorf("writer error: %w", err):
			case <-w.ctx.Done():
			}
		}
	}()

	// Wait for completion or error
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errChan:
		w.cancel() // Cancel other goroutines
		wg.Wait()  // Wait for cleanup
		return err
	case <-done:
		return nil
	case <-w.ctx.Done():
		wg.Wait()
		return ErrCanceled
	}
}

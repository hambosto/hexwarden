package worker

import (
	"fmt"
	"io"
	"sync"
)

func (w *Worker) runPipeline(r io.Reader, wr io.Writer) error {
	tasks := make(chan Task, w.concurrency)
	results := make(chan TaskResult, w.concurrency)
	errChan := make(chan error, 1)

	var (
		worker sync.WaitGroup
		writer sync.WaitGroup
	)

	worker.Add(w.concurrency)
	for range w.concurrency {
		go w.worker(tasks, results, &worker)
	}

	writer.Add(1)
	go w.writeResults(wr, results, &writer, errChan)

	if err := w.setTasks(r, tasks); err != nil {
		close(tasks)  // Signal workers that no more tasks will be sent.
		worker.Wait() // Wait for all workers to finish processing any in-flight tasks and exit.

		close(results) // Signal the writer that no more results will be sent.
		writer.Wait()  // Wait for the writer to finish processing any in-flight results and exit.

		return fmt.Errorf("task distribution error: %v", err)
	}

	return w.waitForComplete(tasks, results, &worker, &writer, errChan)
}

func (w *Worker) setTasks(
	r io.Reader,
	tasks chan<- Task,
) error {
	if w.processing == Encryption {
		return w.readForEncryption(r, tasks)
	}
	return w.readForDecryption(r, tasks)
}

func (w *Worker) worker(
	tasks <-chan Task,
	results chan<- TaskResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	w.processTasks(tasks, results)
}

func (w *Worker) processTasks(tasks <-chan Task, results chan<- TaskResult) {
	for t := range tasks {
		var (
			output []byte
			err    error
		)

		if w.processing == Encryption {
			output, err = w.processor.Encrypt(t.Data)
		} else {
			output, err = w.processor.Decrypt(t.Data)
		}

		size := len(t.Data)
		if w.processing != Encryption {
			size = len(output)
		}

		results <- TaskResult{
			Index: t.Index,
			Data:  output,
			Size:  size,
			Err:   err,
		}
	}
}

func (w *Worker) waitForComplete(
	tasks chan Task,
	results chan TaskResult,
	worker *sync.WaitGroup,
	writer *sync.WaitGroup,
	errChan chan error,
) error {
	close(tasks)
	worker.Wait()
	close(results)
	writer.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

package worker

import (
	"fmt"
	"io"
	"sync"

	"github.com/hambosto/hexwarden/internal/ui"
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
		return fmt.Errorf("task distribution error: %v", err)
	}

	return w.waitForComplete(tasks, results, &worker, &writer, errChan)
}

func (w *Worker) setTasks(
	r io.Reader,
	tasks chan<- Task,
) error {
	if w.processor.ProcessingMode == ui.ModeEncrypt {
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
		output, err := w.processor.ProcessChunk(t.Data)
		size := len(t.Data)
		if w.processor.ProcessingMode != ui.ModeEncrypt {
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

package worker

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/hambosto/hexwarden/internal/ui"
)

func (w *Worker) writeResults(writer io.Writer, results <-chan TaskResult, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	pending := make(map[uint32]TaskResult)
	var nextIndex uint32

	for res := range results {
		if res.Err != nil {
			errChan <- fmt.Errorf("processing chunk %d: %w", res.Index, res.Err)
			return
		}

		pending[res.Index] = res

		for {
			current, exists := pending[nextIndex]
			if !exists {
				break
			}

			if err := w.writeChunk(writer, current); err != nil {
				errChan <- fmt.Errorf("writing chunk %d: %w", nextIndex, err)
				return
			}

			delete(pending, nextIndex)
			nextIndex++
		}
	}
}

func (w *Worker) writeChunk(writer io.Writer, res TaskResult) error {
	if w.processor.ProcessingMode == ui.ModeEncrypt {
		if err := w.writeChunkSize(writer, len(res.Data)); err != nil {
			return err
		}
	}

	if _, err := writer.Write(res.Data); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	if err := w.progress.Add(res.Size); err != nil {
		return fmt.Errorf("progress update failed: %w", err)
	}

	return nil
}

func (w *Worker) writeChunkSize(writer io.Writer, size int) error {
	if size < 0 || size > math.MaxUint32 {
		return fmt.Errorf("chunk size out of range: %d", size)
	}

	var buffer [4]byte
	binary.BigEndian.PutUint32(buffer[:], uint32(size))

	if _, err := writer.Write(buffer[:]); err != nil {
		return fmt.Errorf("chunk size write failed: %w", err)
	}

	return nil
}

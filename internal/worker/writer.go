package worker

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"slices"
)

type resultBuffer struct {
	results map[uint64]TaskResult
	next    uint64
}

func newResultBuffer() *resultBuffer {
	return &resultBuffer{
		results: make(map[uint64]TaskResult),
		next:    0,
	}
}

func (rb *resultBuffer) add(result TaskResult) []TaskResult {
	rb.results[result.Index] = result

	var ready []TaskResult
	for {
		if result, exists := rb.results[rb.next]; exists {
			ready = append(ready, result)
			delete(rb.results, rb.next)
			rb.next++
		} else {
			break
		}
	}

	return ready
}

func (rb *resultBuffer) flush() []TaskResult {
	if len(rb.results) == 0 {
		return nil
	}

	// Get all remaining results sorted by index
	indices := make([]uint64, 0, len(rb.results))
	for idx := range rb.results {
		indices = append(indices, idx)
	}
	slices.Sort(indices)

	results := make([]TaskResult, len(indices))
	for i, idx := range indices {
		results[i] = rb.results[idx]
	}

	return results
}

func (w *Worker) writeResults(writer io.Writer) error {
	buffer := newResultBuffer()

	for {
		select {
		case result, ok := <-w.resultChan:
			if !ok {
				// Channel closed, flush remaining results
				remaining := buffer.flush()
				for _, res := range remaining {
					if err := w.writeResult(writer, res); err != nil {
						return err
					}
				}
				return nil
			}

			if result.Err != nil {
				return fmt.Errorf("processing chunk %d: %w", result.Index, result.Err)
			}

			// Add result to buffer and write any ready results
			ready := buffer.add(result)
			for _, res := range ready {
				if err := w.writeResult(writer, res); err != nil {
					return err
				}
			}

		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
}

func (w *Worker) writeResult(writer io.Writer, result TaskResult) error {
	if w.config.Processing == Encryption {
		if err := w.writeChunkSize(writer, len(result.Data)); err != nil {
			return fmt.Errorf("writing chunk size: %w", err)
		}
	}

	if _, err := writer.Write(result.Data); err != nil {
		return fmt.Errorf("writing chunk data: %w", err)
	}

	if w.bar != nil {
		if err := w.bar.Add(int64(result.Size)); err != nil {
			return fmt.Errorf("updating progress: %w", err)
		}
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

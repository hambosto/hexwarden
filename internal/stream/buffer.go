package stream

import "slices"

type Buffer struct {
	results map[uint64]TaskResult
	next    uint64
}

func NewBuffer() *Buffer {
	return &Buffer{
		results: make(map[uint64]TaskResult),
		next:    0,
	}
}

func (b *Buffer) add(result TaskResult) []TaskResult {
	b.results[result.Index] = result

	var ready []TaskResult
	for {
		if result, exists := b.results[b.next]; exists {
			ready = append(ready, result)
			delete(b.results, b.next)
			b.next++
		} else {
			break
		}
	}

	return ready
}

func (b *Buffer) flush() []TaskResult {
	if len(b.results) == 0 {
		return nil
	}

	// Get all remaining results sorted by index
	indices := make([]uint64, 0, len(b.results))
	for idx := range b.results {
		indices = append(indices, idx)
	}
	slices.Sort(indices)

	results := make([]TaskResult, len(indices))
	for i, idx := range indices {
		results[i] = b.results[idx]
	}

	return results
}

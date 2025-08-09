package streaming

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure"
)

// StreamProcessor handles concurrent encryption/decryption streaming
type StreamProcessor struct {
	processor *infrastructure.Processor
	config    StreamConfig
	pool      *Pool

	// Channels for task processing pipeline
	taskChan   chan constants.Task
	resultChan chan constants.TaskResult

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// StreamConfig holds stream processing configuration
type StreamConfig struct {
	Key         []byte
	Processing  constants.Processing
	Concurrency int
	QueueSize   int
	ChunkSize   int
}

// NewStreamProcessor creates a new stream processor instance
func NewStreamProcessor(config StreamConfig) (*StreamProcessor, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	processor, err := infrastructure.NewProcessor(config.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	config.ApplyDefaults()
	ctx, cancel := context.WithCancel(context.Background())

	s := &StreamProcessor{
		processor: processor,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
	}

	s.pool = NewPool(config.Concurrency, s.processTask)
	return s, nil
}

// Validate validates the stream configuration
func (c *StreamConfig) Validate() error {
	if len(c.Key) != constants.KeySize {
		return constants.ErrInvalidKey
	}
	return nil
}

// ApplyDefaults applies default values to unset configuration fields
func (c *StreamConfig) ApplyDefaults() {
	if c.Concurrency <= 0 {
		c.Concurrency = runtime.NumCPU()
	}
	if c.QueueSize <= 0 {
		c.QueueSize = constants.QueueSize
	}
	if c.ChunkSize <= 0 {
		c.ChunkSize = constants.DefaultChunkSize
	}
}

// Cancel cancels the stream processing
func (s *StreamProcessor) Cancel() {
	s.cancel()
}

// Process processes data from input to output with progress tracking
func (s *StreamProcessor) Process(input io.Reader, output io.Writer, totalSize int64, progressCallback func(int64)) error {
	if input == nil || output == nil {
		return constants.ErrNilStream
	}

	return s.runPipeline(input, output, totalSize, progressCallback)
}

// processTask processes a single task based on the operation type
func (s *StreamProcessor) processTask(task constants.Task) constants.TaskResult {
	var output []byte
	var err error

	switch s.config.Processing {
	case constants.Encryption:
		output, err = s.processor.Encrypt(task.Data)
	case constants.Decryption:
		output, err = s.processor.Decrypt(task.Data)
	default:
		err = fmt.Errorf("unknown processing type: %d", s.config.Processing)
	}

	// Calculate size for progress tracking
	size := s.calculateProgressSize(task.Data, output)

	return constants.TaskResult{
		Index: task.Index,
		Data:  output,
		Size:  size,
		Err:   err,
	}
}

// calculateProgressSize determines the size to use for progress tracking
func (s *StreamProcessor) calculateProgressSize(input, output []byte) int {
	if s.config.Processing == constants.Encryption {
		return len(input) // Track input size for encryption
	}
	return len(output) // Track output size for decryption
}

// runPipeline orchestrates the concurrent processing pipeline
func (s *StreamProcessor) runPipeline(input io.Reader, output io.Writer, totalSize int64, progressCallback func(int64)) error {
	// Initialize buffered channels for better throughput
	s.taskChan = make(chan constants.Task, s.config.QueueSize)
	s.resultChan = make(chan constants.TaskResult, s.config.QueueSize)

	pipeline := &pipeline{
		stream:           s,
		input:            input,
		output:           output,
		totalSize:        totalSize,
		progressCallback: progressCallback,
		errChan:          make(chan error, 3), // Reader, workers, writer
	}

	return pipeline.run()
}

// pipeline manages the concurrent processing stages
type pipeline struct {
	stream           *StreamProcessor
	input            io.Reader
	output           io.Writer
	totalSize        int64
	progressCallback func(int64)
	errChan          chan error
}

// run executes the pipeline stages concurrently
func (p *pipeline) run() error {
	var wg sync.WaitGroup

	// Start all pipeline stages
	wg.Add(1)
	go p.runReader(&wg)

	wg.Add(1)
	go p.runWorkers(&wg)

	wg.Add(1)
	go p.runWriter(&wg)

	// Wait for completion or error
	return p.waitForCompletion(&wg)
}

// runReader reads input and creates tasks
func (p *pipeline) runReader(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(p.stream.taskChan)

	if err := p.stream.readTasks(p.input); err != nil {
		p.sendError(fmt.Errorf("reader error: %w", err))
	}
}

// runWorkers processes tasks using the worker pool
func (p *pipeline) runWorkers(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(p.stream.resultChan)

	if err := p.stream.pool.Process(p.stream.ctx, p.stream.taskChan, p.stream.resultChan); err != nil {
		p.sendError(fmt.Errorf("worker error: %w", err))
	}
}

// runWriter writes processed results
func (p *pipeline) runWriter(wg *sync.WaitGroup) {
	defer wg.Done()

	if err := p.stream.writeResults(p.output, p.progressCallback); err != nil {
		p.sendError(fmt.Errorf("writer error: %w", err))
	}
}

// sendError sends an error to the error channel if possible
func (p *pipeline) sendError(err error) {
	select {
	case p.errChan <- err:
	case <-p.stream.ctx.Done():
	}
}

// waitForCompletion waits for pipeline completion or handles errors
func (p *pipeline) waitForCompletion(wg *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-p.errChan:
		p.stream.cancel() // Cancel other goroutines
		wg.Wait()         // Wait for cleanup
		return err
	case <-done:
		return nil
	case <-p.stream.ctx.Done():
		wg.Wait()
		return constants.ErrCanceled
	}
}

// readTasks reads input based on processing type
func (s *StreamProcessor) readTasks(input io.Reader) error {
	switch s.config.Processing {
	case constants.Encryption:
		return s.readForEncryption(input)
	case constants.Decryption:
		return s.readForDecryption(input)
	default:
		return fmt.Errorf("unknown processing type: %d", s.config.Processing)
	}
}

// readForEncryption reads raw data in fixed-size chunks
func (s *StreamProcessor) readForEncryption(reader io.Reader) error {
	buffer := make([]byte, s.config.ChunkSize)
	var index uint64

	for {
		if err := s.checkCancellation(); err != nil {
			return err
		}

		n, err := reader.Read(buffer)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read failed: %w", err)
		}

		task := constants.Task{
			Data:  make([]byte, n),
			Index: index,
		}
		copy(task.Data, buffer[:n])

		if err := s.sendTask(task); err != nil {
			return err
		}
		index++
	}
}

// readForDecryption reads data with length prefixes
func (s *StreamProcessor) readForDecryption(reader io.Reader) error {
	var index uint64

	for {
		if err := s.checkCancellation(); err != nil {
			return err
		}

		chunkLen, err := s.readChunkSize(reader)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if chunkLen == 0 {
			continue // Skip empty chunks
		}

		data, err := s.readChunkData(reader, chunkLen)
		if err != nil {
			return err
		}

		task := constants.Task{
			Data:  data,
			Index: index,
		}

		if err := s.sendTask(task); err != nil {
			return err
		}
		index++
	}
}

// readChunkSize reads the 4-byte chunk size header
func (s *StreamProcessor) readChunkSize(reader io.Reader) (uint32, error) {
	var sizeBuffer [constants.ChunkHeaderSize]byte
	_, err := io.ReadFull(reader, sizeBuffer[:])
	if err != nil {
		if err == io.EOF {
			return 0, err
		}
		return 0, fmt.Errorf("chunk size read failed: %w", err)
	}
	return binary.BigEndian.Uint32(sizeBuffer[:]), nil
}

// readChunkData reads the chunk data of specified length
func (s *StreamProcessor) readChunkData(reader io.Reader, length uint32) ([]byte, error) {
	if length > math.MaxInt32 {
		return nil, constants.ErrChunkTooLarge
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("chunk data read failed: %w", err)
	}
	return data, nil
}

// checkCancellation checks if the context is cancelled
func (s *StreamProcessor) checkCancellation() error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return nil
	}
}

// sendTask sends a task through the task channel
func (s *StreamProcessor) sendTask(task constants.Task) error {
	select {
	case s.taskChan <- task:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// writeResults processes and writes results in order
func (s *StreamProcessor) writeResults(writer io.Writer, progressCallback func(int64)) error {
	buffer := NewBuffer()

	for {
		select {
		case result, ok := <-s.resultChan:
			if !ok {
				return s.flushRemainingResults(writer, buffer, progressCallback)
			}

			if result.Err != nil {
				return fmt.Errorf("processing chunk %d: %w", result.Index, result.Err)
			}

			ready := buffer.Add(result)
			if err := s.writeReadyResults(writer, ready, progressCallback); err != nil {
				return err
			}

		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}
}

// flushRemainingResults writes any remaining buffered results
func (s *StreamProcessor) flushRemainingResults(writer io.Writer, buffer *Buffer, progressCallback func(int64)) error {
	remaining := buffer.Flush()
	return s.writeReadyResults(writer, remaining, progressCallback)
}

// writeReadyResults writes a slice of ready results
func (s *StreamProcessor) writeReadyResults(writer io.Writer, results []constants.TaskResult, progressCallback func(int64)) error {
	for _, result := range results {
		if err := s.writeResult(writer, result, progressCallback); err != nil {
			return err
		}
	}
	return nil
}

// writeResult writes a single result to the output
func (s *StreamProcessor) writeResult(writer io.Writer, result constants.TaskResult, progressCallback func(int64)) error {
	// Write chunk size header for encryption
	if s.config.Processing == constants.Encryption {
		if err := s.writeChunkSize(writer, len(result.Data)); err != nil {
			return fmt.Errorf("writing chunk size: %w", err)
		}
	}

	// Write chunk data
	if _, err := writer.Write(result.Data); err != nil {
		return fmt.Errorf("writing chunk data: %w", err)
	}

	// Update progress
	if progressCallback != nil {
		progressCallback(int64(result.Size))
	}

	return nil
}

// writeChunkSize writes the chunk size as a 4-byte big-endian integer
func (s *StreamProcessor) writeChunkSize(writer io.Writer, size int) error {
	if size < 0 || size > math.MaxUint32 {
		return fmt.Errorf("chunk size out of range: %d", size)
	}

	var buffer [constants.ChunkHeaderSize]byte
	binary.BigEndian.PutUint32(buffer[:], uint32(size))

	if _, err := writer.Write(buffer[:]); err != nil {
		return fmt.Errorf("chunk size write failed: %w", err)
	}
	return nil
}

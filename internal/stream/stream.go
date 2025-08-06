package stream

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"

	"github.com/hambosto/hexwarden/internal/processor"
	"github.com/hambosto/hexwarden/internal/ui"
)

// Constants
const (
	DefaultChunkSize = 1 * 1024 * 1024 // 1MB
	DefaultQueueSize = 100
	ChunkHeaderSize  = 4 // Size of chunk length header in bytes
)

// Processing represents the operation type
type Processing int

const (
	Encryption Processing = iota
	Decryption
)

func (p Processing) String() string {
	switch p {
	case Encryption:
		return "Encrypting..."
	case Decryption:
		return "Decrypting..."
	default:
		return "Processing..."
	}
}

// Errors
var (
	ErrInvalidKey    = errors.New("key must be 32 bytes")
	ErrNilStream     = errors.New("input and output streams must not be nil")
	ErrCanceled      = errors.New("operation was canceled")
	ErrChunkTooLarge = errors.New("chunk size exceeds maximum allowed")
	ErrInvalidChunk  = errors.New("invalid chunk data")
)

// Task represents a processing task
type Task struct {
	Data  []byte
	Index uint64
}

// TaskResult represents the result of a processed task
type TaskResult struct {
	Index uint64
	Data  []byte
	Size  int
	Err   error
}

// Config holds stream configuration
type Config struct {
	Key         []byte
	Processing  Processing
	Concurrency int
	QueueSize   int
	ChunkSize   int
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Key) != 32 {
		return ErrInvalidKey
	}
	return nil
}

// ApplyDefaults applies default values to unset configuration fields
func (c *Config) ApplyDefaults() {
	if c.Concurrency <= 0 {
		c.Concurrency = runtime.NumCPU()
	}
	if c.QueueSize <= 0 {
		c.QueueSize = DefaultQueueSize
	}
	if c.ChunkSize <= 0 {
		c.ChunkSize = DefaultChunkSize
	}
}

// Stream handles concurrent encryption/decryption streaming
type Stream struct {
	processor *processor.Processor
	bar       *ui.ProgressBar
	config    Config
	pool      *Pool

	// Channels for task processing pipeline
	taskChan   chan Task
	resultChan chan TaskResult

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Stream instance
func New(config Config) (*Stream, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	processor, err := processor.New(config.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	config.ApplyDefaults()
	ctx, cancel := context.WithCancel(context.Background())

	s := &Stream{
		processor: processor,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
	}

	s.pool = NewPool(config.Concurrency, s.processTask)
	return s, nil
}

// Cancel cancels the stream processing
func (s *Stream) Cancel() {
	s.cancel()
}

// Process processes data from input to output with progress tracking
func (s *Stream) Process(input io.Reader, output io.Writer, totalSize int64) error {
	if input == nil || output == nil {
		return ErrNilStream
	}

	s.bar = ui.NewProgressBar(totalSize, s.config.Processing.String())
	return s.runPipeline(input, output)
}

// processTask processes a single task based on the operation type
func (s *Stream) processTask(task Task) TaskResult {
	var output []byte
	var err error

	switch s.config.Processing {
	case Encryption:
		output, err = s.processor.Encrypt(task.Data)
	case Decryption:
		output, err = s.processor.Decrypt(task.Data)
	default:
		err = fmt.Errorf("unknown processing type: %d", s.config.Processing)
	}

	// Calculate size for progress tracking
	size := s.calculateProgressSize(task.Data, output)

	return TaskResult{
		Index: task.Index,
		Data:  output,
		Size:  size,
		Err:   err,
	}
}

// calculateProgressSize determines the size to use for progress tracking
func (s *Stream) calculateProgressSize(input, output []byte) int {
	if s.config.Processing == Encryption {
		return len(input) // Track input size for encryption
	}
	return len(output) // Track output size for decryption
}

// runPipeline orchestrates the concurrent processing pipeline
func (s *Stream) runPipeline(input io.Reader, output io.Writer) error {
	// Initialize buffered channels for better throughput
	s.taskChan = make(chan Task, s.config.QueueSize)
	s.resultChan = make(chan TaskResult, s.config.QueueSize)

	pipeline := &pipeline{
		stream:  s,
		input:   input,
		output:  output,
		errChan: make(chan error, 3), // Reader, workers, writer
	}

	return pipeline.run()
}

// pipeline manages the concurrent processing stages
type pipeline struct {
	stream  *Stream
	input   io.Reader
	output  io.Writer
	errChan chan error
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

	if err := p.stream.writeResults(p.output); err != nil {
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
		return ErrCanceled
	}
}

// readTasks reads input based on processing type
func (s *Stream) readTasks(input io.Reader) error {
	switch s.config.Processing {
	case Encryption:
		return s.readForEncryption(input)
	case Decryption:
		return s.readForDecryption(input)
	default:
		return fmt.Errorf("unknown processing type: %d", s.config.Processing)
	}
}

// readForEncryption reads raw data in fixed-size chunks
func (s *Stream) readForEncryption(reader io.Reader) error {
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

		task := Task{
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
func (s *Stream) readForDecryption(reader io.Reader) error {
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

		task := Task{
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
func (s *Stream) readChunkSize(reader io.Reader) (uint32, error) {
	var sizeBuffer [ChunkHeaderSize]byte
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
func (s *Stream) readChunkData(reader io.Reader, length uint32) ([]byte, error) {
	if length > math.MaxInt32 {
		return nil, ErrChunkTooLarge
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, fmt.Errorf("chunk data read failed: %w", err)
	}
	return data, nil
}

// checkCancellation checks if the context is cancelled
func (s *Stream) checkCancellation() error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	default:
		return nil
	}
}

// sendTask sends a task through the task channel
func (s *Stream) sendTask(task Task) error {
	select {
	case s.taskChan <- task:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// writeResults processes and writes results in order
func (s *Stream) writeResults(writer io.Writer) error {
	buffer := NewBuffer()

	for {
		select {
		case result, ok := <-s.resultChan:
			if !ok {
				return s.flushRemainingResults(writer, buffer)
			}

			if result.Err != nil {
				return fmt.Errorf("processing chunk %d: %w", result.Index, result.Err)
			}

			ready := buffer.add(result)
			if err := s.writeReadyResults(writer, ready); err != nil {
				return err
			}

		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}
}

// flushRemainingResults writes any remaining buffered results
func (s *Stream) flushRemainingResults(writer io.Writer, buffer *Buffer) error {
	remaining := buffer.flush()
	return s.writeReadyResults(writer, remaining)
}

// writeReadyResults writes a slice of ready results
func (s *Stream) writeReadyResults(writer io.Writer, results []TaskResult) error {
	for _, result := range results {
		if err := s.writeResult(writer, result); err != nil {
			return err
		}
	}
	return nil
}

// writeResult writes a single result to the output
func (s *Stream) writeResult(writer io.Writer, result TaskResult) error {
	// Write chunk size header for encryption
	if s.config.Processing == Encryption {
		if err := s.writeChunkSize(writer, len(result.Data)); err != nil {
			return fmt.Errorf("writing chunk size: %w", err)
		}
	}

	// Write chunk data
	if _, err := writer.Write(result.Data); err != nil {
		return fmt.Errorf("writing chunk data: %w", err)
	}

	// Update progress bar
	return s.updateProgress(result.Size)
}

// writeChunkSize writes the chunk size as a 4-byte big-endian integer
func (s *Stream) writeChunkSize(writer io.Writer, size int) error {
	if size < 0 || size > math.MaxUint32 {
		return fmt.Errorf("chunk size out of range: %d", size)
	}

	var buffer [ChunkHeaderSize]byte
	binary.BigEndian.PutUint32(buffer[:], uint32(size))

	if _, err := writer.Write(buffer[:]); err != nil {
		return fmt.Errorf("chunk size write failed: %w", err)
	}
	return nil
}

// updateProgress updates the progress bar if available
func (s *Stream) updateProgress(size int) error {
	if s.bar != nil {
		if err := s.bar.Add(int64(size)); err != nil {
			return fmt.Errorf("updating progress: %w", err)
		}
	}
	return nil
}

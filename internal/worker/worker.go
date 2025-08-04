package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/hambosto/hexwarden/internal/processor"
	"github.com/hambosto/hexwarden/internal/ui"
)

const (
	DefaultChunkSize = 1 * 1024 * 1024 // 1 MB
	DefaultQueueSize = 100
)

type Processing int

const (
	Encryption Processing = iota
	Decryption
)

var (
	ErrInvalidKey = errors.New("key must be 32 bytes")
	ErrNilStream  = errors.New("input and output streams must not be nil")
	ErrCanceled   = errors.New("operation was canceled")
)

type Task struct {
	Data  []byte
	Index uint64
}

type TaskResult struct {
	Index uint64
	Data  []byte
	Size  int
	Err   error
}

type Config struct {
	Key         []byte
	Processing  Processing
	Concurrency int
	QueueSize   int
	ChunkSize   int
}

type Worker struct {
	processor *processor.Processor
	bar       *ui.ProgressBar
	config    Config

	// Channels
	taskChan   chan Task
	resultChan chan TaskResult

	// Synchronization
	ctx        context.Context
	cancel     context.CancelFunc
	workerPool *WorkerPool
}

func New(config Config) (*Worker, error) {
	if len(config.Key) != 32 {
		return nil, ErrInvalidKey
	}

	processor, err := processor.New(config.Key)
	if err != nil {
		return nil, fmt.Errorf("creating processor: %w", err)
	}

	// Set defaults
	if config.Concurrency <= 0 {
		config.Concurrency = runtime.NumCPU()
	}
	if config.QueueSize <= 0 {
		config.QueueSize = DefaultQueueSize
	}
	if config.ChunkSize <= 0 {
		config.ChunkSize = DefaultChunkSize
	}

	ctx, cancel := context.WithCancel(context.Background())

	w := &Worker{
		processor: processor,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
	}

	w.workerPool = NewWorkerPool(config.Concurrency, w.processTask)

	return w, nil
}

func (w *Worker) Process(input io.Reader, output io.Writer, totalSize int64) error {
	if input == nil || output == nil {
		return ErrNilStream
	}

	label := "Encrypting..."
	if w.config.Processing != Encryption {
		label = "Decrypting..."
	}

	w.bar = ui.NewProgressBar(totalSize, label)

	return w.runPipeline(input, output)
}

func (w *Worker) processTask(task Task) TaskResult {
	var (
		output []byte
		err    error
	)

	if w.config.Processing == Encryption {
		output, err = w.processor.Encrypt(task.Data)
	} else {
		output, err = w.processor.Decrypt(task.Data)
	}

	size := len(task.Data)
	if w.config.Processing != Encryption {
		size = len(output)
	}

	return TaskResult{
		Index: task.Index,
		Data:  output,
		Size:  size,
		Err:   err,
	}
}

func (w *Worker) Cancel() {
	w.cancel()
}

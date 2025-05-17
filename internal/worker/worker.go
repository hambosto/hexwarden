package worker

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/hambosto/hexwarden/internal/processor"
	"github.com/hambosto/hexwarden/internal/ui"
)

const (
	DefaultChunkSize = 1024 * 1024
)

type Processing int

const (
	Encryption Processing = iota
	Decryption
)

var (
	ErrInvalidKey = errors.New("key must be 64 bytes")
	ErrNilStream  = errors.New("input and output streams must not be nil")
)

type Task struct {
	Data  []byte
	Index uint32
}

type TaskResult struct {
	Index uint32
	Data  []byte
	Size  int
	Err   error
}

type Worker struct {
	processor   *processor.Processor
	bar         *ui.ProgressBar
	concurrency int
	processing  Processing
}

func New(key []byte, processing Processing) (*Worker, error) {
	if len(key) != 64 {
		return nil, ErrInvalidKey
	}

	p, err := processor.NewProcessor(key)
	if err != nil {
		return nil, fmt.Errorf("creating chunk processor: %w", err)
	}

	concurrency := max(runtime.NumCPU(), 1)

	return &Worker{
		processor:   p,
		concurrency: concurrency,
		processing:  processing,
	}, nil
}

func (w *Worker) WithConcurrency(count int) *Worker {
	if count > 0 {
		w.concurrency = count
	}
	return w
}

func (w *Worker) Process(input io.Reader, output io.Writer, totalSize int64) error {
	if input == nil || output == nil {
		return ErrNilStream
	}

	label := "Encrypting..."
	if w.processing != Encryption {
		label = "Decrypting..."
	}

	w.bar = ui.NewProgressBar(totalSize, label)
	return w.runPipeline(input, output)
}

func (w *Worker) GetCipherNonce() []byte {
	return w.processor.Cipher.GetNonce()
}

func (w *Worker) SetCipherNonce(nonce []byte) error {
	return w.processor.Cipher.SetNonce(nonce)
}

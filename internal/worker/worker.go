package worker

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/hambosto/hexwarden/internal/processor"
	"github.com/hambosto/hexwarden/internal/ui"
	"github.com/schollz/progressbar/v3"
)

const (
	DefaultChunkSize = 1024 * 1024
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
	processor   *processor.ChunkProcessor
	progress    *progressbar.ProgressBar
	concurrency int
}

func New(key []byte, processingMode ui.ProcessorMode) (*Worker, error) {
	if len(key) != 64 {
		return nil, ErrInvalidKey
	}

	p, err := processor.NewChunkProcessor(key, processingMode)
	if err != nil {
		return nil, fmt.Errorf("creating chunk processor: %w", err)
	}

	concurrency := max(runtime.NumCPU(), 1)

	return &Worker{
		processor:   p,
		concurrency: concurrency,
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

	w.setProgress(totalSize)
	return w.runPipeline(input, output)
}

func (w *Worker) GetCipherNonce() []byte {
	return w.processor.Cipher.GetNonce()
}

func (w *Worker) SetCipherNonce(nonce []byte) error {
	return w.processor.Cipher.SetNonce(nonce)
}

func (w *Worker) setProgress(size int64) {
	label := "Encrypting..."
	if w.processor.ProcessingMode != ui.ModeEncrypt {
		label = "Decrypting..."
	}

	w.progress = progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(label),
		progressbar.OptionUseANSICodes(false),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.ThemeUnicode),
	)
}

package ui

import (
	"fmt"
	"time"

	"github.com/hambosto/hexwarden/internal/infrastructure/utils"
	"github.com/schollz/progressbar/v3"
)

// ProgressBar provides progress tracking functionality
type ProgressBar struct {
	bar         *progressbar.ProgressBar
	totalSize   int64
	currentSize int64
	startTime   time.Time
	description string
}

// NewProgressBar creates a new progress bar with the given total size and description
func NewProgressBar(totalSize int64, description string) *ProgressBar {
	bar := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(50),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	return &ProgressBar{
		bar:         bar,
		totalSize:   totalSize,
		currentSize: 0,
		startTime:   time.Now(),
		description: description,
	}
}

// Add increments the progress bar by the given amount
func (p *ProgressBar) Add(size int64) error {
	p.currentSize += size
	if p.currentSize > p.totalSize {
		p.currentSize = p.totalSize
	}

	return p.bar.Add64(size)
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() error {
	return p.bar.Finish()
}

// GetElapsedTime returns the elapsed time since the progress bar was created
func (p *ProgressBar) GetElapsedTime() time.Duration {
	return time.Since(p.startTime)
}

// GetProcessingRate returns the current processing rate in bytes per second
func (p *ProgressBar) GetProcessingRate() float64 {
	elapsed := p.GetElapsedTime()
	if elapsed.Seconds() == 0 {
		return 0
	}

	return float64(p.currentSize) / elapsed.Seconds()
}

// CreateCallback creates a progress callback function for use with streaming
func (p *ProgressBar) CreateCallback() func(int64) {
	return func(size int64) {
		p.Add(size)
	}
}

// ShowFinalStats displays final statistics when processing is complete
func (p *ProgressBar) ShowFinalStats() {
	elapsed := p.GetElapsedTime()
	rate := p.GetProcessingRate()

	fmt.Printf("\nProcessing completed!\n")
	fmt.Printf("Total processed: %s\n", utils.FormatBytes(p.currentSize))
	fmt.Printf("Time elapsed: %v\n", elapsed.Round(time.Second))
	fmt.Printf("Average rate: %s/s\n", utils.FormatBytes(int64(rate)))
}

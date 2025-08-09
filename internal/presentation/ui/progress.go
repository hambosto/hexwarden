package ui

import (
	"github.com/schollz/progressbar/v3"
)

// ProgressBar provides progress tracking functionality
type ProgressBar struct {
	bar         *progressbar.ProgressBar
	description string
}

// NewProgressBar creates a new progress bar with the given total size and description
func NewProgressBar(totalSize int64, description string) *ProgressBar {
	bar := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription(description),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionFullWidth(),
		progressbar.OptionShowBytes(true),
	)

	return &ProgressBar{
		bar:         bar,
		description: description,
	}
}

// Add increments the progress bar by the given amount
func (p *ProgressBar) Add(size int64) error {
	return p.bar.Add64(size)
}

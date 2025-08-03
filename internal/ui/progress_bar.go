package ui

import (
	"github.com/schollz/progressbar/v3"
)

// ProgressBar wraps a progressbar.ProgressBar instance.
type ProgressBar struct {
	bar *progressbar.ProgressBar
}

// NewProgressBar creates a new progress bar with the given size and label.
func NewProgressBar(size int64, label string) *ProgressBar {
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(label),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	return &ProgressBar{bar: bar}
}

// Add increments the progress bar by n and returns any error encountered.
func (p *ProgressBar) Add(n int64) error {
	return p.bar.Add64(n)
}

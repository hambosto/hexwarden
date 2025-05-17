package ui

import (
	"github.com/schollz/progressbar/v3"
)

type ProgressBar struct {
	bar *progressbar.ProgressBar
}

func NewProgressBar(size int64, label string) *ProgressBar {
	bar := progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(label),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.ThemeUnicode),
	)

	return &ProgressBar{bar: bar}
}

func (p *ProgressBar) Add(n int64) error {
	return p.bar.Add64(n)
}

package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	ProgressMsg int64
	QuitMsg     struct{}
)

type Spinner struct {
	program    *tea.Program
	progressCh chan ProgressMsg
	quitCh     chan struct{}
	wg         sync.WaitGroup
	config     Config
}

type Config struct {
	TotalSize      int64
	OperationLabel string
	Style          StyleConfig
}

type StyleConfig struct {
	Spinner  lipgloss.Style
	Label    lipgloss.Style
	Stats    lipgloss.Style
	Complete lipgloss.Style
}

func DefaultConfig(totalSize int64, operationLabel string) Config {
	return Config{
		TotalSize:      totalSize,
		OperationLabel: operationLabel,
		Style: StyleConfig{
			Spinner:  lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
			Label:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E22E")),
			Stats:    lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
			Complete: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7EEDC7")),
		},
	}
}

func NewSpinner(cfg Config) *Spinner {
	return &Spinner{
		progressCh: make(chan ProgressMsg, 100),
		quitCh:     make(chan struct{}),
		config:     cfg,
	}
}

func (s *Spinner) Start() error {
	s.program = tea.NewProgram(newModel(s.config, s.progressCh))

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if _, err := s.program.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running spinner: %v\n", err)
		}
	}()

	return nil
}

func (s *Spinner) Update(size int) {
	select {
	case s.progressCh <- ProgressMsg(size):
	case <-s.quitCh:
		return
	}
}

func (s *Spinner) Stop() {
	close(s.quitCh)
	close(s.progressCh)
	s.wg.Wait()
}

type model struct {
	spinner       spinner.Model
	totalSize     int64
	processedSize int64
	startTime     time.Time
	label         string
	progressCh    chan ProgressMsg
	style         StyleConfig
}

func newModel(cfg Config, progressCh chan ProgressMsg) model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(cfg.Style.Spinner),
	)

	return model{
		spinner:    s,
		totalSize:  cfg.TotalSize,
		startTime:  time.Now(),
		label:      cfg.OperationLabel,
		progressCh: progressCh,
		style:      cfg.Style,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.readProgress(),
	)
}

func (m model) readProgress() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.progressCh
		if !ok {
			return QuitMsg{}
		}
		return msg
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case ProgressMsg:
		m.processedSize += int64(msg)
		if m.processedSize >= m.totalSize {
			return m, tea.Quit
		}
		return m, m.readProgress()
	case QuitMsg:
		return m, tea.Quit
	}

	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)

	return m, tea.Batch(spinnerCmd, m.readProgress())
}

func (m model) View() string {
	var output string

	if m.processedSize < m.totalSize {
		output += m.renderSpinner() + "\n"
	}

	output += m.renderStats()

	return output
}

func (m model) renderSpinner() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.spinner.View(),
		" ",
		m.style.Label.Render(m.label),
	)
}

func (m model) renderStats() string {
	percentComplete := float64(m.processedSize) / float64(m.totalSize) * 100
	bytesInfo := fmt.Sprintf(
		"%s / %s (%.1f%%)",
		formatBytes(m.processedSize),
		formatBytes(m.totalSize),
		percentComplete,
	)

	elapsed := time.Since(m.startTime)
	timeInfo := fmt.Sprintf("Time: %s • ETA: %s",
		formatDuration(elapsed),
		formatDuration(calculateETA(elapsed, m.processedSize, m.totalSize)),
	)

	return m.style.Stats.Render(fmt.Sprintf("%s   %s", bytesInfo, timeInfo))
}

func calculateETA(elapsed time.Duration, processed, total int64) time.Duration {
	if processed <= 0 {
		return 0
	}
	return time.Duration(float64(elapsed) * (float64(total)/float64(processed) - 1))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm %02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

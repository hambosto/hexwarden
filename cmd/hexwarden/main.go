package main

import (
	"fmt"
	"os"

	"github.com/hambosto/hexwarden/internal/app"
	"github.com/hambosto/hexwarden/internal/files"
	"github.com/hambosto/hexwarden/internal/ui"
)

// Config holds application configuration
type Config struct {
	ExcludedDirs    []string
	ExcludedExts    []string
	OverwritePasses int
}

// DefaultConfig returns the default application configuration
func DefaultConfig() *Config {
	return &Config{
		ExcludedDirs:    []string{"vendor/", "node_modules/", ".git", ".github"},
		ExcludedExts:    []string{".go", "go.mod", "go.sum", ".nix", ".gitignore"},
		OverwritePasses: 3,
	}
}

// Dependencies holds all application dependencies
type Dependencies struct {
	Terminal    *ui.Terminal
	Prompt      *ui.Prompt
	FileManager *files.FileManager
	FileFinder  *files.FileFinder
	App         *app.App
}

// NewDependencies creates and initializes all application dependencies
func NewDependencies(config *Config) *Dependencies {
	terminal := ui.NewTerminal()
	prompt := ui.NewPrompt()
	fileManager := files.NewFileManager(config.OverwritePasses)
	fileFinder := files.NewFileFinder(config.ExcludedDirs, config.ExcludedExts)
	app := app.NewApp(fileManager, prompt)

	return &Dependencies{
		Terminal:    terminal,
		Prompt:      prompt,
		FileManager: fileManager,
		FileFinder:  fileFinder,
		App:         app,
	}
}

// Application encapsulates the main application logic
type Application struct {
	deps   *Dependencies
	config *Config
}

// NewApplication creates a new application instance
func NewApplication(config *Config) *Application {
	return &Application{
		deps:   NewDependencies(config),
		config: config,
	}
}

// initializeTerminal sets up the terminal for the application
func (a *Application) initializeTerminal() {
	a.deps.Terminal.Clear()
	a.deps.Terminal.MoveTopLeft()
}

// getEligibleFiles retrieves files that can be processed based on the operation mode
func (a *Application) getEligibleFiles(operation files.ProcessorMode) ([]string, error) {
	eligibleFiles, err := a.deps.FileFinder.FindEligibleFiles(operation)
	if err != nil {
		return nil, fmt.Errorf("failed to find eligible files: %w", err)
	}

	if len(eligibleFiles) == 0 {
		return nil, fmt.Errorf("no eligible files found")
	}

	return eligibleFiles, nil
}

// Run executes the main application workflow
func (a *Application) Run() error {
	// Initialize terminal
	a.initializeTerminal()

	// Get processing mode from user
	operation, err := a.deps.Prompt.GetProcessingMode()
	if err != nil {
		return fmt.Errorf("failed to get processing mode: %w", err)
	}

	// Find eligible files
	eligibleFiles, err := a.getEligibleFiles(files.ProcessorMode(operation))
	if err != nil {
		return err
	}

	// Let user choose a file
	selectedFile, err := a.deps.Prompt.ChooseFile(eligibleFiles)
	if err != nil {
		return fmt.Errorf("failed to select file: %w", err)
	}

	// Process the selected file
	if err := a.deps.App.ProcessFile(selectedFile, operation); err != nil {
		return fmt.Errorf("failed to process file '%s': %w", selectedFile, err)
	}

	return nil
}

func main() {
	config := DefaultConfig()
	app := NewApplication(config)

	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v", err)
		os.Exit(1)
	}
}

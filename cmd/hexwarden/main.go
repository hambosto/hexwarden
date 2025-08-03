package main

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/app"
	"github.com/hambosto/hexwarden/internal/files"
	"github.com/hambosto/hexwarden/internal/ui"
)

func main() {
	terminal := ui.NewTerminal()
	terminal.Clear()
	terminal.MoveTopLeft()

	excludedDirs := []string{"vendor/", "node_modules/", ".git", ".github"}
	excludedExts := []string{".go", "go.mod", "go.sum", ".nix", ".gitignore"}

	prompt := ui.NewPrompt()
	fileManager := files.NewFileManager(3)
	fileFinder := files.NewFileFinder(excludedDirs, excludedExts)
	app := app.NewApp(fileManager, prompt)

	operation, err := prompt.GetProcessingMode()
	if err != nil {
		fmt.Printf("Error: failed to get operation: %v", err)
	}

	files, err := fileFinder.FindEligibleFiles(files.ProcessorMode(operation))
	if err != nil {
		fmt.Printf("Error: failed to list files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No eligible files found.")
	}

	selectedFile, err := prompt.ChooseFile(files)
	if err != nil {
		fmt.Printf("Error: failed to select file: %v", err)
	}

	if err := app.ProcessFile(selectedFile, operation); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

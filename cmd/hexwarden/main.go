package main

import (
	"fmt"
	"os"

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
		fmt.Printf("Error: failed to get operation: %v\n", err)
		os.Exit(1)
	}

	files, err := fileFinder.FindEligibleFiles(files.ProcessorMode(operation))
	if err != nil {
		fmt.Printf("Error: failed to list files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No eligible files found.")
		os.Exit(1)
	}

	selectedFile, err := prompt.ChooseFile(files)
	if err != nil {
		fmt.Printf("Error: failed to select file: %v\n", err)
		os.Exit(1)
	}

	if err := app.ProcessFile(selectedFile, operation); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

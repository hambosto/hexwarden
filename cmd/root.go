package cmd

import (
	"fmt"
	"os"

	"github.com/hambosto/hexwarden/internal/core"
	"github.com/hambosto/hexwarden/internal/ui"
)

func Execute() {
	terminal := ui.NewTerminal()
	terminal.Clear()
	terminal.MoveTopLeft()

	prompt := ui.NewPrompt()
	fileManager := core.NewFileManager(3)
	processor := core.NewProcessor(fileManager, prompt)

	operation, err := prompt.GetProcessingMode()
	if err != nil {
		fmt.Printf("Error: failed to get operation: %v\n", err)
		os.Exit(1)
	}

	fileFinder := ui.NewFileFinder()
	files, err := fileFinder.FindEligibleFiles(operation)
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

	if err := processor.ProcessFile(selectedFile, operation); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

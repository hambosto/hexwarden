package main

import (
	"os"

	"github.com/hambosto/hexwarden/cmd/cli"
	"github.com/hambosto/hexwarden/cmd/interactive"
)

func main() {
	// If command-line arguments are provided, use CLI mode
	if len(os.Args) > 1 {
		// Initialize and execute CLI commands
		cliApp := cli.NewCLI()
		cliApp.Execute()
	} else {
		// No arguments provided, default to interactive mode
		interactiveApp := interactive.NewInteractiveApp()
		interactiveApp.Run()
	}
}

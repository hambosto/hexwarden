package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/presentation/interactive"
)

// CLI represents the command-line interface
type CLI struct {
	rootCmd *cobra.Command
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	cli := &CLI{}
	cli.setupCommands()
	return cli
}

// Execute runs the CLI
func (c *CLI) Execute() error {
	return c.rootCmd.Execute()
}

// setupCommands initializes all CLI commands
func (c *CLI) setupCommands() {
	c.rootCmd = &cobra.Command{
		Use:   "hexwarden",
		Short: "Secure file encryption tool with Reed-Solomon error correction",
		Long: `HexWarden is a secure file encryption tool that uses AES-256-GCM encryption
with Argon2id key derivation and Reed-Solomon error correction codes.

It supports both interactive mode (default) and command-line mode for automation.`,
		Version: constants.AppVersion,
		Run: func(cmd *cobra.Command, args []string) {
			// Default behavior: run interactive mode
			interactiveApp := interactive.NewInteractiveApp()
			interactiveApp.Run()
		},
	}

	// Add subcommands
	c.rootCmd.AddCommand(c.createEncryptCommand())
	c.rootCmd.AddCommand(c.createDecryptCommand())
	c.rootCmd.AddCommand(c.createInteractiveCommand())
}

// createEncryptCommand creates the encrypt subcommand
func (c *CLI) createEncryptCommand() *cobra.Command {
	var (
		inputFile    string
		outputFile   string
		password     string
		deleteSource bool
		secureDelete bool
	)

	cmd := &cobra.Command{
		Use:   "encrypt [flags]",
		Short: "Encrypt a file",
		Long:  "Encrypt a file using AES-256-GCM with Reed-Solomon error correction",
		Example: `  hexwarden encrypt -i document.txt -o document.txt.hex
  hexwarden encrypt -i document.txt -p mypassword --delete-source
  hexwarden encrypt -i document.txt --secure-delete`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runEncrypt(inputFile, outputFile, password, deleteSource, secureDelete)
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file to encrypt (required)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output encrypted file (default: input + .hex)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Encryption password (will prompt if not provided)")
	cmd.Flags().BoolVar(&deleteSource, "delete-source", false, "Delete source file after encryption")
	cmd.Flags().BoolVar(&secureDelete, "secure-delete", false, "Use secure deletion (slower but unrecoverable)")

	if err := cmd.MarkFlagRequired("input"); err != nil {
		// This should not happen in normal circumstances
		panic(fmt.Sprintf("failed to mark input flag as required: %v", err))
	}

	return cmd
}

// createDecryptCommand creates the decrypt subcommand
func (c *CLI) createDecryptCommand() *cobra.Command {
	var (
		inputFile    string
		outputFile   string
		password     string
		deleteSource bool
		secureDelete bool
	)

	cmd := &cobra.Command{
		Use:   "decrypt [flags]",
		Short: "Decrypt a file",
		Long:  "Decrypt a file encrypted with HexWarden",
		Example: `  hexwarden decrypt -i document.txt.hex -o document.txt
  hexwarden decrypt -i document.txt.hex -p mypassword
  hexwarden decrypt -i document.txt.hex --delete-source`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runDecrypt(inputFile, outputFile, password, deleteSource, secureDelete)
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file to decrypt (required)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output decrypted file (default: remove .hex extension)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Decryption password (will prompt if not provided)")
	cmd.Flags().BoolVar(&deleteSource, "delete-source", false, "Delete source file after decryption")
	cmd.Flags().BoolVar(&secureDelete, "secure-delete", false, "Use secure deletion (slower but unrecoverable)")

	if err := cmd.MarkFlagRequired("input"); err != nil {
		// This should not happen in normal circumstances
		panic(fmt.Sprintf("failed to mark input flag as required: %v", err))
	}

	return cmd
}

// createInteractiveCommand creates the interactive subcommand
func (c *CLI) createInteractiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "interactive",
		Short: "Run in interactive mode",
		Long:  "Run HexWarden in interactive mode with guided prompts",
		Run: func(cmd *cobra.Command, args []string) {
			interactiveApp := interactive.NewInteractiveApp()
			interactiveApp.Run()
		},
	}
}

// runEncrypt handles the encrypt command
func (c *CLI) runEncrypt(inputFile, outputFile, password string, deleteSource, secureDelete bool) error {
	// Validate input file
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Set default output file if not provided
	if outputFile == "" {
		outputFile = inputFile + constants.FileExtension
	}

	// Check if output file already exists
	if _, err := os.Stat(outputFile); err == nil {
		return fmt.Errorf("output file already exists: %s", outputFile)
	}

	// Create CLI processor
	processor := NewCLIProcessor()

	// Run encryption
	return processor.Encrypt(inputFile, outputFile, password, deleteSource, secureDelete)
}

// runDecrypt handles the decrypt command
func (c *CLI) runDecrypt(inputFile, outputFile, password string, deleteSource, secureDelete bool) error {
	// Validate input file
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Set default output file if not provided
	if outputFile == "" {
		if len(inputFile) > len(constants.FileExtension) &&
			inputFile[len(inputFile)-len(constants.FileExtension):] == constants.FileExtension {
			outputFile = inputFile[:len(inputFile)-len(constants.FileExtension)]
		} else {
			return fmt.Errorf("cannot determine output filename, please specify with -o flag")
		}
	}

	// Check if output file already exists
	if _, err := os.Stat(outputFile); err == nil {
		return fmt.Errorf("output file already exists: %s", outputFile)
	}

	// Create CLI processor
	processor := NewCLIProcessor()

	// Run decryption
	return processor.Decrypt(inputFile, outputFile, password, deleteSource, secureDelete)
}

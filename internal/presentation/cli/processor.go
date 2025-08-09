package cli

import (
	"fmt"
	"syscall"

	"golang.org/x/term"

	"github.com/hambosto/hexwarden/internal/business/operations"
	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/data/files"
	"github.com/hambosto/hexwarden/internal/presentation/ui"
)

// CLIProcessor handles CLI-based encryption and decryption operations
type CLIProcessor struct {
	encryptor   *operations.Encryptor
	decryptor   *operations.Decryptor
	fileManager *files.Manager
}

// NewCLIProcessor creates a new CLI processor instance
func NewCLIProcessor() *CLIProcessor {
	return &CLIProcessor{
		encryptor:   operations.NewEncryptor(),
		decryptor:   operations.NewDecryptor(),
		fileManager: files.NewManager(),
	}
}

// Encrypt encrypts a file using CLI parameters
func (p *CLIProcessor) Encrypt(inputFile, outputFile, password string, deleteSource, secureDelete bool) error {
	// Get password if not provided
	if password == "" {
		var err error
		password, err = p.promptPassword("Enter encryption password: ")
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}

		confirmPassword, err := p.promptPassword("Confirm password: ")
		if err != nil {
			return fmt.Errorf("failed to confirm password: %w", err)
		}

		if password != confirmPassword {
			return constants.ErrPasswordMismatch
		}
	}

	// Get file info for progress tracking
	fileInfo, err := p.fileManager.GetFileInfo(inputFile)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create progress bar
	progressBar := ui.NewProgressBar(fileInfo.Size(), "Encrypting")
	defer func() {
		if err := progressBar.Finish(); err != nil {
			// Log error but don't override the main error
			_ = err // Explicitly ignore the close error to avoid overriding the main error
		}
	}()

	fmt.Printf("Encrypting: %s -> %s\n", inputFile, outputFile)

	// Perform encryption
	err = p.encryptor.EncryptFile(inputFile, outputFile, password, progressBar.CreateCallback())
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Show final stats
	progressBar.ShowFinalStats()

	// Handle source file deletion if requested
	if deleteSource {
		deleteOption := constants.DeleteStandard
		if secureDelete {
			deleteOption = constants.DeleteSecure
		}

		fmt.Printf("Deleting source file: %s\n", inputFile)
		if err := p.fileManager.Remove(inputFile, deleteOption); err != nil {
			fmt.Printf("Warning: Failed to delete source file: %v\n", err)
		} else {
			fmt.Printf("Source file deleted successfully\n")
		}
	}

	fmt.Printf("✓ File encrypted successfully: %s\n", outputFile)
	return nil
}

// Decrypt decrypts a file using CLI parameters
func (p *CLIProcessor) Decrypt(inputFile, outputFile, password string, deleteSource, secureDelete bool) error {
	// Get password if not provided
	if password == "" {
		var err error
		password, err = p.promptPassword("Enter decryption password: ")
		if err != nil {
			return fmt.Errorf("failed to get password: %w", err)
		}
	}

	// Get file info for progress tracking
	fileInfo, err := p.fileManager.GetFileInfo(inputFile)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create progress bar
	progressBar := ui.NewProgressBar(fileInfo.Size(), "Decrypting")
	defer func() {
		if err := progressBar.Finish(); err != nil {
			// Log error but don't override the main error
			_ = err // Explicitly ignore the close error to avoid overriding the main error
		}
	}()

	fmt.Printf("Decrypting: %s -> %s\n", inputFile, outputFile)

	// Perform decryption
	err = p.decryptor.DecryptFile(inputFile, outputFile, password, progressBar.CreateCallback())
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Show final stats
	progressBar.ShowFinalStats()

	// Handle source file deletion if requested
	if deleteSource {
		deleteOption := constants.DeleteStandard
		if secureDelete {
			deleteOption = constants.DeleteSecure
		}

		fmt.Printf("Deleting source file: %s\n", inputFile)
		if err := p.fileManager.Remove(inputFile, deleteOption); err != nil {
			fmt.Printf("Warning: Failed to delete source file: %v\n", err)
		} else {
			fmt.Printf("Source file deleted successfully\n")
		}
	}

	fmt.Printf("✓ File decrypted successfully: %s\n", outputFile)
	return nil
}

// promptPassword prompts for a password without echoing to terminal
func (p *CLIProcessor) promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // Add newline after password input
	return string(bytePassword), nil
}

package interactive

import (
	"fmt"
	"os"

	"github.com/hambosto/hexwarden/internal/business/operations"
	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/data/files"
	"github.com/hambosto/hexwarden/internal/presentation/ui"
)

// InteractiveApp encapsulates the main interactive application
type InteractiveApp struct {
	terminal    *ui.Terminal
	prompt      *ui.Prompt
	fileManager *files.Manager
	fileFinder  *files.Finder
	encryptor   *operations.Encryptor
	decryptor   *operations.Decryptor
}

// NewInteractiveApp creates a new interactive application instance
func NewInteractiveApp() *InteractiveApp {
	return &InteractiveApp{
		terminal:    ui.NewTerminal(),
		prompt:      ui.NewPrompt(),
		fileManager: files.NewManager(),
		fileFinder:  files.NewFinder(),
		encryptor:   operations.NewEncryptor(),
		decryptor:   operations.NewDecryptor(),
	}
}

// Run executes the main interactive application workflow
func (a *InteractiveApp) Run() {
	// Set up terminal
	a.terminal.SetTitle(fmt.Sprintf("%s v%s", constants.AppName, constants.AppVersion))
	a.initializeTerminal()

	// Show banner
	a.terminal.PrintBanner()
	a.terminal.PrintSeparator()

	// Main application loop
	if err := a.runInteractiveLoop(); err != nil {
		a.handleError(err)
		os.Exit(1)
	}

	// Cleanup
	a.terminal.Cleanup()
}

// initializeTerminal sets up the terminal for the application
func (a *InteractiveApp) initializeTerminal() {
	a.terminal.Clear()
	a.terminal.MoveTopLeft()
}

// runInteractiveLoop executes the main interactive workflow
func (a *InteractiveApp) runInteractiveLoop() error {
	// Get processing mode from user
	operation, err := a.prompt.GetProcessingMode()
	if err != nil {
		return fmt.Errorf("failed to get processing mode: %w", err)
	}

	// Find eligible files
	eligibleFiles, err := a.getEligibleFiles(operation)
	if err != nil {
		return err
	}

	// Show file information
	fileInfos, err := a.fileFinder.GetFileInfo(eligibleFiles)
	if err != nil {
		return fmt.Errorf("failed to get file information: %w", err)
	}
	a.prompt.ShowFileInfo(fileInfos)

	// Let user choose a file
	selectedFile, err := a.prompt.ChooseFile(eligibleFiles)
	if err != nil {
		return fmt.Errorf("failed to select file: %w", err)
	}

	// Show processing info
	a.prompt.ShowProcessingInfo(operation, selectedFile)

	// Process the selected file
	if err := a.processFile(selectedFile, operation); err != nil {
		return fmt.Errorf("failed to process file '%s': %w", selectedFile, err)
	}

	return nil
}

// getEligibleFiles retrieves files that can be processed based on the operation mode
func (a *InteractiveApp) getEligibleFiles(operation constants.ProcessorMode) ([]string, error) {
	eligibleFiles, err := a.fileFinder.FindEligibleFiles(operation)
	if err != nil {
		return nil, fmt.Errorf("failed to find eligible files: %w", err)
	}

	if len(eligibleFiles) == 0 {
		return nil, fmt.Errorf("no eligible files found for %s operation", operation)
	}

	return eligibleFiles, nil
}

// processFile handles the file processing workflow
func (a *InteractiveApp) processFile(inputPath string, mode constants.ProcessorMode) error {
	outputPath := a.fileFinder.GetOutputPath(inputPath, mode)

	// Validate paths
	if err := a.fileManager.ValidatePath(inputPath, true); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	if err := a.fileManager.ValidatePath(outputPath, false); err != nil {
		if confirm, confirmErr := a.prompt.ConfirmFileOverwrite(outputPath); confirmErr != nil || !confirm {
			return fmt.Errorf("operation cancelled")
		}
	}

	// Process based on mode
	var err error
	switch mode {
	case constants.ModeEncrypt:
		err = a.encryptFile(inputPath, outputPath)
	case constants.ModeDecrypt:
		err = a.decryptFile(inputPath, outputPath)
	default:
		return fmt.Errorf("unknown processing mode: %v", mode)
	}

	if err != nil {
		return err
	}

	// Ask to delete source file
	fileType := "original"
	if mode == constants.ModeDecrypt {
		fileType = "encrypted"
	}

	if shouldDelete, deleteType, err := a.prompt.ConfirmFileRemoval(inputPath, fmt.Sprintf("Delete %s file", fileType)); err == nil && shouldDelete {
		if err := a.fileManager.Remove(inputPath, deleteType); err != nil {
			a.prompt.ShowWarning(fmt.Sprintf("Failed to delete source file: %v", err))
		} else {
			a.prompt.ShowSuccess(fmt.Sprintf("Source file deleted: %s", inputPath))
		}
	}

	a.prompt.ShowSuccess(fmt.Sprintf("File processed successfully: %s", outputPath))
	return nil
}

// encryptFile handles file encryption
func (a *InteractiveApp) encryptFile(srcPath, destPath string) error {
	// Get password
	password, err := a.prompt.GetEncryptionPassword()
	if err != nil {
		return fmt.Errorf("password prompt failed: %w", err)
	}

	// Get file info for progress tracking
	fileInfo, err := a.fileManager.GetFileInfo(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create progress bar
	progressBar := ui.NewProgressBar(fileInfo.Size(), "Encrypting")
	defer func() {
		if err := progressBar.Finish(); err != nil {
			// Log error but don't override the main error
		}
	}()

	// Perform encryption
	err = a.encryptor.EncryptFile(srcPath, destPath, password, progressBar.CreateCallback())
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Show final stats
	progressBar.ShowFinalStats()
	return nil
}

// decryptFile handles file decryption
func (a *InteractiveApp) decryptFile(srcPath, destPath string) error {
	// Get password
	password, err := a.prompt.GetDecryptionPassword()
	if err != nil {
		return fmt.Errorf("password prompt failed: %w", err)
	}

	// Get file info for progress tracking
	fileInfo, err := a.fileManager.GetFileInfo(srcPath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create progress bar
	progressBar := ui.NewProgressBar(fileInfo.Size(), "Decrypting")
	defer func() {
		if err := progressBar.Finish(); err != nil {
			// Log error but don't override the main error
		}
	}()

	// Perform decryption
	err = a.decryptor.DecryptFile(srcPath, destPath, password, progressBar.CreateCallback())
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Show final stats
	progressBar.ShowFinalStats()
	return nil
}

// handleError handles application errors
func (a *InteractiveApp) handleError(err error) {
	a.terminal.PrintError(fmt.Sprintf("Application error: %v", err))

	// Show additional help for common errors
	switch err {
	case constants.ErrPasswordMismatch:
		a.prompt.ShowInfo("Passwords must match exactly. Please try again.")
	case constants.ErrNoFilesAvailable:
		a.prompt.ShowInfo("No files found for the selected operation. Make sure you're in the right directory.")
	case constants.ErrUserCanceled:
		a.prompt.ShowInfo("Operation cancelled by user.")
	}
}

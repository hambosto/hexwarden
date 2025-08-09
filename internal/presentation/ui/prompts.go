package ui

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/utils"
)

// Prompt provides methods for interactive command-line prompts
type Prompt struct{}

// NewPrompt creates a new Prompt instance
func NewPrompt() *Prompt {
	return &Prompt{}
}

// ConfirmFileOverwrite prompts the user to confirm overwriting an existing file
func (p *Prompt) ConfirmFileOverwrite(path string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Output file %s already exists. Overwrite?", path),
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return false, fmt.Errorf("%w: %v", constants.ErrPromptFailed, err)
	}

	return result, nil
}

// GetEncryptionPassword prompts for and confirms a password for encryption
func (p *Prompt) GetEncryptionPassword() (string, error) {
	password, err := p.getPassword("Enter password:")
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	confirm, err := p.getPassword("Confirm password:")
	if err != nil {
		return "", fmt.Errorf("failed to confirm password: %w", err)
	}

	if password != confirm {
		return "", constants.ErrPasswordMismatch
	}

	return password, nil
}

// GetDecryptionPassword prompts for a password for decryption
func (p *Prompt) GetDecryptionPassword() (string, error) {
	return p.getPassword("Enter password:")
}

// getPassword is a helper method to prompt for a password
func (p *Prompt) getPassword(message string) (string, error) {
	var password string
	prompt := &survey.Password{
		Message: message,
	}

	if err := survey.AskOne(prompt, &password); err != nil {
		return "", fmt.Errorf("%w: %v", constants.ErrPromptFailed, err)
	}

	return password, nil
}

// ConfirmFileRemoval prompts for confirmation and deletion method for file removal
func (p *Prompt) ConfirmFileRemoval(path, message string) (bool, constants.DeleteOption, error) {
	confirmed, err := p.confirmAction(fmt.Sprintf("%s %s", message, path))
	if err != nil {
		return false, "", fmt.Errorf("failed to confirm file removal: %w", err)
	}

	if !confirmed {
		return false, "", nil
	}

	deleteType, err := p.selectDeleteOption()
	if err != nil {
		return false, "", fmt.Errorf("failed to select delete option: %w", err)
	}

	return true, deleteType, nil
}

// confirmAction is a helper method for yes/no confirmation prompts
func (p *Prompt) confirmAction(message string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return false, fmt.Errorf("%w: %v", constants.ErrPromptFailed, err)
	}

	return result, nil
}

// selectDeleteOption prompts the user to select a deletion method
func (p *Prompt) selectDeleteOption() (constants.DeleteOption, error) {
	options := []string{
		string(constants.DeleteStandard),
		string(constants.DeleteSecure),
	}

	selected, err := p.selectFromOptions("Select delete type:", options)
	if err != nil {
		return "", err
	}

	return constants.DeleteOption(selected), nil
}

// GetProcessingMode prompts the user to select a processing operation
func (p *Prompt) GetProcessingMode() (constants.ProcessorMode, error) {
	options := []string{
		string(constants.ModeEncrypt),
		string(constants.ModeDecrypt),
	}

	selected, err := p.selectFromOptions("Select Operation:", options)
	if err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}

	return constants.ProcessorMode(selected), nil
}

// ChooseFile prompts the user to select a file from the provided list
func (p *Prompt) ChooseFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", constants.ErrNoFilesAvailable
	}

	selected, err := p.selectFromOptions("Select file:", files)
	if err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return selected, nil
}

// selectFromOptions is a helper method for selection prompts
func (p *Prompt) selectFromOptions(message string, options []string) (string, error) {
	var selected string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", fmt.Errorf("%w: %v", constants.ErrPromptFailed, err)
	}

	return selected, nil
}

// ShowFileInfo displays information about files to be processed
func (p *Prompt) ShowFileInfo(files []constants.FileInfo) {
	if len(files) == 0 {
		fmt.Println("No files found.")
		return
	}

	fmt.Printf("\nFound %d file(s):\n", len(files))
	for i, file := range files {
		status := "unencrypted"
		if file.IsEncrypted {
			status = "encrypted"
		}

		fmt.Printf("%d. %s (%s, %s)\n",
			i+1,
			file.Path,
			utils.FormatBytes(file.Size),
			status,
		)
	}
	fmt.Println()
}

// ShowProcessingInfo displays information about the processing pipeline
func (p *Prompt) ShowProcessingInfo(mode constants.ProcessorMode, file string) {
	operation := "Encrypting"
	if mode == constants.ModeDecrypt {
		operation = "Decrypting"
	}

	fmt.Printf("\n%s file: %s\n", operation, file)
}

// ShowSuccess displays a success message to the user
func (p *Prompt) ShowSuccess(message string) {
	fmt.Printf("✓ %s\n", message)
}

// ShowWarning displays a warning message to the user
func (p *Prompt) ShowWarning(message string) {
	fmt.Printf("⚠ %s\n", message)
}

// ShowInfo displays an informational message to the user
func (p *Prompt) ShowInfo(message string) {
	fmt.Printf("ℹ %s\n", message)
}

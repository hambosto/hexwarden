// Package ui provides interactive command-line prompts for file operations.
package ui

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

// DeleteOption represents different file deletion methods.
type DeleteOption string

// ProcessorMode represents file processing operations.
type ProcessorMode string

// Available deletion options.
const (
	DeleteStandard DeleteOption = "Normal Delete (faster, but recoverable)"
	DeleteSecure   DeleteOption = "Secure Delete (slower, but unrecoverable)"
)

// Available processing modes.
const (
	ModeEncrypt ProcessorMode = "Encrypt"
	ModeDecrypt ProcessorMode = "Decrypt"
)

// Common errors.
var (
	ErrPasswordMismatch = errors.New("passwords do not match")
	ErrNoFilesAvailable = errors.New("no files available for selection")
)

// Prompt provides methods for interactive command-line prompts.
type Prompt struct{}

// NewPrompt creates a new Prompt instance.
func NewPrompt() *Prompt {
	return &Prompt{}
}

// ConfirmFileOverwrite prompts the user to confirm overwriting an existing file.
func (p *Prompt) ConfirmFileOverwrite(path string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Output file %s already exists. Overwrite?", path),
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return false, fmt.Errorf("failed to confirm file overwrite: %w", err)
	}

	return result, nil
}

// GetEncryptionPassword prompts for and confirms a password for encryption.
// Returns an error if the passwords don't match or if input fails.
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
		return "", ErrPasswordMismatch
	}

	return password, nil
}

// getPassword is a helper method to prompt for a password.
func (p *Prompt) getPassword(message string) (string, error) {
	var password string
	prompt := &survey.Password{
		Message: message,
	}

	return password, survey.AskOne(prompt, &password)
}

// ConfirmFileRemoval prompts for confirmation and deletion method for file removal.
// Returns whether to proceed, the deletion method, and any error.
func (p *Prompt) ConfirmFileRemoval(path, message string) (bool, DeleteOption, error) {
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

// confirmAction is a helper method for yes/no confirmation prompts.
func (p *Prompt) confirmAction(message string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
	}

	return result, survey.AskOne(prompt, &result)
}

// selectDeleteOption prompts the user to select a deletion method.
func (p *Prompt) selectDeleteOption() (DeleteOption, error) {
	options := []string{
		string(DeleteStandard),
		string(DeleteSecure),
	}

	selected, err := p.selectFromOptions("Select delete type:", options)
	if err != nil {
		return "", err
	}

	return DeleteOption(selected), nil
}

// GetProcessingMode prompts the user to select a processing operation.
func (p *Prompt) GetProcessingMode() (ProcessorMode, error) {
	options := []string{
		string(ModeEncrypt),
		string(ModeDecrypt),
	}

	selected, err := p.selectFromOptions("Select Operation:", options)
	if err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}

	return ProcessorMode(selected), nil
}

// ChooseFile prompts the user to select a file from the provided list.
func (p *Prompt) ChooseFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", ErrNoFilesAvailable
	}

	selected, err := p.selectFromOptions("Select file:", files)
	if err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return selected, nil
}

// selectFromOptions is a helper method for selection prompts.
func (p *Prompt) selectFromOptions(message string, options []string) (string, error) {
	var selected string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}

	return selected, survey.AskOne(prompt, &selected)
}

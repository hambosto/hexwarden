package ui

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

type (
	DeleteOption  string
	ProcessorMode string
)

const (
	DeleteStandard DeleteOption  = "Normal Delete (faster, but recoverable)"
	DeleteSecure   DeleteOption  = "Secure Delete (slower, but unrecoverable)"
	ModeEncrypt    ProcessorMode = "Encrypt"
	ModeDecrypt    ProcessorMode = "Decrypt"
)

type Prompt struct{}

func NewPrompt() *Prompt {
	return &Prompt{}
}

func (p *Prompt) ConfirmFileOverwrite(path string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Output file %s already exists. Overwrite?", path),
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (p *Prompt) GetEncryptionPassword() (string, error) {
	var password, confirm string

	passwordPrompt := &survey.Password{
		Message: "Enter password:",
	}
	if err := survey.AskOne(passwordPrompt, &password); err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	confirmPrompt := &survey.Password{
		Message: "Confirm password:",
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return "", fmt.Errorf("failed to confirm password: %w", err)
	}

	if password != confirm {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func (p *Prompt) ConfirmFileRemoval(path, message string) (bool, DeleteOption, error) {
	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("%s %s", message, path),
	}
	if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
		return false, "", err
	}
	if !confirm {
		return false, "", nil
	}

	var deleteType string
	deleteOptions := []string{
		string(DeleteStandard),
		string(DeleteSecure),
	}
	deletePrompt := &survey.Select{
		Message: "Select delete type:",
		Options: deleteOptions,
	}
	if err := survey.AskOne(deletePrompt, &deleteType); err != nil {
		return false, "", err
	}

	return true, DeleteOption(deleteType), nil
}

func (p *Prompt) GetProcessingMode() (ProcessorMode, error) {
	var operationType string
	operationOptions := []string{
		string(ModeEncrypt),
		string(ModeDecrypt),
	}

	prompt := &survey.Select{
		Message: "Select Operation:",
		Options: operationOptions,
	}
	if err := survey.AskOne(prompt, &operationType); err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}

	return ProcessorMode(operationType), nil
}

func (p *Prompt) ChooseFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files available for selection")
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select file:",
		Options: files,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return selected, nil
}

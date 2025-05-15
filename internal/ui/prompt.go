package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
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

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Output file %s already exists. Overwrite?", path)).
				Value(&result),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return false, err
	}

	return result, nil
}

func (p *Prompt) GetEncryptionPassword() (string, error) {
	var (
		password string
		confirm  string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter password:").
				EchoMode(huh.EchoModePassword).
				Value(&password),
			huh.NewInput().
				Title("Confirm password:").
				EchoMode(huh.EchoModePassword).
				Value(&confirm),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	if password != confirm {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func (p *Prompt) ConfirmFileRemoval(path, message string) (bool, DeleteOption, error) {
	var (
		result     bool
		deleteType string
	)

	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("%s %s", message, path)).
				Value(&result),
		),
	)

	err := confirmForm.Run()
	if err != nil {
		return false, "", err
	}

	if !result {
		return false, "", nil
	}

	deleteOptions := []string{
		string(DeleteStandard),
		string(DeleteSecure),
	}

	deleteForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select delete type").
				Options(
					huh.NewOptions(deleteOptions...)...,
				).
				Value(&deleteType),
		),
	).WithTheme(huh.ThemeDracula())

	err = deleteForm.Run()
	if err != nil {
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

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Operation:").
				Options(
					huh.NewOptions(operationOptions...)...,
				).
				Value(&operationType),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}

	return ProcessorMode(operationType), nil
}

func (p *Prompt) ChooseFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files available for selection")
	}

	var selectedFile string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select file:").
				Options(
					huh.NewOptions(files...)...,
				).
				Value(&selectedFile),
		),
	).WithTheme(huh.ThemeDracula())

	err := form.Run()
	if err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return selectedFile, nil
}

package app

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/hambosto/hexwarden/internal/files"
	"github.com/hambosto/hexwarden/internal/header"
	"github.com/hambosto/hexwarden/internal/kdf"
	"github.com/hambosto/hexwarden/internal/ui"
	"github.com/hambosto/hexwarden/internal/worker"
)

// FileHandler defines the interface for file operations.
type FileHandler interface {
	Remove(path string, option ui.DeleteOption) error
	CreateFile(path string) (*os.File, error)
	ValidatePath(path string, mustExist bool) error
	OpenFile(path string) (*os.File, os.FileInfo, error)
}

// UserInterface defines the interface for user interactions.
type UserInterface interface {
	ConfirmFileOverwrite(path string) (bool, error)
	GetEncryptionPassword() (string, error)
	ConfirmFileRemoval(path string, message string) (bool, ui.DeleteOption, error)
	GetProcessingMode() (ui.ProcessorMode, error)
	ChooseFile(files []string) (string, error)
}

// Config holds the configuration for processing operations.
type Config struct {
	SourcePath      string
	DestinationPath string
	Password        string
	Mode            ui.ProcessorMode
}

// App encapsulates the application logic with its dependencies.
type App struct {
	files FileHandler
	ui    UserInterface
}

// New creates a new App instance with the provided dependencies.
func NewApp(files FileHandler, ui UserInterface) *App {
	return &App{
		files: files,
		ui:    ui,
	}
}

// Process handles the main processing logic based on the provided configuration.
func (a *App) Process(cfg Config) error {
	if err := a.validate(cfg); err != nil {
		return err
	}

	switch cfg.Mode {
	case ui.ModeEncrypt:
		return a.encrypt(cfg)
	case ui.ModeDecrypt:
		return a.decrypt(cfg)
	default:
		return fmt.Errorf("unsupported operation mode: %s", cfg.Mode)
	}
}

// ProcessFile is a convenience method that processes a single file.
func (a *App) ProcessFile(inputPath string, mode ui.ProcessorMode) error {
	outputPath := getOutputPath(inputPath, mode)

	cfg := Config{
		SourcePath:      inputPath,
		DestinationPath: outputPath,
		Mode:            mode,
	}

	return a.Process(cfg)
}

// encrypt handles the encryption workflow.
func (a *App) encrypt(cfg Config) error {
	srcFile, srcInfo, err := a.files.OpenFile(cfg.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	destFile, err := a.files.CreateFile(cfg.DestinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close destination file: %v\n", closeErr)
		}
	}()

	password := cfg.Password
	if password == "" {
		password, err = a.ui.GetEncryptionPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	key, salt, err := deriveKey(password)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	fmt.Printf("Encrypting %s...\n", cfg.SourcePath)

	if err := a.runEncryption(srcFile, destFile, srcInfo, key, salt); err != nil {
		// Clean up incomplete file on error
		if removeErr := os.Remove(cfg.DestinationPath); removeErr != nil {
			return fmt.Errorf("encryption failed and couldn't clean up incomplete file: %v (original error: %w)", removeErr, err)
		}
		return fmt.Errorf("encryption failed: %w", err)
	}

	if err := a.cleanupSource(cfg.SourcePath, true); err != nil {
		return fmt.Errorf("source cleanup failed: %w", err)
	}

	fmt.Printf("File encrypted successfully: %s\n", cfg.DestinationPath)
	return nil
}

// decrypt handles the decryption workflow.
func (a *App) decrypt(cfg Config) error {
	srcFile, _, err := a.files.OpenFile(cfg.SourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	hdr, err := header.Read(srcFile)
	if err != nil {
		return fmt.Errorf("header reading failed: %w", err)
	}

	password := cfg.Password
	if password == "" {
		password, err = a.ui.GetEncryptionPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	key, err := kdf.DeriveKey([]byte(password), hdr.Salt)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	if !hdr.VerifyPassword(key) {
		return fmt.Errorf("password verification failed")
	}

	if hdr.OriginalSize > math.MaxInt64 {
		return fmt.Errorf("file too large: size exceeds maximum allowed value for processing")
	}

	destFile, err := a.files.CreateFile(cfg.DestinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close destination file: %v\n", closeErr)
		}
	}()

	fmt.Printf("Decrypting %s...\n", cfg.SourcePath)

	if err := a.runDecryption(srcFile, destFile, key, *hdr); err != nil {
		// Clean up incomplete file on error
		if removeErr := os.Remove(cfg.DestinationPath); removeErr != nil {
			return fmt.Errorf("decryption failed and couldn't clean up incomplete file: %v (original error: %w)", removeErr, err)
		}
		return fmt.Errorf("decryption failed: %w", err)
	}

	if err := a.cleanupSource(cfg.SourcePath, false); err != nil {
		return fmt.Errorf("source cleanup failed: %w", err)
	}

	fmt.Printf("File decrypted successfully: %s\n", cfg.DestinationPath)
	return nil
}

// runEncryption performs the actual encryption process.
func (a *App) runEncryption(src, dest *os.File, srcInfo os.FileInfo, key, salt []byte) error {
	w, err := worker.New(worker.Config{
		Key:        key,
		Processing: worker.Encryption,
	})
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	fileSize := srcInfo.Size()
	if fileSize < 0 {
		return fmt.Errorf("invalid file size: %d", fileSize)
	}

	hdr, err := header.New(salt, uint64(fileSize), key)
	if err != nil {
		return fmt.Errorf("header creation failed: %w", err)
	}

	if err := hdr.Write(dest); err != nil {
		return fmt.Errorf("header writing failed: %w", err)
	}

	if err := w.Process(src, dest, fileSize); err != nil {
		return fmt.Errorf("encryption processing failed: %w", err)
	}

	return nil
}

// runDecryption performs the actual decryption process.
func (a *App) runDecryption(src, dest *os.File, key []byte, hdr header.Header) error {
	w, err := worker.New(worker.Config{
		Key:        key,
		Processing: worker.Decryption,
	})
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	originalSize := int64(hdr.OriginalSize)
	if err := w.Process(src, dest, originalSize); err != nil {
		return fmt.Errorf("decryption processing failed: %w", err)
	}

	return nil
}

// validate ensures the configuration is valid and handles file overwrite confirmation.
func (a *App) validate(cfg Config) error {
	if err := a.files.ValidatePath(cfg.SourcePath, true); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	if err := a.files.ValidatePath(cfg.DestinationPath, false); err != nil {
		confirm, confirmErr := a.ui.ConfirmFileOverwrite(cfg.DestinationPath)
		if confirmErr != nil {
			return fmt.Errorf("overwrite confirmation failed: %w", confirmErr)
		}
		if !confirm {
			return fmt.Errorf("operation cancelled by user")
		}
	}

	return nil
}

// cleanupSource handles the cleanup of source files after processing.
func (a *App) cleanupSource(path string, isEncryption bool) error {
	fileType := "original"
	if !isEncryption {
		fileType = "encrypted"
	}

	shouldDelete, deleteType, err := a.ui.ConfirmFileRemoval(path, fmt.Sprintf("Delete %s file", fileType))
	if err != nil {
		return fmt.Errorf("deletion prompt failed: %w", err)
	}

	if shouldDelete {
		if err := a.files.Remove(path, deleteType); err != nil {
			return fmt.Errorf("file deletion failed: %w", err)
		}
	}

	return nil
}

// deriveKey generates a salt and derives a key from the password.
func deriveKey(password string) (key, salt []byte, err error) {
	salt, err = kdf.GenerateSalt()
	if err != nil {
		return nil, nil, fmt.Errorf("salt generation failed: %w", err)
	}

	key, err = kdf.DeriveKey([]byte(password), salt)
	if err != nil {
		return nil, nil, fmt.Errorf("key derivation failed: %w", err)
	}

	return key, salt, nil
}

// getOutputPath determines the output file path based on the input path and mode.
func getOutputPath(inputPath string, mode ui.ProcessorMode) string {
	if mode == ui.ModeEncrypt {
		return inputPath + files.FileExtension
	}
	return strings.TrimSuffix(inputPath, files.FileExtension)
}

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

// App handles file encryption and decryption operations.
type App struct {
	fileManager *files.FileManager
	prompt      *ui.Prompt
}

// New creates a new App instance.
func NewApp(fm *files.FileManager, p *ui.Prompt) *App {
	return &App{
		fileManager: fm,
		prompt:      p,
	}
}

// ProcessFile encrypts or decrypts a file based on the mode.
func (a *App) ProcessFile(inputPath string, mode ui.ProcessorMode) error {
	outputPath := getOutputPath(inputPath, mode)

	// Validate paths
	if err := a.fileManager.ValidatePath(inputPath, true); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	if err := a.fileManager.ValidatePath(outputPath, false); err != nil {
		if confirm, err := a.prompt.ConfirmFileOverwrite(outputPath); err != nil || !confirm {
			return fmt.Errorf("operation cancelled")
		}
	}

	var err error
	switch mode {
	case ui.ModeEncrypt:
		err = a.encryptFile(inputPath, outputPath)
	case ui.ModeDecrypt:
		err = a.decryptFile(inputPath, outputPath)
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	if err != nil {
		if err := os.Remove(outputPath); err != nil {
			return fmt.Errorf("failed to remove incomplete file: %w", err)
		}
		return err
	}

	// Ask to delete source file
	fileType := "original"
	if mode == ui.ModeDecrypt {
		fileType = "encrypted"
	}

	if shouldDelete, deleteType, err := a.prompt.ConfirmFileRemoval(inputPath, fmt.Sprintf("Delete %s file", fileType)); err == nil && shouldDelete {
		if err := a.fileManager.Remove(inputPath, deleteType); err != nil {
			return fmt.Errorf("file deletion failed: %w", err)
		}
	}

	fmt.Printf("File processed successfully: %s\n", outputPath)
	return nil
}

// encryptFile handles file encryption.
func (a *App) encryptFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	password, err := a.prompt.GetEncryptionPassword()
	if err != nil {
		return fmt.Errorf("password prompt failed: %w", err)
	}

	salt, err := kdf.GenerateSalt()
	if err != nil {
		return fmt.Errorf("salt generation failed: %w", err)
	}

	key, err := kdf.DeriveKey([]byte(password), salt)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	fmt.Printf("Encrypting %s...\n", srcPath)

	// Write header
	hdr, err := header.New(salt, uint64(srcInfo.Size()), key)
	if err != nil {
		return fmt.Errorf("header creation failed: %w", err)
	}

	if err := hdr.Write(destFile); err != nil {
		return fmt.Errorf("header writing failed: %w", err)
	}

	// Encrypt data
	w, err := worker.New(worker.Config{Key: key, Processing: worker.Encryption})
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	return w.Process(srcFile, destFile, srcInfo.Size())
}

// decryptFile handles file decryption.
func (a *App) decryptFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	hdr, err := header.Read(srcFile)
	if err != nil {
		return fmt.Errorf("header reading failed: %w", err)
	}

	password, err := a.prompt.GetEncryptionPassword()
	if err != nil {
		return fmt.Errorf("password prompt failed: %w", err)
	}

	key, err := kdf.DeriveKey([]byte(password), hdr.Salt)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	if !hdr.VerifyPassword(key) {
		return fmt.Errorf("password verification failed")
	}

	if hdr.OriginalSize > math.MaxInt64 {
		return fmt.Errorf("file too large")
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer destFile.Close()

	fmt.Printf("Decrypting %s...\n", srcPath)

	// Decrypt data
	w, err := worker.New(worker.Config{Key: key, Processing: worker.Decryption})
	if err != nil {
		return fmt.Errorf("failed to create worker: %w", err)
	}

	return w.Process(srcFile, destFile, int64(hdr.OriginalSize))
}

// getOutputPath determines output path based on mode.
func getOutputPath(inputPath string, mode ui.ProcessorMode) string {
	if mode == ui.ModeEncrypt {
		return inputPath + files.FileExtension
	}
	return strings.TrimSuffix(inputPath, files.FileExtension)
}

package core

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/hambosto/hexwarden/internal/header"
	"github.com/hambosto/hexwarden/internal/kdf"
	"github.com/hambosto/hexwarden/internal/ui"
	"github.com/hambosto/hexwarden/internal/worker"
)

type FileProcessor struct {
	files FileHandler
	ui    UserInteraction
}

func NewFileProcessor(files FileHandler, ui UserInteraction) *FileProcessor {
	return &FileProcessor{
		files: files,
		ui:    ui,
	}
}

func (p *FileProcessor) Process(cfg Config) error {
	if err := p.validate(cfg); err != nil {
		return err
	}

	switch cfg.Mode {
	case ui.ModeEncrypt:
		return p.encrypt(cfg)
	case ui.ModeDecrypt:
		return p.decrypt(cfg)
	default:
		return fmt.Errorf("unsupported operation mode: %s", cfg.Mode)
	}
}

func (p *FileProcessor) ProcessFile(inputPath string, mode ui.ProcessorMode) error {
	outputPath := getOutputPath(inputPath, mode)

	cfg := Config{
		SourcePath:      inputPath,
		DestinationPath: outputPath,
		Mode:            mode,
	}

	return p.Process(cfg)
}

func (p *FileProcessor) validate(cfg Config) error {
	if err := p.files.ValidatePath(cfg.SourcePath, true); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	if err := p.files.ValidatePath(cfg.DestinationPath, false); err != nil {
		confirm, err := p.ui.ConfirmFileOverwrite(cfg.DestinationPath)
		if err != nil {
			return fmt.Errorf("overwrite confirmation failed: %w", err)
		}
		if !confirm {
			return fmt.Errorf("operation cancelled by user")
		}
	}

	return nil
}

func (p *FileProcessor) encrypt(cfg Config) error {
	srcFile, srcInfo, err := p.files.OpenFile(cfg.SourcePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := p.files.CreateFile(cfg.DestinationPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	password := cfg.Password
	if password == "" {
		password, err = p.ui.GetEncryptionPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	key, salt, err := deriveKey(password)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypting %s...\n", cfg.SourcePath)

	if err = p.runEncryption(srcFile, destFile, srcInfo, key, salt); err != nil {
		if err := destFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close destination file: %v\n", err)
		}
		if err := os.Remove(cfg.DestinationPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove incomplete file: %v\n", err)
		}
		return err
	}

	if err = p.cleanupSource(cfg.SourcePath, true); err != nil {
		return err
	}

	fmt.Printf("File encrypted successfully: %s\n", cfg.DestinationPath)
	return nil
}

func (p *FileProcessor) decrypt(cfg Config) error {
	srcFile, _, err := p.files.OpenFile(cfg.SourcePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	header, err := header.ReadHeader(srcFile)
	if err != nil {
		return fmt.Errorf("header reading failed: %w", err)
	}

	password := cfg.Password
	if password == "" {
		password, err = p.ui.GetEncryptionPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	kdfHandler, err := kdf.NewDeriver(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize key derivation: %v", err)
	}

	key, err := kdfHandler.DeriveKey([]byte(password), header.Salt)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	destFile, err := p.files.CreateFile(cfg.DestinationPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	fmt.Printf("Decrypting %s...\n", cfg.SourcePath)

	if header.OriginalSize > uint64(math.MaxInt64) {
		if err := destFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close destination file: %v\n", err)
		}
		if err := os.Remove(cfg.DestinationPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove incomplete file: %v\n", err)
		}
		return fmt.Errorf("file too large: size exceeds maximum allowed value for processing")
	}

	if err = p.runDecryption(srcFile, destFile, key, header); err != nil {
		if err := destFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close destination file: %v\n", err)
		}
		if err := os.Remove(cfg.DestinationPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove incomplete file: %v\n", err)
		}
		return err
	}

	if err = p.cleanupSource(cfg.SourcePath, false); err != nil {
		return err
	}

	fmt.Printf("File decrypted successfully: %s\n", cfg.DestinationPath)
	return nil
}

func (p *FileProcessor) runEncryption(src *os.File, dest *os.File, srcInfo os.FileInfo, key []byte, salt []byte) error {
	worker, err := worker.New(key, worker.Encryption)
	if err != nil {
		return fmt.Errorf("encryption setup failed: %w", err)
	}

	fileSize := srcInfo.Size()
	if fileSize < 0 {
		return fmt.Errorf("invalid file size: %d", fileSize)
	}

	h, err := header.NewHeader(salt, uint64(fileSize), worker.GetCipherNonce())
	if err != nil {
		return fmt.Errorf("header creation failed: %w", err)
	}

	if err = header.WriteHeader(dest, h); err != nil {
		return fmt.Errorf("header writing failed: %w", err)
	}

	if err = worker.Process(src, dest, srcInfo.Size()); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return nil
}

func (p *FileProcessor) runDecryption(src *os.File, dest *os.File, key []byte, header header.Header) error {
	worker, err := worker.New(key, worker.Decryption)
	if err != nil {
		return fmt.Errorf("decryption setup failed: %w", err)
	}

	if err := worker.SetCipherNonce(header.Nonce); err != nil {
		return fmt.Errorf("failed to set cipher nonce: %w", err)
	}

	if header.OriginalSize > math.MaxInt64 {
		return fmt.Errorf("file size exceeds maximum allowed value for decryption")
	}

	originalSize := int64(header.OriginalSize)
	if err := worker.Process(src, dest, originalSize); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return nil
}

func (p *FileProcessor) cleanupSource(path string, isEncryption bool) error {
	fileType := "original"
	if !isEncryption {
		fileType = "encrypted"
	}

	shouldDelete, deleteType, err := p.ui.ConfirmFileRemoval(
		path,
		fmt.Sprintf("Delete %s file", fileType),
	)
	if err != nil {
		return fmt.Errorf("deletion prompt failed: %w", err)
	}

	if shouldDelete {
		if err := p.files.Remove(path, deleteType); err != nil {
			return fmt.Errorf("file deletion failed: %w", err)
		}
	}

	return nil
}

func deriveKey(password string) ([]byte, []byte, error) {
	deriver, err := kdf.NewDeriver(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("key derivation setup failed: %v", err)
	}

	salt, err := deriver.GenerateSalt()
	if err != nil {
		return nil, nil, fmt.Errorf("salt generation failed: %w", err)
	}

	key, err := deriver.DeriveKey([]byte(password), salt)
	if err != nil {
		return nil, nil, fmt.Errorf("key derivation failed: %w", err)
	}

	return key, salt, nil
}

func getOutputPath(inputPath string, mode ui.ProcessorMode) string {
	if mode == ui.ModeEncrypt {
		return inputPath + ui.FileExtension
	}
	return strings.TrimSuffix(inputPath, ui.FileExtension)
}

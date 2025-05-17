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

type Processor struct {
	files    FileHandler
	interact UserInteraction
}

func NewProcessor(files FileHandler, interact UserInteraction) *Processor {
	return &Processor{
		files:    files,
		interact: interact,
	}
}

func (p *Processor) Process(process Process) error {
	if err := p.validateConfig(process); err != nil {
		return err
	}

	switch process.Mode {
	case ui.ModeEncrypt:
		return p.encryptFile(process)
	case ui.ModeDecrypt:
		return p.decryptFile(process)
	default:
		return fmt.Errorf("unsupported operation mode: %s", process.Mode)
	}
}

func (p *Processor) ProcessFile(inputPath string, mode ui.ProcessorMode) error {
	outputPath := calculateOutputPath(inputPath, mode)

	config := Process{
		SourcePath:      inputPath,
		DestinationPath: outputPath,
		Mode:            mode,
	}

	return p.Process(config)
}

func (p *Processor) validateConfig(process Process) error {
	if err := p.files.ValidatePath(process.SourcePath, true); err != nil {
		return fmt.Errorf("source validation failed: %w", err)
	}

	if err := p.files.ValidatePath(process.DestinationPath, false); err != nil {
		confirm, promptErr := p.interact.ConfirmFileOverwrite(process.DestinationPath)
		if promptErr != nil {
			return fmt.Errorf("overwrite confirmation failed: %w", promptErr)
		}
		if !confirm {
			return fmt.Errorf("operation cancelled by user")
		}
	}

	return nil
}

func (p *Processor) encryptFile(process Process) error {
	sourceFile, sourceInfo, err := p.files.OpenFile(process.SourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := p.files.CreateFile(process.DestinationPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	password := process.Password
	if password == "" {
		password, err = p.interact.GetEncryptionPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	key, salt, err := deriveEncryptionKey(password)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypting %s...\n", process.SourcePath)

	if err = p.performEncryption(sourceFile, destFile, sourceInfo, key, salt); err != nil {
		if closeErr := destFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close destination file: %v\n", closeErr)
		}
		if rmErr := os.Remove(process.DestinationPath); rmErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove incomplete file: %v\n", rmErr)
		}
		return err
	}

	if err = p.handleSourceCleanup(process.SourcePath, true); err != nil {
		return err
	}

	fmt.Printf("File encrypted successfully: %s\n", process.DestinationPath)
	return nil
}

func (p *Processor) decryptFile(process Process) error {
	sourceFile, _, err := p.files.OpenFile(process.SourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	fileHeader, err := header.ReadHeader(sourceFile)
	if err != nil {
		return fmt.Errorf("header reading failed: %w", err)
	}

	password := process.Password
	if password == "" {
		password, err = p.interact.GetEncryptionPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	kdfHandler, err := kdf.NewDeriver(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize key derivation: %v", err)
	}

	key, err := kdfHandler.DeriveKey([]byte(password), fileHeader.Salt)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	destFile, err := p.files.CreateFile(process.DestinationPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	fmt.Printf("Decrypting %s...\n", process.SourcePath)

	if fileHeader.OriginalSize > uint64(math.MaxInt64) {
		if closeErr := destFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close destination file: %v\n", closeErr)
		}
		if rmErr := os.Remove(process.DestinationPath); rmErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove incomplete file: %v\n", rmErr)
		}
		return fmt.Errorf("file too large: size exceeds maximum allowed value for processing")
	}

	if err = p.performDecryption(sourceFile, destFile, key, fileHeader); err != nil {
		if closeErr := destFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close destination file: %v\n", closeErr)
		}
		if rmErr := os.Remove(process.DestinationPath); rmErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to remove incomplete file: %v\n", rmErr)
		}
		return err
	}

	if err = p.handleSourceCleanup(process.SourcePath, false); err != nil {
		return err
	}

	fmt.Printf("File decrypted successfully: %s\n", process.DestinationPath)
	return nil
}

func (p *Processor) performEncryption(source *os.File, dest *os.File, sourceInfo os.FileInfo, key []byte, salt []byte) error {
	processor, err := worker.New(key, worker.Encryption)
	if err != nil {
		return fmt.Errorf("encryption setup failed: %w", err)
	}

	fileSize := sourceInfo.Size()
	if fileSize < 0 {
		return fmt.Errorf("invalid file size: %d", fileSize)
	}

	h, err := header.NewHeader(salt, uint64(fileSize), processor.GetCipherNonce())
	if err != nil {
		return fmt.Errorf("header creation failed: %w", err)
	}

	if err = header.WriteHeader(dest, h); err != nil {
		return fmt.Errorf("header writing failed: %w", err)
	}

	if err = processor.Process(source, dest, sourceInfo.Size()); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return nil
}

func (p *Processor) performDecryption(source *os.File, dest *os.File, key []byte, fileHeader header.Header) error {
	processor, err := worker.New(key, worker.Decryption)
	if err != nil {
		return fmt.Errorf("decryption setup failed: %w", err)
	}

	if err := processor.SetCipherNonce(fileHeader.Nonce); err != nil {
		return fmt.Errorf("failed to set cipher nonce: %w", err)
	}

	if fileHeader.OriginalSize > math.MaxInt64 {
		return fmt.Errorf("file size exceeds maximum allowed value for decryption")
	}

	originalSize := int64(fileHeader.OriginalSize)
	if err := processor.Process(source, dest, originalSize); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return nil
}

func (p *Processor) handleSourceCleanup(path string, isEncryption bool) error {
	fileType := "original"
	if !isEncryption {
		fileType = "encrypted"
	}

	shouldDelete, deleteType, err := p.interact.ConfirmFileRemoval(
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

func deriveEncryptionKey(password string) ([]byte, []byte, error) {
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

func calculateOutputPath(inputPath string, mode ui.ProcessorMode) string {
	if mode == ui.ModeEncrypt {
		return inputPath + ui.FileExtension
	}
	return strings.TrimSuffix(inputPath, ui.FileExtension)
}

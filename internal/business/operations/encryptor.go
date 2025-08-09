package operations

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/data/files"
	"github.com/hambosto/hexwarden/internal/data/streaming"
	"github.com/hambosto/hexwarden/internal/infrastructure/crypto"
)

// Encryptor handles file encryption operations
type Encryptor struct {
	fileManager *files.Manager
	fileFinder  *files.Finder
}

// NewEncryptor creates a new encryptor instance
func NewEncryptor() *Encryptor {
	return &Encryptor{
		fileManager: files.NewManager(),
		fileFinder:  files.NewFinder(),
	}
}

// EncryptFile encrypts a file from source to destination
func (e *Encryptor) EncryptFile(srcPath, destPath, password string, progressCallback func(int64)) error {
	// Open source file
	srcFile, srcInfo, err := e.fileManager.OpenFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := e.fileManager.CreateFile(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Generate salt for key derivation
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password
	key, err := crypto.DeriveKey([]byte(password), salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	// Validate file size
	originalSize := srcInfo.Size()
	if originalSize < 0 {
		return fmt.Errorf("invalid file size: %d", originalSize)
	}

	// Create and write header
	header, err := crypto.NewHeader(salt, uint64(originalSize), key)
	if err != nil {
		return fmt.Errorf("failed to create header: %w", err)
	}

	if err := header.Write(destFile); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Create stream processor for encryption
	config := streaming.StreamConfig{
		Key:         key,
		Processing:  constants.Encryption,
		Concurrency: constants.MaxConcurrency,
		QueueSize:   constants.QueueSize,
		ChunkSize:   constants.DefaultChunkSize,
	}

	processor, err := streaming.NewStreamProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create stream processor: %w", err)
	}

	// Process the file
	return processor.Process(srcFile, destFile, originalSize, progressCallback)
}

package operations

import (
	"fmt"
	"math"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/data/files"
	"github.com/hambosto/hexwarden/internal/data/streaming"
	"github.com/hambosto/hexwarden/internal/infrastructure/crypto"
)

// Decryptor handles file decryption operations
type Decryptor struct {
	fileManager *files.Manager
	fileFinder  *files.Finder
}

// NewDecryptor creates a new decryptor instance
func NewDecryptor() *Decryptor {
	return &Decryptor{
		fileManager: files.NewManager(),
		fileFinder:  files.NewFinder(),
	}
}

// DecryptFile decrypts a file from source to destination
func (d *Decryptor) DecryptFile(srcPath, destPath, password string, progressCallback func(int64)) error {
	// Open source file
	srcFile, _, err := d.fileManager.OpenFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Read and parse header
	header, err := crypto.ReadHeader(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Derive key from password and verify
	key, err := crypto.DeriveKey([]byte(password), header.Salt())
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	if err := header.VerifyKey(key); err != nil {
		return fmt.Errorf("header verification failed: %w", err)
	}

	// Validate original size
	originalSize := header.OriginalSize()
	if originalSize > math.MaxInt64 {
		return fmt.Errorf("file too large: %d bytes", originalSize)
	}

	// Create destination file
	destFile, err := d.fileManager.CreateFile(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Create stream processor for decryption
	config := streaming.StreamConfig{
		Key:         key,
		Processing:  constants.Decryption,
		Concurrency: constants.MaxConcurrency,
		QueueSize:   constants.QueueSize,
		ChunkSize:   constants.DefaultChunkSize,
	}

	processor, err := streaming.NewStreamProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create stream processor: %w", err)
	}

	// Process the file (remaining data after header)
	return processor.Process(srcFile, destFile, int64(originalSize), progressCallback)
}

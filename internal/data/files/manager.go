package files

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/utils"
)

// Manager handles file creation, deletion (standard and secure), and validation
type Manager struct{}

// NewManager creates a new file manager instance
func NewManager() *Manager {
	return &Manager{}
}

// Remove deletes the file at the given path using the provided deletion option
func (m *Manager) Remove(path string, option constants.DeleteOption) error {
	switch option {
	case constants.DeleteStandard:
		return os.Remove(path)
	case constants.DeleteSecure:
		return m.secureDelete(path)
	default:
		return fmt.Errorf("unsupported delete option: %s", option)
	}
}

// CreateFile creates and returns a new file at the given path
func (m *Manager) CreateFile(path string) (*os.File, error) {
	output, err := os.Create(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", constants.ErrFileCreateFailed, err)
	}
	return output, nil
}

// ValidatePath checks whether a file at the given path should or should not exist
func (m *Manager) ValidatePath(path string, mustExist bool) error {
	fileInfo, err := os.Stat(path)

	if mustExist {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", constants.ErrFileNotFound, path)
		}
		if err != nil {
			return fmt.Errorf("%w: %v", constants.ErrFileOpenFailed, err)
		}
		if fileInfo.Size() == 0 {
			return fmt.Errorf("%w: %s", constants.ErrFileEmpty, path)
		}
	} else {
		if err == nil {
			return fmt.Errorf("%w: %s", constants.ErrFileExists, path)
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("error accessing file: %w", err)
		}
	}
	return nil
}

// OpenFile opens a file and returns both the file handle and its metadata
func (m *Manager) OpenFile(path string) (*os.File, os.FileInfo, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", constants.ErrFileOpenFailed, err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return file, info, nil
}

// GetFileInfo returns file information without opening the file
func (m *Manager) GetFileInfo(path string) (os.FileInfo, error) {
	info, err := os.Stat(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", constants.ErrFileNotFound, path)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return info, nil
}

// FileExists checks if a file exists at the given path
func (m *Manager) FileExists(path string) bool {
	_, err := os.Stat(filepath.Clean(path))
	return err == nil
}

// secureDelete securely deletes a file by overwriting its contents with random data
func (m *Manager) secureDelete(path string) error {
	file, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("%w: failed to open file for secure deletion: %v", constants.ErrSecureDeleteFailed, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("%w: failed to get file info: %v", constants.ErrSecureDeleteFailed, err)
	}

	// Perform multiple overwrite passes
	for pass := 0; pass < constants.OverwritePasses; pass++ {
		if err := m.randomOverwrite(file, info.Size()); err != nil {
			return fmt.Errorf("%w: secure overwrite pass %d failed: %v", constants.ErrSecureDeleteFailed, pass+1, err)
		}
	}

	// Close the file before removing it
	file.Close()

	// Finally remove the file
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("%w: failed to remove file: %v", constants.ErrSecureDeleteFailed, err)
	}

	return nil
}

// randomOverwrite writes cryptographically secure random bytes over the file content
func (m *Manager) randomOverwrite(file *os.File, size int64) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to file start: %w", err)
	}

	buffer := make([]byte, 4096)
	remaining := size

	for remaining > 0 {
		writeSize := utils.MinInt64(remaining, int64(len(buffer)))

		if _, err := rand.Read(buffer[:writeSize]); err != nil {
			return fmt.Errorf("failed to generate random data: %w", err)
		}

		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return fmt.Errorf("failed to write random data: %w", err)
		}

		remaining -= writeSize
	}

	return file.Sync()
}

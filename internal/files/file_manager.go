package files

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hambosto/hexwarden/internal/ui"
)

// FileManager handles file creation, deletion (standard and secure), and validation.
type FileManager struct {
	overwritePasses int
}

// NewFileManager returns a new FileManager with the specified number of secure overwrite passes.
// If the provided value is non-positive, a default of 3 passes is used.
func NewFileManager(overwritePasses int) *FileManager {
	if overwritePasses <= 0 {
		overwritePasses = 3
	}
	return &FileManager{
		overwritePasses: overwritePasses,
	}
}

// Remove deletes the file at the given path using the provided deletion option.
// It supports standard deletion and secure deletion (multiple overwrites).
func (fm *FileManager) Remove(path string, option ui.DeleteOption) error {
	switch option {
	case ui.DeleteStandard:
		return os.Remove(path)
	case ui.DeleteSecure:
		return secureDelete(path, fm.overwritePasses)
	default:
		return fmt.Errorf("unsupported delete option: %s", option)
	}
}

// CreateFile creates and returns a new file at the given path.
// The path is sanitized before use.
func (fm *FileManager) CreateFile(path string) (*os.File, error) {
	output, err := os.Create(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return output, nil
}

// ValidatePath checks whether a file at the given path should or should not exist.
// If mustExist is true, it checks for existence and non-zero size.
// If mustExist is false, it ensures the file does not exist.
func (fm *FileManager) ValidatePath(path string, mustExist bool) error {
	fileInfo, err := os.Stat(path)

	if mustExist {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		if fileInfo.Size() == 0 {
			return fmt.Errorf("file is empty: %s", path)
		}
	} else {
		if err == nil {
			return fmt.Errorf("file already exists: %s", path)
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("error accessing file: %w", err)
		}
	}
	return nil
}

// OpenFile opens a file and returns both the file handle and its metadata.
func (fm *FileManager) OpenFile(path string) (*os.File, os.FileInfo, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return file, info, nil
}

// secureDelete securely deletes a file by overwriting its contents with random data
// for the given number of passes before removing the file.
func secureDelete(path string, passes int) error {
	file, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file for secure deletion: %w", err)
	}
	defer file.Close() //nolint:errcheck

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	for pass := range passes {
		if err := randomOverwrite(file, info.Size()); err != nil {
			return fmt.Errorf("secure overwrite pass %d failed: %w", pass+1, err)
		}
	}

	return os.Remove(path)
}

// randomOverwrite writes cryptographically secure random bytes over the file content.
func randomOverwrite(file *os.File, size int64) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to file start: %w", err)
	}

	buffer := make([]byte, 4096)
	remaining := size

	for remaining > 0 {
		writeSize := min(remaining, int64(len(buffer)))

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

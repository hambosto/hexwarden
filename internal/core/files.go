package core

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hambosto/hexwarden/internal/ui"
)

type FileHandler interface {
	Remove(path string, option ui.DeleteOption) error
	CreateFile(path string) (*os.File, error)
	ValidatePath(path string, mustExist bool) error
	OpenFile(path string) (*os.File, os.FileInfo, error)
}

type UserInteraction interface {
	ConfirmFileOverwrite(path string) (bool, error)
	GetEncryptionPassword() (string, error)
	ConfirmFileRemoval(path string, message string) (bool, ui.DeleteOption, error)
	GetProcessingMode() (ui.ProcessorMode, error)
	ChooseFile(files []string) (string, error)
}

type Config struct {
	SourcePath      string
	DestinationPath string
	Password        string
	Mode            ui.ProcessorMode
}

type FileManager struct {
	overwritePasses int
}

func NewFileManager(overwritePasses int) *FileManager {
	if overwritePasses <= 0 {
		overwritePasses = 3
	}
	return &FileManager{
		overwritePasses: 3,
	}
}

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

func (fm *FileManager) CreateFile(path string) (*os.File, error) {
	output, err := os.Create(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return output, nil
}

func (fm *FileManager) ValidatePath(path string, mustExist bool) error {
	fileInfo, err := os.Stat(path)
	if mustExist {
		if os.IsExist(err) {
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
		return nil
	}
	return nil
}

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

func secureDelete(path string, passes int) error {
	file, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file for secure deletion: %w", err)
	}
	defer file.Close()

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

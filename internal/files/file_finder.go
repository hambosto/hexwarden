package files

import (
	"os"
	"path/filepath"
	"strings"
)

// FileExtension defines the suffix used for encrypted files.
const (
	FileExtension = ".hex"
)

// ProcessorMode represents the operation mode of the file processor.
type ProcessorMode string

const (
	// ModeEncrypt indicates files should be encrypted.
	ModeEncrypt ProcessorMode = "Encrypt"
	// ModeDecrypt indicates files should be decrypted.
	ModeDecrypt ProcessorMode = "Decrypt"
)

// FileFinder is responsible for finding files eligible for processing,
// while excluding specified directories and file extensions.
type FileFinder struct {
	excludedDirs []string
	excludedExts []string
}

// NewFileFinder creates a new FileFinder with given excluded directories and file extensions.
func NewFileFinder(excludedDirs, excludedExts []string) *FileFinder {
	return &FileFinder{
		excludedDirs: excludedDirs,
		excludedExts: excludedExts,
	}
}

// FindEligibleFiles walks the current directory tree and returns a list of files
// eligible for encryption or decryption, based on the specified mode.
func (f *FileFinder) FindEligibleFiles(mode ProcessorMode) ([]string, error) {
	var files []string

	// Walk through all files and directories starting from current directory.
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.isFileEligible(path, info, mode) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// isFileEligible checks if a given file should be processed based on its extension,
// whether it's hidden, or excluded via config, and depending on the selected mode.
func (f *FileFinder) isFileEligible(path string, info os.FileInfo, mode ProcessorMode) bool {
	// Skip directories, hidden files, or excluded paths.
	if info.IsDir() || strings.HasPrefix(info.Name(), ".") || f.shouldSkipPath(path) {
		return false
	}

	// Check file suffix to determine if it's encrypted.
	isEncrypted := strings.HasSuffix(path, FileExtension)

	// Include unencrypted files in encrypt mode, and encrypted files in decrypt mode.
	return (mode == ModeEncrypt && !isEncrypted) || (mode == ModeDecrypt && isEncrypted)
}

// shouldSkipPath returns true if the file should be excluded based on directory or extension rules.
func (f *FileFinder) shouldSkipPath(path string) bool {
	for _, dir := range f.excludedDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}

	for _, ext := range f.excludedExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

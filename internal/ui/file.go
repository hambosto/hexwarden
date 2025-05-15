package ui

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	FileExtension = ".encrypted"
)

type FileFinder struct {
	excludedDirs []string
	excludedExts []string
}

func NewFileFinder() *FileFinder {
	return &FileFinder{
		excludedDirs: []string{"vendor/", "node_modules/", ".git", ".github"},   // TODO: add more directories
		excludedExts: []string{".go", "go.mod", "go.sum", ".nix", ".gitignore"}, // TODO: add more extensions
	}
}

func (f *FileFinder) FindEligibleFiles(mode ProcessorMode) ([]string, error) {
	var files []string
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

func (f *FileFinder) isFileEligible(path string, info os.FileInfo, mode ProcessorMode) bool {
	if info.IsDir() || strings.HasPrefix(info.Name(), ".") || f.shouldSkipPath(path) {
		return false
	}

	isEncrypted := strings.HasSuffix(path, FileExtension)

	return (mode == ModeEncrypt && !isEncrypted) ||
		(mode == ModeDecrypt && isEncrypted)
}

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

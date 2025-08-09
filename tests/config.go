package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfig holds configuration for test execution
type TestConfig struct {
	TempDir     string
	CleanupDirs []string
}

// NewTestConfig creates a new test configuration
func NewTestConfig() *TestConfig {
	return &TestConfig{
		CleanupDirs: make([]string, 0),
	}
}

// SetupTempDir creates a temporary directory for testing
func (c *TestConfig) SetupTempDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "hexwarden-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	c.TempDir = tmpDir
	c.CleanupDirs = append(c.CleanupDirs, tmpDir)
	return tmpDir
}

// Cleanup removes all temporary directories created during testing
func (c *TestConfig) Cleanup(t *testing.T) {
	for _, dir := range c.CleanupDirs {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("Warning: Failed to cleanup temp dir %s: %v", dir, err)
		}
	}
	c.CleanupDirs = c.CleanupDirs[:0]
}

// CreateTestFile creates a test file with the given content
func (c *TestConfig) CreateTestFile(t *testing.T, name string, content []byte) string {
	if c.TempDir == "" {
		c.SetupTempDir(t)
	}

	path := filepath.Join(c.TempDir, name)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}

	return path
}

// CreateTestFiles creates multiple test files
func (c *TestConfig) CreateTestFiles(t *testing.T, files map[string][]byte) {
	for name, content := range files {
		c.CreateTestFile(t, name, content)
	}
}

// GetTestDataDir returns the path to test data directory
func GetTestDataDir() string {
	return filepath.Join(".", "testdata")
}

// EnsureTestDataDir creates the test data directory if it doesn't exist
func EnsureTestDataDir(t *testing.T) string {
	dir := GetTestDataDir()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("Failed to create test data dir: %v", err)
	}
	return dir
}

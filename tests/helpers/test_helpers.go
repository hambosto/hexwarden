package helpers

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
)

// TestData contains common test data and utilities
type TestData struct {
	ValidKey32   []byte
	ValidKey24   []byte
	ValidKey16   []byte
	InvalidKey   []byte
	ValidSalt    []byte
	WeakSalt     []byte
	TestPassword string
	TestData     []byte
	LargeData    []byte
}

// NewTestData creates a new TestData instance with initialized test data
func NewTestData() *TestData {
	td := &TestData{
		ValidKey32:   make([]byte, 32),
		ValidKey24:   make([]byte, 24),
		ValidKey16:   make([]byte, 16),
		InvalidKey:   make([]byte, 15), // Invalid size
		ValidSalt:    make([]byte, constants.SaltSize),
		WeakSalt:     make([]byte, constants.SaltSize), // All zeros - weak
		TestPassword: "test-password-123",
		TestData:     []byte("Hello, World! This is test data for encryption."),
		LargeData:    make([]byte, 1024*1024), // 1MB of test data
	}

	// Fill with random data
	if _, err := rand.Read(td.ValidKey32); err != nil {
		panic("Failed to generate random key: " + err.Error())
	}
	if _, err := rand.Read(td.ValidKey24); err != nil {
		panic("Failed to generate random key: " + err.Error())
	}
	if _, err := rand.Read(td.ValidKey16); err != nil {
		panic("Failed to generate random key: " + err.Error())
	}
	rand.Read(td.InvalidKey)
	rand.Read(td.ValidSalt)
	rand.Read(td.LargeData)

	// WeakSalt is intentionally all zeros

	return td
}

// CreateTempFile creates a temporary file with the given content
func CreateTempFile(t *testing.T, content []byte) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "hexwarden-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.Write(content); err != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close temp file: %v", closeErr)
		}
		if removeErr := os.Remove(tmpFile.Name()); removeErr != nil {
			t.Logf("Warning: Failed to remove temp file: %v", removeErr)
		}
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		if removeErr := os.Remove(tmpFile.Name()); removeErr != nil {
			t.Logf("Warning: Failed to remove temp file: %v", removeErr)
		}
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}

// CreateTempDir creates a temporary directory for testing
func CreateTempDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "hexwarden-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	return tmpDir
}

// CleanupTempFile removes a temporary file
func CleanupTempFile(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Logf("Warning: Failed to cleanup temp file %s: %v", path, err)
	}
}

// CleanupTempDir removes a temporary directory and all its contents
func CleanupTempDir(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil {
		t.Logf("Warning: Failed to cleanup temp dir %s: %v", path, err)
	}
}

// AssertError checks that an error occurred and optionally matches expected error
func AssertError(t *testing.T, err error, expectedErr error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
	if expectedErr != nil && err != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

// AssertNoError checks that no error occurred
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}

// AssertEqual checks that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotEqual checks that two values are not equal
func AssertNotEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected == actual {
		t.Fatalf("Expected values to be different, but both were %v", expected)
	}
}

// AssertBytesEqual checks that two byte slices are equal
func AssertBytesEqual(t *testing.T, expected, actual []byte) {
	t.Helper()
	if !bytes.Equal(expected, actual) {
		t.Fatalf("Expected bytes %x, got %x", expected, actual)
	}
}

// AssertBytesNotEqual checks that two byte slices are not equal
func AssertBytesNotEqual(t *testing.T, expected, actual []byte) {
	t.Helper()
	if bytes.Equal(expected, actual) {
		t.Fatalf("Expected bytes to be different, but both were %x", expected)
	}
}

// AssertFileExists checks that a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist", path)
	}
}

// AssertFileNotExists checks that a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Expected file %s to not exist", path)
	}
}

// ReadFileContent reads the entire content of a file
func ReadFileContent(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return content
}

// WriteFileContent writes content to a file
func WriteFileContent(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// MockReader implements io.Reader for testing
type MockReader struct {
	data   []byte
	pos    int
	err    error
	closed bool
}

// NewMockReader creates a new MockReader
func NewMockReader(data []byte) *MockReader {
	return &MockReader{data: data}
}

// NewMockReaderWithError creates a MockReader that returns an error
func NewMockReaderWithError(data []byte, err error) *MockReader {
	return &MockReader{data: data, err: err}
}

// Read implements io.Reader
func (m *MockReader) Read(p []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	if m.err != nil {
		return 0, m.err
	}
	if m.pos >= len(m.data) {
		return 0, io.EOF
	}

	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

// Close simulates closing the reader
func (m *MockReader) Close() error {
	m.closed = true
	return nil
}

// MockWriter implements io.Writer for testing
type MockWriter struct {
	data   []byte
	err    error
	closed bool
}

// NewMockWriter creates a new MockWriter
func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

// NewMockWriterWithError creates a MockWriter that returns an error
func NewMockWriterWithError(err error) *MockWriter {
	return &MockWriter{err: err}
}

// Write implements io.Writer
func (m *MockWriter) Write(p []byte) (n int, err error) {
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	if m.err != nil {
		return 0, m.err
	}
	m.data = append(m.data, p...)
	return len(p), nil
}

// Data returns the written data
func (m *MockWriter) Data() []byte {
	return m.data
}

// Close simulates closing the writer
func (m *MockWriter) Close() error {
	m.closed = true
	return nil
}

// Reset clears the written data
func (m *MockWriter) Reset() {
	m.data = m.data[:0]
}

// CreateTestFiles creates a set of test files in the given directory
func CreateTestFiles(t *testing.T, dir string, files map[string][]byte) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		WriteFileContent(t, path, content)
	}
}

// ProgressTracker tracks progress updates for testing
type ProgressTracker struct {
	Updates []int64
	Total   int64
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{}
}

// Update records a progress update
func (p *ProgressTracker) Update(bytes int64) {
	p.Updates = append(p.Updates, bytes)
	p.Total += bytes
}

// GetCallback returns a callback function for progress tracking
func (p *ProgressTracker) GetCallback() func(int64) {
	return p.Update
}

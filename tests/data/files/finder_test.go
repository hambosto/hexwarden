package files

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/data/files"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestNewFinder(t *testing.T) {
	finder := files.NewFinder()
	if finder == nil {
		t.Error("Expected finder to be non-nil")
	}
}

func TestFinder_FindEligibleFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := helpers.CreateTempDir(t)
	defer helpers.CleanupTempDir(t, tmpDir)

	// Change to temp directory for testing
	originalDir, err := os.Getwd()
	helpers.AssertNoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	helpers.AssertNoError(t, err)

	// Create test files
	testFiles := map[string][]byte{
		"test1.txt": []byte("test content 1"),
		"test2.doc": []byte("test content 2"),
		"encrypted.txt" + constants.FileExtension: []byte("encrypted content"),
		"another.pdf" + constants.FileExtension:   []byte("encrypted pdf"),
		"subdir/nested.txt":                       []byte("nested content"),
		"subdir/nested" + constants.FileExtension: []byte("nested encrypted"),
		".hidden.txt": []byte("hidden file"),
		"test.go":     []byte("go source file"), // Should be excluded
		"README.md":   []byte("readme content"),
	}

	helpers.CreateTestFiles(t, tmpDir, testFiles)

	finder := files.NewFinder()

	t.Run("Encrypt mode", func(t *testing.T) {
		eligibleFiles, err := finder.FindEligibleFiles(constants.ModeEncrypt)
		helpers.AssertNoError(t, err)

		// Should find unencrypted files, excluding hidden files and excluded extensions
		expectedFiles := []string{
			"test1.txt",
			"test2.doc",
			"README.md",
			filepath.Join("subdir", "nested.txt"),
		}

		if len(eligibleFiles) != len(expectedFiles) {
			t.Errorf("Expected %d files, got %d: %v", len(expectedFiles), len(eligibleFiles), eligibleFiles)
		}

		// Check that all expected files are found
		for _, expected := range expectedFiles {
			found := slices.Contains(eligibleFiles, expected)
			if !found {
				t.Errorf("Expected file %s not found in eligible files", expected)
			}
		}
	})

	t.Run("Decrypt mode", func(t *testing.T) {
		eligibleFiles, err := finder.FindEligibleFiles(constants.ModeDecrypt)
		helpers.AssertNoError(t, err)

		// Should find encrypted files only
		expectedFiles := []string{
			"encrypted.txt" + constants.FileExtension,
			"another.pdf" + constants.FileExtension,
			filepath.Join("subdir", "nested"+constants.FileExtension),
		}

		if len(eligibleFiles) != len(expectedFiles) {
			t.Errorf("Expected %d files, got %d: %v", len(expectedFiles), len(eligibleFiles), eligibleFiles)
		}

		// Check that all expected files are found
		for _, expected := range expectedFiles {
			found := slices.Contains(eligibleFiles, expected)
			if !found {
				t.Errorf("Expected file %s not found in eligible files", expected)
			}
		}
	})
}

func TestFinder_IsEncryptedFile(t *testing.T) {
	finder := files.NewFinder()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Encrypted file",
			path:     "test.txt" + constants.FileExtension,
			expected: true,
		},
		{
			name:     "Regular file",
			path:     "test.txt",
			expected: false,
		},
		{
			name:     "File with extension in middle",
			path:     "test" + constants.FileExtension + ".backup",
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "Just extension",
			path:     constants.FileExtension,
			expected: true,
		},
		{
			name:     "Path with directory",
			path:     filepath.Join("dir", "test.txt"+constants.FileExtension),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := finder.IsEncryptedFile(tt.path)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestFinder_GetOutputPath(t *testing.T) {
	finder := files.NewFinder()

	tests := []struct {
		name     string
		input    string
		mode     constants.ProcessorMode
		expected string
	}{
		{
			name:     "Encrypt mode - add extension",
			input:    "test.txt",
			mode:     constants.ModeEncrypt,
			expected: "test.txt" + constants.FileExtension,
		},
		{
			name:     "Decrypt mode - remove extension",
			input:    "test.txt" + constants.FileExtension,
			mode:     constants.ModeDecrypt,
			expected: "test.txt",
		},
		{
			name:     "Encrypt mode with path",
			input:    filepath.Join("dir", "test.doc"),
			mode:     constants.ModeEncrypt,
			expected: filepath.Join("dir", "test.doc") + constants.FileExtension,
		},
		{
			name:     "Decrypt mode with path",
			input:    filepath.Join("dir", "test.doc") + constants.FileExtension,
			mode:     constants.ModeDecrypt,
			expected: filepath.Join("dir", "test.doc"),
		},
		{
			name:     "Decrypt mode without extension",
			input:    "test.txt",
			mode:     constants.ModeDecrypt,
			expected: "test.txt", // Should return as-is if no extension to remove
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := finder.GetOutputPath(tt.input, tt.mode)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestFinder_GetFileInfo(t *testing.T) {
	// Create temporary files for testing
	tmpDir := helpers.CreateTempDir(t)
	defer helpers.CleanupTempDir(t, tmpDir)

	testFiles := map[string][]byte{
		"small.txt": []byte("small"),
		"large.txt": make([]byte, 1000),
		"encrypted.txt" + constants.FileExtension: []byte("encrypted content"),
	}

	helpers.CreateTestFiles(t, tmpDir, testFiles)

	// Create file paths
	filePaths := []string{
		filepath.Join(tmpDir, "small.txt"),
		filepath.Join(tmpDir, "large.txt"),
		filepath.Join(tmpDir, "encrypted.txt"+constants.FileExtension),
		filepath.Join(tmpDir, "nonexistent.txt"), // This should be skipped
	}

	finder := files.NewFinder()
	fileInfos, err := finder.GetFileInfo(filePaths)
	helpers.AssertNoError(t, err)

	// Should get info for 3 files (nonexistent should be skipped)
	helpers.AssertEqual(t, 3, len(fileInfos))

	// Check file info details
	for _, info := range fileInfos {
		switch filepath.Base(info.Path) {
		case "small.txt":
			helpers.AssertEqual(t, int64(5), info.Size)
			helpers.AssertEqual(t, false, info.IsEncrypted)
			helpers.AssertEqual(t, true, info.IsEligible)
		case "large.txt":
			helpers.AssertEqual(t, int64(1000), info.Size)
			helpers.AssertEqual(t, false, info.IsEncrypted)
			helpers.AssertEqual(t, true, info.IsEligible)
		case "encrypted.txt" + constants.FileExtension:
			helpers.AssertEqual(t, int64(17), info.Size) // "encrypted content"
			helpers.AssertEqual(t, true, info.IsEncrypted)
			helpers.AssertEqual(t, true, info.IsEligible)
		default:
			t.Errorf("Unexpected file in results: %s", info.Path)
		}
	}
}

func TestFinder_ExclusionRules(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := helpers.CreateTempDir(t)
	defer helpers.CleanupTempDir(t, tmpDir)

	// Change to temp directory for testing
	originalDir, err := os.Getwd()
	helpers.AssertNoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	helpers.AssertNoError(t, err)

	// Create test files including excluded ones
	testFiles := map[string][]byte{
		"normal.txt":                []byte("normal file"),
		"main.go":                   []byte("go source"),     // Excluded extension
		"go.mod":                    []byte("go module"),     // Excluded extension
		"vendor/lib.txt":            []byte("vendor file"),   // Excluded directory
		"node_modules/package.json": []byte("node package"),  // Excluded directory
		".git/config":               []byte("git config"),    // Excluded directory
		".hidden.txt":               []byte("hidden file"),   // Hidden file
		"build/output.exe":          []byte("build output"),  // Excluded directory
		"dist/bundle.js":            []byte("dist bundle"),   // Excluded directory
		"target/classes.jar":        []byte("target jar"),    // Excluded directory
		"binary.exe":                []byte("executable"),    // Excluded extension
		"library.dll":               []byte("dll library"),   // Excluded extension
		"shared.so":                 []byte("shared object"), // Excluded extension
		"dynamic.dylib":             []byte("dynamic lib"),   // Excluded extension
	}

	helpers.CreateTestFiles(t, tmpDir, testFiles)

	finder := files.NewFinder()
	eligibleFiles, err := finder.FindEligibleFiles(constants.ModeEncrypt)
	helpers.AssertNoError(t, err)

	// Should only find normal.txt
	helpers.AssertEqual(t, 1, len(eligibleFiles))
	helpers.AssertEqual(t, "normal.txt", eligibleFiles[0])
}

func TestFinder_EmptyDirectory(t *testing.T) {
	// Create an empty temporary directory
	tmpDir := helpers.CreateTempDir(t)
	defer helpers.CleanupTempDir(t, tmpDir)

	// Change to temp directory for testing
	originalDir, err := os.Getwd()
	helpers.AssertNoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	helpers.AssertNoError(t, err)

	finder := files.NewFinder()

	t.Run("Encrypt mode - empty directory", func(t *testing.T) {
		eligibleFiles, err := finder.FindEligibleFiles(constants.ModeEncrypt)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 0, len(eligibleFiles))
	})

	t.Run("Decrypt mode - empty directory", func(t *testing.T) {
		eligibleFiles, err := finder.FindEligibleFiles(constants.ModeDecrypt)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 0, len(eligibleFiles))
	})
}

func TestFinder_NestedDirectories(t *testing.T) {
	// Create a temporary directory with nested structure
	tmpDir := helpers.CreateTempDir(t)
	defer helpers.CleanupTempDir(t, tmpDir)

	// Change to temp directory for testing
	originalDir, err := os.Getwd()
	helpers.AssertNoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Warning: Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	helpers.AssertNoError(t, err)

	// Create nested directory structure
	testFiles := map[string][]byte{
		"root.txt":                                   []byte("root file"),
		"level1/file1.txt":                           []byte("level 1 file"),
		"level1/level2/file2.txt":                    []byte("level 2 file"),
		"level1/level2/level3/file3.txt":             []byte("level 3 file"),
		"level1/encrypted" + constants.FileExtension: []byte("encrypted in level1"),
	}

	helpers.CreateTestFiles(t, tmpDir, testFiles)

	finder := files.NewFinder()

	t.Run("Find all unencrypted files", func(t *testing.T) {
		eligibleFiles, err := finder.FindEligibleFiles(constants.ModeEncrypt)
		helpers.AssertNoError(t, err)

		expectedFiles := []string{
			"root.txt",
			filepath.Join("level1", "file1.txt"),
			filepath.Join("level1", "level2", "file2.txt"),
			filepath.Join("level1", "level2", "level3", "file3.txt"),
		}

		helpers.AssertEqual(t, len(expectedFiles), len(eligibleFiles))

		// Check that all expected files are found
		for _, expected := range expectedFiles {
			found := slices.Contains(eligibleFiles, expected)
			if !found {
				t.Errorf("Expected file %s not found in eligible files", expected)
			}
		}
	})

	t.Run("Find encrypted files", func(t *testing.T) {
		eligibleFiles, err := finder.FindEligibleFiles(constants.ModeDecrypt)
		helpers.AssertNoError(t, err)

		expectedFiles := []string{
			filepath.Join("level1", "encrypted"+constants.FileExtension),
		}

		helpers.AssertEqual(t, len(expectedFiles), len(eligibleFiles))
		helpers.AssertEqual(t, expectedFiles[0], eligibleFiles[0])
	})
}

// BenchmarkFinder_FindEligibleFiles benchmarks file finding performance
func BenchmarkFinder_FindEligibleFiles(b *testing.B) {
	// Create a temporary directory with many files
	tmpDir := CreateTempDir(b)
	defer CleanupTempDir(b, tmpDir)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			// Log error but don't fail benchmark
			_ = err
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create many test files
	testFiles := make(map[string][]byte)
	for i := range 100 {
		testFiles[fmt.Sprintf("file%d.txt", i)] = []byte("test content")
		if i%10 == 0 {
			testFiles[fmt.Sprintf("encrypted%d.txt%s", i, constants.FileExtension)] = []byte("encrypted")
		}
	}

	CreateTestFiles(b, tmpDir, testFiles)

	finder := files.NewFinder()

	for b.Loop() {
		_, err := finder.FindEligibleFiles(constants.ModeEncrypt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function for benchmarks that need testing.TB interface
func CreateTestFiles(tb testing.TB, dir string, files map[string][]byte) {
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			tb.Fatalf("Failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			tb.Fatalf("Failed to write file %s: %v", path, err)
		}
	}
}

func CreateTempDir(tb testing.TB) string {
	tmpDir, err := os.MkdirTemp("", "hexwarden-test-*")
	if err != nil {
		tb.Fatalf("Failed to create temp dir: %v", err)
	}
	return tmpDir
}

func CleanupTempDir(tb testing.TB, path string) {
	if err := os.RemoveAll(path); err != nil {
		tb.Logf("Warning: Failed to cleanup temp dir %s: %v", path, err)
	}
}

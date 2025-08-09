package compression

import (
	"fmt"
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/compression"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestNewCompressor(t *testing.T) {
	tests := []struct {
		name  string
		level constants.CompressionLevel
	}{
		{
			name:  "No compression",
			level: constants.LevelNoCompression,
		},
		{
			name:  "Best speed",
			level: constants.LevelBestSpeed,
		},
		{
			name:  "Default compression",
			level: constants.LevelDefaultCompression,
		},
		{
			name:  "Best compression",
			level: constants.LevelBestCompression,
		},
		{
			name:  "Invalid level (too low) - should use default",
			level: constants.CompressionLevel(-1),
		},
		{
			name:  "Invalid level (too high) - should use default",
			level: constants.CompressionLevel(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor, err := compression.NewCompressor(tt.level)
			helpers.AssertNoError(t, err)
			if compressor == nil {
				t.Error("Expected compressor to be non-nil")
			}
		})
	}
}

func TestNewDefaultCompressor(t *testing.T) {
	compressor, err := compression.NewDefaultCompressor()
	helpers.AssertNoError(t, err)
	if compressor == nil {
		t.Error("Expected default compressor to be non-nil")
	}
}

func TestCompressor_Compress(t *testing.T) {
	testData := helpers.NewTestData()
	compressor, err := compression.NewDefaultCompressor()
	helpers.AssertNoError(t, err)

	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "Empty data",
			input:    []byte{},
			expected: []byte{}, // Empty data should return as-is
		},
		{
			name:  "Small data",
			input: []byte("Hello, World!"),
		},
		{
			name:  "Test data",
			input: testData.TestData,
		},
		{
			name:  "Large data",
			input: testData.LargeData,
		},
		{
			name:  "Repetitive data (highly compressible)",
			input: createRepetitiveData(1000),
		},
		{
			name:  "Random data (less compressible)",
			input: testData.ValidKey32,
		},
		{
			name:  "Single byte",
			input: []byte{0x42},
		},
		{
			name:  "Binary data",
			input: createBinaryData(256),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := compressor.Compress(tt.input)
			helpers.AssertNoError(t, err)

			if len(tt.input) == 0 {
				// Empty data should return as-is
				helpers.AssertBytesEqual(t, tt.expected, compressed)
			} else {
				// Non-empty data should be compressed (different from input)
				if compressed == nil {
					t.Error("Expected compressed data to be non-nil")
				}
				// Compressed data should be different from input (unless very small)
				if len(tt.input) > 10 {
					helpers.AssertBytesNotEqual(t, tt.input, compressed)
				}
			}
		})
	}
}

func TestCompressor_Decompress(t *testing.T) {
	testData := helpers.NewTestData()
	compressor, err := compression.NewDefaultCompressor()
	helpers.AssertNoError(t, err)

	// First compress some data to get valid compressed data
	validCompressed, err := compressor.Compress(testData.TestData)
	helpers.AssertNoError(t, err)

	tests := []struct {
		name        string
		input       []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Empty data",
			input:       []byte{},
			expectError: false,
		},
		{
			name:        "Valid compressed data",
			input:       validCompressed,
			expectError: false,
		},
		{
			name:        "Invalid compressed data",
			input:       []byte("this is not compressed data"),
			expectError: true,
			expectedErr: constants.ErrDecompressionFailed,
		},
		{
			name:        "Corrupted compressed data",
			input:       corruptCompressedData(validCompressed),
			expectError: true,
			expectedErr: constants.ErrDecompressionFailed,
		},
		{
			name:        "Truncated compressed data",
			input:       validCompressed[:len(validCompressed)/2],
			expectError: true,
			expectedErr: constants.ErrDecompressionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decompressed, err := compressor.Decompress(tt.input)

			if tt.expectError {
				helpers.AssertError(t, err, tt.expectedErr)
			} else {
				helpers.AssertNoError(t, err)
				if len(tt.input) == 0 {
					// Empty data should return as-is
					helpers.AssertBytesEqual(t, tt.input, decompressed)
				} else {
					// Valid compressed data should decompress to original
					helpers.AssertBytesEqual(t, testData.TestData, decompressed)
				}
			}
		})
	}
}

func TestCompressor_CompressDecompressRoundTrip(t *testing.T) {
	testData := helpers.NewTestData()

	compressionLevels := []constants.CompressionLevel{
		constants.LevelNoCompression,
		constants.LevelBestSpeed,
		constants.LevelDefaultCompression,
		constants.LevelBestCompression,
	}

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Small text",
			data: []byte("Hello, World!"),
		},
		{
			name: "Test data",
			data: testData.TestData,
		},
		{
			name: "Large data",
			data: testData.LargeData[:10000], // Use smaller subset for faster tests
		},
		{
			name: "Repetitive data",
			data: createRepetitiveData(1000),
		},
		{
			name: "Binary data",
			data: createBinaryData(500),
		},
		{
			name: "Single byte",
			data: []byte{0xFF},
		},
	}

	for _, level := range compressionLevels {
		t.Run(fmt.Sprintf("Level_%d", level), func(t *testing.T) {
			compressor, err := compression.NewCompressor(level)
			helpers.AssertNoError(t, err)

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Compress
					compressed, err := compressor.Compress(tc.data)
					helpers.AssertNoError(t, err)

					// Decompress
					decompressed, err := compressor.Decompress(compressed)
					helpers.AssertNoError(t, err)

					// Verify round trip
					helpers.AssertBytesEqual(t, tc.data, decompressed)
				})
			}
		})
	}
}

func TestCompressor_CompressionEfficiency(t *testing.T) {
	compressor, err := compression.NewDefaultCompressor()
	helpers.AssertNoError(t, err)

	t.Run("Highly compressible data", func(t *testing.T) {
		// Create data with lots of repetition
		repetitiveData := createRepetitiveData(10000)
		compressed, err := compressor.Compress(repetitiveData)
		helpers.AssertNoError(t, err)

		// Compressed size should be significantly smaller
		compressionRatio := float64(len(compressed)) / float64(len(repetitiveData))
		if compressionRatio > 0.5 {
			t.Logf("Warning: Compression ratio %.2f is higher than expected for repetitive data", compressionRatio)
		}
		t.Logf("Repetitive data: %d bytes -> %d bytes (ratio: %.2f)",
			len(repetitiveData), len(compressed), compressionRatio)
	})

	t.Run("Less compressible data", func(t *testing.T) {
		// Use random data which is less compressible
		testData := helpers.NewTestData()
		randomData := testData.LargeData[:10000]
		compressed, err := compressor.Compress(randomData)
		helpers.AssertNoError(t, err)

		compressionRatio := float64(len(compressed)) / float64(len(randomData))
		t.Logf("Random data: %d bytes -> %d bytes (ratio: %.2f)",
			len(randomData), len(compressed), compressionRatio)
	})
}

func TestCompressor_DifferentLevelsComparison(t *testing.T) {
	testData := createRepetitiveData(5000) // Use compressible data

	levels := []constants.CompressionLevel{
		constants.LevelBestSpeed,
		constants.LevelDefaultCompression,
		constants.LevelBestCompression,
	}

	results := make(map[constants.CompressionLevel][]byte)

	for _, level := range levels {
		t.Run(fmt.Sprintf("Level_%d", level), func(t *testing.T) {
			compressor, err := compression.NewCompressor(level)
			helpers.AssertNoError(t, err)

			compressed, err := compressor.Compress(testData)
			helpers.AssertNoError(t, err)
			results[level] = compressed

			// Verify decompression works
			decompressed, err := compressor.Decompress(compressed)
			helpers.AssertNoError(t, err)
			helpers.AssertBytesEqual(t, testData, decompressed)

			t.Logf("Level %d: %d bytes -> %d bytes",
				level, len(testData), len(compressed))
		})
	}

	// Generally, best compression should produce smaller output than best speed
	// (though this isn't guaranteed for all data types)
	bestSpeedSize := len(results[constants.LevelBestSpeed])
	bestCompressionSize := len(results[constants.LevelBestCompression])

	t.Logf("Best speed size: %d, Best compression size: %d",
		bestSpeedSize, bestCompressionSize)
}

// BenchmarkCompressor_Compress benchmarks compression performance
func BenchmarkCompressor_Compress(b *testing.B) {
	compressor, err := compression.NewDefaultCompressor()
	if err != nil {
		b.Fatal(err)
	}

	testData := createRepetitiveData(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compressor.Compress(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCompressor_Decompress benchmarks decompression performance
func BenchmarkCompressor_Decompress(b *testing.B) {
	compressor, err := compression.NewDefaultCompressor()
	if err != nil {
		b.Fatal(err)
	}

	testData := createRepetitiveData(1000)
	compressed, err := compressor.Compress(testData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compressor.Decompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCompressor_RoundTrip benchmarks full compress-decompress cycle
func BenchmarkCompressor_RoundTrip(b *testing.B) {
	compressor, err := compression.NewDefaultCompressor()
	if err != nil {
		b.Fatal(err)
	}

	testData := createRepetitiveData(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compressed, err := compressor.Compress(testData)
		if err != nil {
			b.Fatal(err)
		}
		_, err = compressor.Decompress(compressed)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions

func createRepetitiveData(size int) []byte {
	pattern := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	data := make([]byte, size)

	for i := 0; i < size; i++ {
		data[i] = pattern[i%len(pattern)]
	}

	return data
}

func createBinaryData(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i % 256)
	}
	return data
}

func corruptCompressedData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	corrupted := make([]byte, len(data))
	copy(corrupted, data)
	// Corrupt the middle of the data
	if len(corrupted) > 2 {
		corrupted[len(corrupted)/2] ^= 0xFF
	}
	return corrupted
}

package utils

import (
	"fmt"
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/utils"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestNewPadder(t *testing.T) {
	tests := []struct {
		name        string
		blockSize   int
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid block size 16",
			blockSize:   16,
			expectError: false,
		},
		{
			name:        "Valid block size 8",
			blockSize:   8,
			expectError: false,
		},
		{
			name:        "Valid block size 32",
			blockSize:   32,
			expectError: false,
		},
		{
			name:        "Valid block size 1",
			blockSize:   1,
			expectError: false,
		},
		{
			name:        "Valid block size 255",
			blockSize:   255,
			expectError: false,
		},
		{
			name:        "Invalid block size 0",
			blockSize:   0,
			expectError: true,
			expectedErr: constants.ErrPaddingFailed,
		},
		{
			name:        "Invalid block size negative",
			blockSize:   -1,
			expectError: true,
			expectedErr: constants.ErrPaddingFailed,
		},
		{
			name:        "Invalid block size too large",
			blockSize:   256,
			expectError: true,
			expectedErr: constants.ErrPaddingFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padder, err := utils.NewPadder(tt.blockSize)

			if tt.expectError {
				helpers.AssertError(t, err, tt.expectedErr)
				if padder != nil {
					t.Error("Expected padder to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if padder == nil {
					t.Error("Expected padder to be non-nil when no error occurs")
				}
			}
		})
	}
}

func TestNewDefaultPadder(t *testing.T) {
	padder, err := utils.NewDefaultPadder()
	helpers.AssertNoError(t, err)
	if padder == nil {
		t.Error("Expected default padder to be non-nil")
	}
}

func TestPadder_Pad(t *testing.T) {
	padder, err := utils.NewPadder(16)
	helpers.AssertNoError(t, err)

	tests := []struct {
		name         string
		input        []byte
		expectError  bool
		expectedErr  error
		expectedSize int
	}{
		{
			name:         "Empty data",
			input:        []byte{},
			expectError:  false,
			expectedSize: 16, // Full block of padding
		},
		{
			name:         "Single byte",
			input:        []byte{0x01},
			expectError:  false,
			expectedSize: 16, // 1 byte + 15 bytes padding
		},
		{
			name:         "15 bytes (one less than block)",
			input:        make([]byte, 15),
			expectError:  false,
			expectedSize: 16, // 15 bytes + 1 byte padding
		},
		{
			name:         "16 bytes (exact block)",
			input:        make([]byte, 16),
			expectError:  false,
			expectedSize: 32, // 16 bytes + 16 bytes padding
		},
		{
			name:         "17 bytes (one more than block)",
			input:        make([]byte, 17),
			expectError:  false,
			expectedSize: 32, // 17 bytes + 15 bytes padding
		},
		{
			name:         "Large data",
			input:        make([]byte, 1000),
			expectError:  false,
			expectedSize: 1008, // 1000 bytes + 8 bytes padding (1000 % 16 = 8, so 16-8=8 padding)
		},
		{
			name:        "Nil data",
			input:       nil,
			expectError: true,
			expectedErr: constants.ErrPaddingFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			padded, err := padder.Pad(tt.input)

			if tt.expectError {
				helpers.AssertError(t, err, tt.expectedErr)
				if padded != nil {
					t.Error("Expected padded data to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if padded == nil {
					t.Error("Expected padded data to be non-nil when no error occurs")
				}

				// Check size
				helpers.AssertEqual(t, tt.expectedSize, len(padded))

				// Check that size is multiple of block size
				if len(padded)%16 != 0 {
					t.Errorf("Padded data size (%d) is not multiple of block size", len(padded))
				}

				// Check that original data is preserved
				if len(tt.input) > 0 {
					helpers.AssertBytesEqual(t, tt.input, padded[:len(tt.input)])
				}

				// Check padding bytes
				paddingLength := len(padded) - len(tt.input)
				expectedPaddingByte := byte(paddingLength)
				for i := len(tt.input); i < len(padded); i++ {
					if padded[i] != expectedPaddingByte {
						t.Errorf("Invalid padding byte at position %d: expected %d, got %d",
							i, expectedPaddingByte, padded[i])
					}
				}
			}
		})
	}
}

func TestPadder_Unpad(t *testing.T) {
	padder, err := utils.NewPadder(16)
	helpers.AssertNoError(t, err)

	// Create valid padded data for testing
	testData := []byte("Hello, World!")
	validPadded, err := padder.Pad(testData)
	helpers.AssertNoError(t, err)

	tests := []struct {
		name        string
		input       []byte
		expectError bool
		expectedErr error
		expected    []byte
	}{
		{
			name:        "Valid padded data",
			input:       validPadded,
			expectError: false,
			expected:    testData,
		},
		{
			name:        "Full block padding",
			input:       createPaddedData([]byte{}, 16),
			expectError: false,
			expected:    []byte{},
		},
		{
			name:        "Single byte with padding",
			input:       createPaddedData([]byte{0x42}, 16),
			expectError: false,
			expected:    []byte{0x42},
		},
		{
			name:        "Empty data",
			input:       []byte{},
			expectError: true,
			expectedErr: constants.ErrUnpaddingFailed,
		},
		{
			name:        "Data not multiple of block size",
			input:       []byte{0x01, 0x02, 0x03},
			expectError: true,
			expectedErr: constants.ErrUnpaddingFailed,
		},
		{
			name:        "Invalid padding - zero padding",
			input:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x00},
			expectError: true,
			expectedErr: constants.ErrUnpaddingFailed,
		},
		{
			name:        "Invalid padding - too large",
			input:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x11},
			expectError: true,
			expectedErr: constants.ErrUnpaddingFailed,
		},
		{
			name:        "Invalid padding - inconsistent bytes",
			input:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x03, 0x04},
			expectError: true,
			expectedErr: constants.ErrUnpaddingFailed,
		},
		{
			name:        "Padding larger than data",
			input:       []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expectError: true,
			expectedErr: constants.ErrUnpaddingFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unpadded, err := padder.Unpad(tt.input)

			if tt.expectError {
				helpers.AssertError(t, err, tt.expectedErr)
				if unpadded != nil {
					t.Error("Expected unpadded data to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if unpadded == nil && len(tt.expected) > 0 {
					t.Error("Expected unpadded data to be non-nil when no error occurs")
				}

				helpers.AssertBytesEqual(t, tt.expected, unpadded)
			}
		})
	}
}

func TestPadder_PadUnpadRoundTrip(t *testing.T) {
	blockSizes := []int{1, 8, 16, 32, 64}
	testData := [][]byte{
		{},
		{0x42},
		[]byte("Hello, World!"),
		[]byte("This is a longer test string for padding"),
		make([]byte, 100),
		make([]byte, 255),
		make([]byte, 1000),
	}

	for _, blockSize := range blockSizes {
		t.Run(fmt.Sprintf("BlockSize_%d", blockSize), func(t *testing.T) {
			padder, err := utils.NewPadder(blockSize)
			helpers.AssertNoError(t, err)

			for i, data := range testData {
				t.Run(fmt.Sprintf("Data_%d", i), func(t *testing.T) {
					// Pad
					padded, err := padder.Pad(data)
					helpers.AssertNoError(t, err)

					// Verify padding properties
					if len(padded)%blockSize != 0 {
						t.Errorf("Padded data size (%d) is not multiple of block size (%d)",
							len(padded), blockSize)
					}

					// Unpad
					unpadded, err := padder.Unpad(padded)
					helpers.AssertNoError(t, err)

					// Verify round trip
					helpers.AssertBytesEqual(t, data, unpadded)
				})
			}
		})
	}
}

func TestPadder_DifferentBlockSizes(t *testing.T) {
	testData := []byte("Test data for different block sizes")

	blockSizes := []int{1, 2, 4, 8, 16, 32, 64, 128, 255}
	for _, blockSize := range blockSizes {
		t.Run(fmt.Sprintf("BlockSize_%d", blockSize), func(t *testing.T) {
			padder, err := utils.NewPadder(blockSize)
			helpers.AssertNoError(t, err)

			padded, err := padder.Pad(testData)
			helpers.AssertNoError(t, err)

			// Check that padded size is correct
			expectedPadding := blockSize - (len(testData) % blockSize)
			expectedSize := len(testData) + expectedPadding
			helpers.AssertEqual(t, expectedSize, len(padded))

			// Check that it's a multiple of block size
			if len(padded)%blockSize != 0 {
				t.Errorf("Padded size (%d) is not multiple of block size (%d)",
					len(padded), blockSize)
			}

			// Verify round trip
			unpadded, err := padder.Unpad(padded)
			helpers.AssertNoError(t, err)
			helpers.AssertBytesEqual(t, testData, unpadded)
		})
	}
}

func TestPadder_EdgeCases(t *testing.T) {
	t.Run("Maximum valid block size", func(t *testing.T) {
		padder, err := utils.NewPadder(255)
		helpers.AssertNoError(t, err)

		data := []byte("test")
		padded, err := padder.Pad(data)
		helpers.AssertNoError(t, err)

		unpadded, err := padder.Unpad(padded)
		helpers.AssertNoError(t, err)
		helpers.AssertBytesEqual(t, data, unpadded)
	})

	t.Run("Minimum valid block size", func(t *testing.T) {
		padder, err := utils.NewPadder(1)
		helpers.AssertNoError(t, err)

		data := []byte("test")
		padded, err := padder.Pad(data)
		helpers.AssertNoError(t, err)

		// With block size 1, every byte needs padding
		expectedSize := len(data) + 1 // Always add 1 byte padding
		helpers.AssertEqual(t, expectedSize, len(padded))

		unpadded, err := padder.Unpad(padded)
		helpers.AssertNoError(t, err)
		helpers.AssertBytesEqual(t, data, unpadded)
	})
}

// BenchmarkPadder_Pad benchmarks padding performance
func BenchmarkPadder_Pad(b *testing.B) {
	padder, err := utils.NewDefaultPadder()
	if err != nil {
		b.Fatal(err)
	}

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := padder.Pad(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPadder_Unpad benchmarks unpadding performance
func BenchmarkPadder_Unpad(b *testing.B) {
	padder, err := utils.NewDefaultPadder()
	if err != nil {
		b.Fatal(err)
	}

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	padded, err := padder.Pad(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := padder.Unpad(padded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to create properly padded data for testing
func createPaddedData(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padded := make([]byte, len(data)+padding)
	copy(padded, data)

	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	return padded
}

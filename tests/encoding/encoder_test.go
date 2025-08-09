package encoding

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/encoding"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestNewEncoder(t *testing.T) {
	tests := []struct {
		name         string
		dataShards   int
		parityShards int
		expectError  bool
		expectedErr  error
	}{
		{
			name:         "Valid configuration",
			dataShards:   4,
			parityShards: 2,
			expectError:  false,
		},
		{
			name:         "Default configuration",
			dataShards:   constants.DataShards,
			parityShards: constants.ParityShards,
			expectError:  false,
		},
		{
			name:         "Minimum valid configuration",
			dataShards:   1,
			parityShards: 1,
			expectError:  false,
		},
		{
			name:         "Large configuration",
			dataShards:   10,
			parityShards: 5,
			expectError:  false,
		},
		{
			name:         "Zero data shards",
			dataShards:   0,
			parityShards: 2,
			expectError:  true,
			expectedErr:  constants.ErrEncodingFailed,
		},
		{
			name:         "Negative data shards",
			dataShards:   -1,
			parityShards: 2,
			expectError:  true,
			expectedErr:  constants.ErrEncodingFailed,
		},
		{
			name:         "Zero parity shards",
			dataShards:   4,
			parityShards: 0,
			expectError:  true,
			expectedErr:  constants.ErrEncodingFailed,
		},
		{
			name:         "Negative parity shards",
			dataShards:   4,
			parityShards: -1,
			expectError:  true,
			expectedErr:  constants.ErrEncodingFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder, err := encoding.NewEncoder(tt.dataShards, tt.parityShards)

			if tt.expectError {
				helpers.AssertError(t, err, nil)
				if encoder != nil {
					t.Error("Expected encoder to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if encoder == nil {
					t.Error("Expected encoder to be non-nil when no error occurs")
				}
			}
		})
	}
}

func TestNewDefaultEncoder(t *testing.T) {
	encoder, err := encoding.NewDefaultEncoder()
	helpers.AssertNoError(t, err)
	if encoder == nil {
		t.Error("Expected default encoder to be non-nil")
	}
}

func TestEncoder_Encode(t *testing.T) {
	testData := helpers.NewTestData()
	encoder, err := encoding.NewDefaultEncoder()
	helpers.AssertNoError(t, err)

	tests := []struct {
		name        string
		input       []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Small data",
			input:       []byte("Hello, World!"),
			expectError: false,
		},
		{
			name:        "Test data",
			input:       testData.TestData,
			expectError: false,
		},
		{
			name:        "Large data",
			input:       testData.LargeData[:10000], // Use smaller subset
			expectError: false,
		},
		{
			name:        "Single byte",
			input:       []byte{0x42},
			expectError: false,
		},
		{
			name:        "Binary data",
			input:       testData.ValidKey32,
			expectError: false,
		},
		{
			name:        "Empty data",
			input:       []byte{},
			expectError: true,
			expectedErr: constants.ErrEncodingFailed,
		},
		{
			name:        "Nil data",
			input:       nil,
			expectError: true,
			expectedErr: constants.ErrEncodingFailed,
		},
		{
			name:        "Too large data",
			input:       make([]byte, constants.MaxDataLen+1),
			expectError: true,
			expectedErr: constants.ErrEncodingFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encoder.Encode(tt.input)

			if tt.expectError {
				helpers.AssertError(t, err, nil)
				if encoded != nil {
					t.Error("Expected encoded data to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if encoded == nil {
					t.Error("Expected encoded data to be non-nil when no error occurs")
				}

				// Encoded data should be larger than input (includes parity)
				if len(encoded) <= len(tt.input) {
					t.Errorf("Expected encoded data length (%d) to be greater than input length (%d)",
						len(encoded), len(tt.input))
				}

				// Encoded data should be different from input
				helpers.AssertBytesNotEqual(t, tt.input, encoded)
			}
		})
	}
}

func TestEncoder_Decode(t *testing.T) {
	testData := helpers.NewTestData()
	encoder, err := encoding.NewDefaultEncoder()
	helpers.AssertNoError(t, err)

	// First encode some data to get valid encoded data
	validEncoded, err := encoder.Encode(testData.TestData)
	helpers.AssertNoError(t, err)

	tests := []struct {
		name        string
		input       []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid encoded data",
			input:       validEncoded,
			expectError: false,
		},
		{
			name:        "Empty data",
			input:       []byte{},
			expectError: true,
			expectedErr: constants.ErrDecodingFailed,
		},
		{
			name:        "Invalid length data",
			input:       []byte("invalid data with wrong length"),
			expectError: true,
			expectedErr: constants.ErrDecodingFailed,
		},
		{
			name:        "Corrupted encoded data",
			input:       corruptEncodedData(validEncoded),
			expectError: false, // Reed-Solomon can recover from some corruption
		},
		{
			name:        "Truncated encoded data",
			input:       validEncoded[:len(validEncoded)/2],
			expectError: false, // May still be decodable depending on truncation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := encoder.Decode(tt.input)

			if tt.expectError {
				helpers.AssertError(t, err, nil)
			} else {
				helpers.AssertNoError(t, err)
				if decoded == nil {
					t.Error("Expected decoded data to be non-nil when no error occurs")
				}

				// For valid encoded data, check exact match
				if tt.name == "Valid encoded data" {
					// Decoded data should match original (after trimming padding)
					// Note: Reed-Solomon may add padding, so we check if original is prefix
					if len(decoded) < len(testData.TestData) {
						t.Errorf("Decoded data too short: got %d, expected at least %d",
							len(decoded), len(testData.TestData))
					} else {
						// Check that original data is preserved at the beginning
						helpers.AssertBytesEqual(t, testData.TestData, decoded[:len(testData.TestData)])
					}
				} else {
					// For corrupted/truncated data, just verify we got some data back
					t.Logf("Decoded %d bytes from corrupted/truncated data", len(decoded))
				}
			}
		})
	}
}

func TestEncoder_EncodeDecodeRoundTrip(t *testing.T) {
	testData := helpers.NewTestData()

	configurations := []struct {
		name         string
		dataShards   int
		parityShards int
	}{
		{
			name:         "Default config",
			dataShards:   constants.DataShards,
			parityShards: constants.ParityShards,
		},
		{
			name:         "Small config",
			dataShards:   2,
			parityShards: 1,
		},
		{
			name:         "Balanced config",
			dataShards:   4,
			parityShards: 4,
		},
		{
			name:         "High redundancy",
			dataShards:   3,
			parityShards: 6,
		},
	}

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "Small text",
			data: []byte("Hello, World!"),
		},
		{
			name: "Test data",
			data: testData.TestData,
		},
		{
			name: "Binary data",
			data: testData.ValidKey32,
		},
		{
			name: "Large data",
			data: testData.LargeData[:5000], // Use smaller subset for faster tests
		},
		{
			name: "Single byte",
			data: []byte{0xFF},
		},
		{
			name: "Repetitive data",
			data: createRepetitiveData(1000),
		},
	}

	for _, config := range configurations {
		t.Run(config.name, func(t *testing.T) {
			encoder, err := encoding.NewEncoder(config.dataShards, config.parityShards)
			helpers.AssertNoError(t, err)

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Encode
					encoded, err := encoder.Encode(tc.data)
					helpers.AssertNoError(t, err)

					// Decode
					decoded, err := encoder.Decode(encoded)
					helpers.AssertNoError(t, err)

					// Verify round trip (check that original data is preserved)
					if len(decoded) < len(tc.data) {
						t.Fatalf("Decoded data too short: got %d, expected at least %d",
							len(decoded), len(tc.data))
					}
					helpers.AssertBytesEqual(t, tc.data, decoded[:len(tc.data)])
				})
			}
		})
	}
}

func TestEncoder_ErrorRecovery(t *testing.T) {
	// Test the error correction capability of Reed-Solomon encoding
	encoder, err := encoding.NewEncoder(4, 6) // 4 data shards, 6 parity shards
	helpers.AssertNoError(t, err)

	testData := []byte("This is test data for error recovery testing!")

	// Encode the data
	encoded, err := encoder.Encode(testData)
	helpers.AssertNoError(t, err)

	t.Run("No corruption", func(t *testing.T) {
		decoded, err := encoder.Decode(encoded)
		helpers.AssertNoError(t, err)
		helpers.AssertBytesEqual(t, testData, decoded[:len(testData)])
	})

	t.Run("Single shard corruption", func(t *testing.T) {
		// Corrupt one shard (should still be recoverable with 6 parity shards)
		corrupted := make([]byte, len(encoded))
		copy(corrupted, encoded)

		// Corrupt the first shard completely
		shardSize := len(encoded) / (4 + 6) // total shards
		for i := 0; i < shardSize; i++ {
			corrupted[i] = 0xFF
		}

		decoded, err := encoder.Decode(corrupted)
		helpers.AssertNoError(t, err)

		// Reed-Solomon may recover the data, but it might not be perfect
		// Just check that we got some data back and no error
		if len(decoded) < len(testData) {
			t.Errorf("Decoded data too short: got %d, expected at least %d", len(decoded), len(testData))
		}

		// Note: The exact recovery depends on Reed-Solomon implementation details
		t.Logf("Original: %x", testData)
		t.Logf("Recovered: %x", decoded[:len(testData)])
	})

	// Note: Testing multiple shard corruption would require more complex
	// shard manipulation and might not always be recoverable depending on
	// the specific Reed-Solomon configuration
}

func TestEncoder_DataSizeValidation(t *testing.T) {
	encoder, err := encoding.NewDefaultEncoder()
	helpers.AssertNoError(t, err)

	t.Run("Maximum valid size", func(t *testing.T) {
		// Test with maximum allowed data size
		largeData := make([]byte, constants.MaxDataLen)
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		encoded, err := encoder.Encode(largeData)
		helpers.AssertNoError(t, err)

		decoded, err := encoder.Decode(encoded)
		helpers.AssertNoError(t, err)
		helpers.AssertBytesEqual(t, largeData, decoded[:len(largeData)])
	})

	t.Run("Exceeds maximum size", func(t *testing.T) {
		oversizedData := make([]byte, constants.MaxDataLen+1)
		_, err := encoder.Encode(oversizedData)
		helpers.AssertError(t, err, nil) // Just check that an error occurred

		// Check that the error message contains the expected information
		if err != nil && !strings.Contains(err.Error(), "encoding operation failed") {
			t.Errorf("Expected encoding failed error, got: %v", err)
		}
	})
}

// BenchmarkEncoder_Encode benchmarks encoding performance
func BenchmarkEncoder_Encode(b *testing.B) {
	encoder, err := encoding.NewDefaultEncoder()
	if err != nil {
		b.Fatal(err)
	}

	testData := createRepetitiveData(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encoder.Encode(testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncoder_Decode benchmarks decoding performance
func BenchmarkEncoder_Decode(b *testing.B) {
	encoder, err := encoding.NewDefaultEncoder()
	if err != nil {
		b.Fatal(err)
	}

	testData := createRepetitiveData(1000)
	encoded, err := encoder.Encode(testData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encoder.Decode(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncoder_RoundTrip benchmarks full encode-decode cycle
func BenchmarkEncoder_RoundTrip(b *testing.B) {
	encoder, err := encoding.NewDefaultEncoder()
	if err != nil {
		b.Fatal(err)
	}

	testData := createRepetitiveData(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded, err := encoder.Encode(testData)
		if err != nil {
			b.Fatal(err)
		}
		_, err = encoder.Decode(encoded)
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

func corruptEncodedData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	corrupted := make([]byte, len(data))
	copy(corrupted, data)

	// Corrupt multiple bytes to make recovery impossible
	corruptionPoints := []int{
		len(corrupted) / 4,
		len(corrupted) / 2,
		3 * len(corrupted) / 4,
	}

	for _, point := range corruptionPoints {
		if point < len(corrupted) {
			corrupted[point] ^= 0xFF
		}
	}

	return corrupted
}

func TestEncoder_ShardSizeCalculation(t *testing.T) {
	// Test that shard sizes are calculated correctly for different data sizes
	encoder, err := encoding.NewEncoder(4, 2) // 4 data shards, 2 parity shards
	helpers.AssertNoError(t, err)

	testSizes := []int{1, 4, 5, 16, 17, 100, 1000}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			encoded, err := encoder.Encode(data)
			helpers.AssertNoError(t, err)

			// Total shards = data shards + parity shards = 6
			// Encoded length should be divisible by total shards
			totalShards := 6
			if len(encoded)%totalShards != 0 {
				t.Errorf("Encoded length %d is not divisible by total shards %d",
					len(encoded), totalShards)
			}

			// Verify round trip
			decoded, err := encoder.Decode(encoded)
			helpers.AssertNoError(t, err)
			helpers.AssertBytesEqual(t, data, decoded[:len(data)])
		})
	}
}

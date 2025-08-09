package crypto

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/crypto"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestNewHeader(t *testing.T) {
	testData := helpers.NewTestData()

	tests := []struct {
		name         string
		salt         []byte
		originalSize uint64
		key          []byte
		expectError  bool
	}{
		{
			name:         "Valid inputs",
			salt:         testData.ValidSalt,
			originalSize: 1024,
			key:          testData.ValidKey32,
			expectError:  false,
		},
		{
			name:         "Zero original size",
			salt:         testData.ValidSalt,
			originalSize: 0,
			key:          testData.ValidKey32,
			expectError:  false,
		},
		{
			name:         "Large original size",
			salt:         testData.ValidSalt,
			originalSize: 1<<63 - 1, // Max uint64
			key:          testData.ValidKey32,
			expectError:  false,
		},
		{
			name:         "Invalid salt size",
			salt:         []byte("short"),
			originalSize: 1024,
			key:          testData.ValidKey32,
			expectError:  true,
		},
		{
			name:         "Weak salt",
			salt:         testData.WeakSalt,
			originalSize: 1024,
			key:          testData.ValidKey32,
			expectError:  true,
		},
		{
			name:         "Nil salt",
			salt:         nil,
			originalSize: 1024,
			key:          testData.ValidKey32,
			expectError:  true,
		},
		{
			name:         "Empty key",
			salt:         testData.ValidSalt,
			originalSize: 1024,
			key:          []byte{},
			expectError:  true,
		},
		{
			name:         "Nil key",
			salt:         testData.ValidSalt,
			originalSize: 1024,
			key:          nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, err := crypto.NewHeader(tt.salt, tt.originalSize, tt.key)

			if tt.expectError {
				helpers.AssertError(t, err, nil)
				if header != nil {
					t.Error("Expected header to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if header == nil {
					t.Error("Expected header to be non-nil when no error occurs")
				}

				// Verify header properties
				helpers.AssertBytesEqual(t, tt.salt, header.Salt())
				helpers.AssertEqual(t, tt.originalSize, header.OriginalSize())

				// Nonce should be non-nil and correct size
				nonce := header.Nonce()
				if len(nonce) != constants.NonceSizeBytes {
					t.Errorf("Expected nonce size %d, got %d", constants.NonceSizeBytes, len(nonce))
				}

				// Verify key works with header
				err = header.VerifyKey(tt.key)
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestHeader_VerifyKey(t *testing.T) {
	testData := helpers.NewTestData()
	originalSize := uint64(1024)

	header, err := crypto.NewHeader(testData.ValidSalt, originalSize, testData.ValidKey32)
	helpers.AssertNoError(t, err)

	tests := []struct {
		name        string
		key         []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Correct key",
			key:         testData.ValidKey32,
			expectError: false,
		},
		{
			name:        "Wrong key",
			key:         testData.ValidKey24, // Different key
			expectError: true,
			expectedErr: constants.ErrAuthFailure,
		},
		{
			name:        "Nil key",
			key:         nil,
			expectError: true,
		},
		{
			name:        "Empty key",
			key:         []byte{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := header.VerifyKey(tt.key)

			if tt.expectError {
				helpers.AssertError(t, err, nil)
				if tt.expectedErr != nil {
					helpers.AssertError(t, err, tt.expectedErr)
				}
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestHeader_WriteTo_ReadHeader_RoundTrip(t *testing.T) {
	testData := helpers.NewTestData()

	testCases := []struct {
		name         string
		salt         []byte
		originalSize uint64
		key          []byte
	}{
		{
			name:         "Standard case",
			salt:         testData.ValidSalt,
			originalSize: 1024,
			key:          testData.ValidKey32,
		},
		{
			name:         "Zero size",
			salt:         testData.ValidSalt,
			originalSize: 0,
			key:          testData.ValidKey32,
		},
		{
			name:         "Large size",
			salt:         testData.ValidSalt,
			originalSize: 1<<32 - 1,
			key:          testData.ValidKey32,
		},
		{
			name:         "Different key size",
			salt:         testData.ValidSalt,
			originalSize: 2048,
			key:          testData.ValidKey24,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Create header
			originalHeader, err := crypto.NewHeader(tt.salt, tt.originalSize, tt.key)
			helpers.AssertNoError(t, err)

			// Write to buffer
			var buf bytes.Buffer
			n, err := originalHeader.WriteTo(&buf)
			helpers.AssertNoError(t, err)
			helpers.AssertEqual(t, int64(constants.TotalHeaderSize), n)
			helpers.AssertEqual(t, constants.TotalHeaderSize, buf.Len())

			// Read back from buffer
			readHeader, err := crypto.ReadHeader(&buf)
			helpers.AssertNoError(t, err)

			// Verify all fields match
			helpers.AssertBytesEqual(t, originalHeader.Salt(), readHeader.Salt())
			helpers.AssertEqual(t, originalHeader.OriginalSize(), readHeader.OriginalSize())
			helpers.AssertBytesEqual(t, originalHeader.Nonce(), readHeader.Nonce())

			// Verify key verification works on read header
			err = readHeader.VerifyKey(tt.key)
			helpers.AssertNoError(t, err)

			// Verify wrong key fails on read header
			wrongKey := make([]byte, len(tt.key))
			copy(wrongKey, tt.key)
			wrongKey[0] ^= 0xFF
			err = readHeader.VerifyKey(wrongKey)
			helpers.AssertError(t, err, constants.ErrAuthFailure)
		})
	}
}

func TestHeader_Write(t *testing.T) {
	testData := helpers.NewTestData()
	header, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
	helpers.AssertNoError(t, err)

	var buf bytes.Buffer
	err = header.Write(&buf)
	helpers.AssertNoError(t, err)
	helpers.AssertEqual(t, constants.TotalHeaderSize, buf.Len())
}

func TestReadHeader_InvalidData(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedErr error
	}{
		{
			name:        "Empty data",
			data:        []byte{},
			expectedErr: constants.ErrIncompleteRead,
		},
		{
			name:        "Too short data",
			data:        make([]byte, constants.TotalHeaderSize-1),
			expectedErr: constants.ErrIncompleteRead,
		},
		{
			name:        "Invalid magic bytes",
			data:        createInvalidMagicHeader(),
			expectedErr: constants.ErrInvalidMagic,
		},
		{
			name:        "Corrupted checksum",
			data:        createCorruptedChecksumHeader(),
			expectedErr: constants.ErrChecksumMismatch,
		},
		{
			name:        "Tampered data",
			data:        createTamperedHeader(),
			expectedErr: constants.ErrChecksumMismatch, // Checksum fails before tampering detection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewReader(tt.data)
			header, err := crypto.ReadHeader(buf)

			helpers.AssertError(t, err, nil)
			if header != nil {
				t.Error("Expected header to be nil when error occurs")
			}

			// For incomplete read errors, check the specific error type
			if tt.expectedErr == constants.ErrIncompleteRead {
				if !strings.Contains(err.Error(), "incomplete header read") {
					t.Errorf("Expected incomplete read error, got: %v", err)
				}
			} else if tt.expectedErr != nil {
				helpers.AssertError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestHeader_TamperDetection(t *testing.T) {
	testData := helpers.NewTestData()
	header, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
	helpers.AssertNoError(t, err)

	// Write header to buffer
	var buf bytes.Buffer
	err = header.Write(&buf)
	helpers.AssertNoError(t, err)

	originalData := buf.Bytes()

	// Test tampering at different positions
	tamperPositions := []struct {
		name     string
		position int
		expected error
	}{
		{"Magic bytes", 0, constants.ErrInvalidMagic},
		{"Salt", 10, constants.ErrChecksumMismatch},           // Checksum fails first
		{"Original size", 40, constants.ErrChecksumMismatch},  // Checksum fails first
		{"Nonce", 50, constants.ErrChecksumMismatch},          // Checksum fails first
		{"Integrity hash", 70, constants.ErrChecksumMismatch}, // Checksum fails first
		{"Auth tag", 100, constants.ErrChecksumMismatch},      // Checksum fails first
		{"Checksum", constants.TotalHeaderSize - 1, constants.ErrChecksumMismatch},
	}

	for _, tp := range tamperPositions {
		t.Run("Tamper "+tp.name, func(t *testing.T) {
			// Create tampered data
			tamperedData := make([]byte, len(originalData))
			copy(tamperedData, originalData)
			tamperedData[tp.position] ^= 0xFF

			// Try to read tampered header
			buf := bytes.NewReader(tamperedData)
			readHeader, err := crypto.ReadHeader(buf)

			// All tampering should result in checksum mismatch or invalid magic
			helpers.AssertError(t, err, tp.expected)
			if readHeader != nil {
				t.Error("Expected header to be nil when read error occurs")
			}
		})
	}
}

func TestHeader_NonceUniqueness(t *testing.T) {
	testData := helpers.NewTestData()

	// Create multiple headers with same inputs
	headers := make([]*crypto.Header, 10)
	for i := 0; i < 10; i++ {
		header, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
		helpers.AssertNoError(t, err)
		headers[i] = header
	}

	// All nonces should be different
	for i := 0; i < len(headers); i++ {
		for j := i + 1; j < len(headers); j++ {
			helpers.AssertBytesNotEqual(t, headers[i].Nonce(), headers[j].Nonce())
		}
	}
}

func TestHeader_DifferentKeysProduceDifferentHeaders(t *testing.T) {
	testData := helpers.NewTestData()

	header1, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
	helpers.AssertNoError(t, err)

	header2, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey24)
	helpers.AssertNoError(t, err)

	// Headers should have different auth tags (due to different keys)
	var buf1, buf2 bytes.Buffer
	err = header1.Write(&buf1)
	helpers.AssertNoError(t, err)
	err = header2.Write(&buf2)
	helpers.AssertNoError(t, err)

	helpers.AssertBytesNotEqual(t, buf1.Bytes(), buf2.Bytes())

	// Each header should only verify with its own key
	err = header1.VerifyKey(testData.ValidKey32)
	helpers.AssertNoError(t, err)
	err = header1.VerifyKey(testData.ValidKey24)
	helpers.AssertError(t, err, constants.ErrAuthFailure)

	err = header2.VerifyKey(testData.ValidKey24)
	helpers.AssertNoError(t, err)
	err = header2.VerifyKey(testData.ValidKey32)
	helpers.AssertError(t, err, constants.ErrAuthFailure)
}

// BenchmarkNewHeader benchmarks header creation
func BenchmarkNewHeader(b *testing.B) {
	testData := helpers.NewTestData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHeader_Write benchmarks header writing
func BenchmarkHeader_Write(b *testing.B) {
	testData := helpers.NewTestData()
	header, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := header.Write(&buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkReadHeader benchmarks header reading
func BenchmarkReadHeader(b *testing.B) {
	testData := helpers.NewTestData()
	header, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
	if err != nil {
		b.Fatal(err)
	}

	var buf bytes.Buffer
	err = header.Write(&buf)
	if err != nil {
		b.Fatal(err)
	}
	headerData := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewReader(headerData)
		_, err := crypto.ReadHeader(buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHeader_VerifyKey benchmarks key verification
func BenchmarkHeader_VerifyKey(b *testing.B) {
	testData := helpers.NewTestData()
	header, err := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := header.VerifyKey(testData.ValidKey32)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions for creating invalid test data

func createInvalidMagicHeader() []byte {
	data := make([]byte, constants.TotalHeaderSize)
	copy(data, "XXXX") // Invalid magic bytes
	return data
}

func createCorruptedChecksumHeader() []byte {
	testData := helpers.NewTestData()
	header, _ := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)

	var buf bytes.Buffer
	header.Write(&buf)
	data := buf.Bytes()

	// Corrupt the checksum (last 4 bytes)
	data[len(data)-1] ^= 0xFF
	return data
}

func createTamperedHeader() []byte {
	testData := helpers.NewTestData()
	header, _ := crypto.NewHeader(testData.ValidSalt, 1024, testData.ValidKey32)

	var buf bytes.Buffer
	header.Write(&buf)
	data := buf.Bytes()

	// Tamper with salt data (position 10)
	data[10] ^= 0xFF
	return data
}

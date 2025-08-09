package crypto

import (
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/crypto"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestNewAESCipher(t *testing.T) {
	testData := helpers.NewTestData()

	tests := []struct {
		name        string
		key         []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid 32-byte key (AES-256)",
			key:         testData.ValidKey32,
			expectError: false,
		},
		{
			name:        "Valid 24-byte key (AES-192)",
			key:         testData.ValidKey24,
			expectError: false,
		},
		{
			name:        "Valid 16-byte key (AES-128)",
			key:         testData.ValidKey16,
			expectError: false,
		},
		{
			name:        "Invalid key size",
			key:         testData.InvalidKey,
			expectError: true,
			expectedErr: constants.ErrInvalidKeySize,
		},
		{
			name:        "Nil key",
			key:         nil,
			expectError: true,
			expectedErr: constants.ErrInvalidKeySize,
		},
		{
			name:        "Empty key",
			key:         []byte{},
			expectError: true,
			expectedErr: constants.ErrInvalidKeySize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := crypto.NewAESCipher(tt.key)

			if tt.expectError {
				helpers.AssertError(t, err, tt.expectedErr)
				if cipher != nil {
					t.Error("Expected cipher to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if cipher == nil {
					t.Error("Expected cipher to be non-nil when no error occurs")
				}
			}
		})
	}
}

func TestAESCipher_Encrypt(t *testing.T) {
	testData := helpers.NewTestData()
	cipher, err := crypto.NewAESCipher(testData.ValidKey32)
	helpers.AssertNoError(t, err)

	tests := []struct {
		name        string
		plaintext   []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid plaintext",
			plaintext:   testData.TestData,
			expectError: false,
		},
		{
			name:        "Large plaintext",
			plaintext:   testData.LargeData,
			expectError: false,
		},
		{
			name:        "Single byte",
			plaintext:   []byte{0x42},
			expectError: false,
		},
		{
			name:        "Empty plaintext",
			plaintext:   []byte{},
			expectError: true,
			expectedErr: constants.ErrEmptyPlaintext,
		},
		{
			name:        "Nil plaintext",
			plaintext:   nil,
			expectError: true,
			expectedErr: constants.ErrEmptyPlaintext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := cipher.Encrypt(tt.plaintext)

			if tt.expectError {
				helpers.AssertError(t, err, tt.expectedErr)
				if ciphertext != nil {
					t.Error("Expected ciphertext to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if ciphertext == nil {
					t.Error("Expected ciphertext to be non-nil when no error occurs")
				}

				// Ciphertext should be longer than plaintext (includes nonce + auth tag)
				if len(ciphertext) <= len(tt.plaintext) {
					t.Errorf("Expected ciphertext length (%d) to be greater than plaintext length (%d)",
						len(ciphertext), len(tt.plaintext))
				}

				// Ciphertext should not equal plaintext
				helpers.AssertBytesNotEqual(t, tt.plaintext, ciphertext)
			}
		})
	}
}

func TestAESCipher_Decrypt(t *testing.T) {
	testData := helpers.NewTestData()
	cipher, err := crypto.NewAESCipher(testData.ValidKey32)
	helpers.AssertNoError(t, err)

	// First encrypt some data to get valid ciphertext
	validCiphertext, err := cipher.Encrypt(testData.TestData)
	helpers.AssertNoError(t, err)

	tests := []struct {
		name        string
		ciphertext  []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid ciphertext",
			ciphertext:  validCiphertext,
			expectError: false,
		},
		{
			name:        "Empty ciphertext",
			ciphertext:  []byte{},
			expectError: true,
			expectedErr: constants.ErrEmptyCiphertext,
		},
		{
			name:        "Nil ciphertext",
			ciphertext:  nil,
			expectError: true,
			expectedErr: constants.ErrEmptyCiphertext,
		},
		{
			name:        "Too short ciphertext",
			ciphertext:  []byte{0x01, 0x02},
			expectError: true,
			expectedErr: constants.ErrDecryptionFailed,
		},
		{
			name:        "Invalid ciphertext",
			ciphertext:  []byte("this is not valid ciphertext data"),
			expectError: true,
			expectedErr: constants.ErrDecryptionFailed,
		},
		{
			name:        "Corrupted ciphertext",
			ciphertext:  corruptData(validCiphertext),
			expectError: true,
			expectedErr: constants.ErrDecryptionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plaintext, err := cipher.Decrypt(tt.ciphertext)

			if tt.expectError {
				helpers.AssertError(t, err, tt.expectedErr)
				if plaintext != nil {
					t.Error("Expected plaintext to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if plaintext == nil {
					t.Error("Expected plaintext to be non-nil when no error occurs")
				}

				// Decrypted plaintext should match original
				helpers.AssertBytesEqual(t, testData.TestData, plaintext)
			}
		})
	}
}

func TestAESCipher_EncryptDecryptRoundTrip(t *testing.T) {
	testData := helpers.NewTestData()

	testCases := []struct {
		name string
		key  []byte
		data []byte
	}{
		{
			name: "AES-256 with test data",
			key:  testData.ValidKey32,
			data: testData.TestData,
		},
		{
			name: "AES-192 with test data",
			key:  testData.ValidKey24,
			data: testData.TestData,
		},
		{
			name: "AES-128 with test data",
			key:  testData.ValidKey16,
			data: testData.TestData,
		},
		{
			name: "AES-256 with large data",
			key:  testData.ValidKey32,
			data: testData.LargeData,
		},
		{
			name: "AES-256 with single byte",
			key:  testData.ValidKey32,
			data: []byte{0xFF},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := crypto.NewAESCipher(tt.key)
			helpers.AssertNoError(t, err)

			// Encrypt
			ciphertext, err := cipher.Encrypt(tt.data)
			helpers.AssertNoError(t, err)

			// Decrypt
			decrypted, err := cipher.Decrypt(ciphertext)
			helpers.AssertNoError(t, err)

			// Verify round trip
			helpers.AssertBytesEqual(t, tt.data, decrypted)
		})
	}
}

func TestAESCipher_NonceUniqueness(t *testing.T) {
	testData := helpers.NewTestData()
	cipher, err := crypto.NewAESCipher(testData.ValidKey32)
	helpers.AssertNoError(t, err)

	// Encrypt the same data multiple times
	ciphertexts := make([][]byte, 10)
	for i := range 10 {
		ciphertext, err := cipher.Encrypt(testData.TestData)
		helpers.AssertNoError(t, err)
		ciphertexts[i] = ciphertext
	}

	// All ciphertexts should be different (due to random nonces)
	for i := range ciphertexts {
		for j := i + 1; j < len(ciphertexts); j++ {
			helpers.AssertBytesNotEqual(t, ciphertexts[i], ciphertexts[j])
		}
	}

	// But all should decrypt to the same plaintext
	for i, ciphertext := range ciphertexts {
		decrypted, err := cipher.Decrypt(ciphertext)
		helpers.AssertNoError(t, err)
		helpers.AssertBytesEqual(t, testData.TestData, decrypted)
		t.Logf("Ciphertext %d decrypted successfully", i)
	}
}

func TestAESCipher_DifferentKeysProduceDifferentResults(t *testing.T) {
	testData := helpers.NewTestData()

	cipher1, err := crypto.NewAESCipher(testData.ValidKey32)
	helpers.AssertNoError(t, err)

	// Create a different key
	differentKey := make([]byte, 32)
	copy(differentKey, testData.ValidKey32)
	differentKey[0] ^= 0xFF // Flip bits in first byte

	cipher2, err := crypto.NewAESCipher(differentKey)
	helpers.AssertNoError(t, err)

	// Encrypt with both ciphers
	ciphertext1, err := cipher1.Encrypt(testData.TestData)
	helpers.AssertNoError(t, err)

	ciphertext2, err := cipher2.Encrypt(testData.TestData)
	helpers.AssertNoError(t, err)

	// Results should be different
	helpers.AssertBytesNotEqual(t, ciphertext1, ciphertext2)

	// Each cipher should only be able to decrypt its own ciphertext
	decrypted1, err := cipher1.Decrypt(ciphertext1)
	helpers.AssertNoError(t, err)
	helpers.AssertBytesEqual(t, testData.TestData, decrypted1)

	decrypted2, err := cipher2.Decrypt(ciphertext2)
	helpers.AssertNoError(t, err)
	helpers.AssertBytesEqual(t, testData.TestData, decrypted2)

	// Cross-decryption should fail
	_, err = cipher1.Decrypt(ciphertext2)
	helpers.AssertError(t, err, constants.ErrDecryptionFailed)

	_, err = cipher2.Decrypt(ciphertext1)
	helpers.AssertError(t, err, constants.ErrDecryptionFailed)
}

// BenchmarkAESCipher_Encrypt benchmarks encryption performance
func BenchmarkAESCipher_Encrypt(b *testing.B) {
	testData := helpers.NewTestData()
	cipher, err := crypto.NewAESCipher(testData.ValidKey32)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, err := cipher.Encrypt(testData.TestData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAESCipher_Decrypt benchmarks decryption performance
func BenchmarkAESCipher_Decrypt(b *testing.B) {
	testData := helpers.NewTestData()
	cipher, err := crypto.NewAESCipher(testData.ValidKey32)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-encrypt data for benchmarking
	ciphertext, err := cipher.Encrypt(testData.TestData)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, err := cipher.Decrypt(ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// corruptData corrupts a byte slice by flipping bits in the last byte
func corruptData(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	corrupted := make([]byte, len(data))
	copy(corrupted, data)
	corrupted[len(corrupted)-1] ^= 0xFF
	return corrupted
}

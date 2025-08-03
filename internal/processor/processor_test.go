package processor

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		keyLen  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid key length",
			keyLen:  32,
			wantErr: false,
		},
		{
			name:    "valid key length - longer than minimum",
			keyLen:  64,
			wantErr: false,
		},
		{
			name:    "key too short",
			keyLen:  16,
			wantErr: true,
			errMsg:  "encryption key must be at least 32 bytes long",
		},
		{
			name:    "key exactly minimum length",
			keyLen:  32,
			wantErr: false,
		},
		{
			name:    "key one byte short",
			keyLen:  31,
			wantErr: true,
			errMsg:  "encryption key must be at least 32 bytes long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLen)
			_, err := rand.Read(key)
			if err != nil {
				t.Fatalf("failed to generate random key: %v", err)
			}

			processor, err := New(key)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("New() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error = %v", err)
				return
			}

			if processor == nil {
				t.Error("New() returned nil processor")
				return
			}

			// Verify all components are initialized
			if processor.cipher == nil {
				t.Error("cipher not initialized")
			}
			if processor.encoder == nil {
				t.Error("encoder not initialized")
			}
			if processor.compression == nil {
				t.Error("compression not initialized")
			}
			if processor.padder == nil {
				t.Error("padder not initialized")
			}
		})
	}
}

func TestProcessor_EncryptDecrypt(t *testing.T) {
	// Generate a valid key
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	processor, err := New(key)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "small data",
			data: []byte("hello world"),
		},
		{
			name: "medium data",
			data: []byte("This is a longer message that should test the compression and padding functionality more thoroughly."),
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC},
		},
		{
			name: "large data",
			data: bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyz"), 100),
		},
		{
			name: "highly compressible data",
			data: bytes.Repeat([]byte("A"), 1000),
		},
		{
			name: "random data (low compression)",
			data: func() []byte {
				data := make([]byte, 256)
				rand.Read(data)
				return data
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := processor.Encrypt(tt.data)
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}

			if encrypted == nil {
				t.Error("Encrypt() returned nil")
				return
			}

			// Encrypted data should be different from original (unless empty)
			if len(tt.data) > 0 && bytes.Equal(tt.data, encrypted) {
				t.Error("Encrypt() returned data identical to input")
			}

			// Test decryption
			decrypted, err := processor.Decrypt(encrypted)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			// Decrypted data should match original
			if !bytes.Equal(tt.data, decrypted) {
				t.Errorf("Decrypt() result doesn't match original data")
				t.Errorf("Original:  %v", tt.data)
				t.Errorf("Decrypted: %v", decrypted)
			}
		})
	}
}

func TestProcessor_EncryptDecrypt_MultipleRounds(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	processor, err := New(key)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	data := []byte("test data for multiple rounds")

	// Test multiple encrypt/decrypt cycles
	for i := range 5 {
		encrypted, err := processor.Encrypt(data)
		if err != nil {
			t.Errorf("Round %d: Encrypt() error = %v", i, err)
			return
		}

		decrypted, err := processor.Decrypt(encrypted)
		if err != nil {
			t.Errorf("Round %d: Decrypt() error = %v", i, err)
			return
		}

		if !bytes.Equal(data, decrypted) {
			t.Errorf("Round %d: data mismatch", i)
			return
		}
	}
}

func TestProcessor_EncryptDecrypt_DifferentProcessors(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	processor1, err := New(key)
	if err != nil {
		t.Fatalf("failed to create processor1: %v", err)
	}

	processor2, err := New(key)
	if err != nil {
		t.Fatalf("failed to create processor2: %v", err)
	}

	data := []byte("test data for different processors")

	// Encrypt with processor1
	encrypted, err := processor1.Encrypt(data)
	if err != nil {
		t.Errorf("processor1.Encrypt() error = %v", err)
		return
	}

	// Decrypt with processor2 (same key)
	decrypted, err := processor2.Decrypt(encrypted)
	if err != nil {
		t.Errorf("processor2.Decrypt() error = %v", err)
		return
	}

	if !bytes.Equal(data, decrypted) {
		t.Error("data mismatch between different processors with same key")
	}
}

func TestProcessor_EncryptDecrypt_DifferentKeys(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	_, err := rand.Read(key1)
	if err != nil {
		t.Fatalf("failed to generate key1: %v", err)
	}
	_, err = rand.Read(key2)
	if err != nil {
		t.Fatalf("failed to generate key2: %v", err)
	}

	processor1, err := New(key1)
	if err != nil {
		t.Fatalf("failed to create processor1: %v", err)
	}

	processor2, err := New(key2)
	if err != nil {
		t.Fatalf("failed to create processor2: %v", err)
	}

	data := []byte("test data for different keys")

	// Encrypt with processor1
	encrypted, err := processor1.Encrypt(data)
	if err != nil {
		t.Errorf("processor1.Encrypt() error = %v", err)
		return
	}

	// Try to decrypt with processor2 (different key) - should fail
	_, err = processor2.Decrypt(encrypted)
	if err == nil {
		t.Error("processor2.Decrypt() should have failed with different key")
	}
}

func TestProcessor_Decrypt_InvalidData(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	processor, err := New(key)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "random invalid data",
			data: []byte("this is not encrypted data"),
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "truncated data",
			data: []byte{0x01, 0x02},
		},
		{
			name: "malformed encoded data",
			data: []byte("!@#$%^&*()"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.Decrypt(tt.data)
			if err == nil {
				t.Error("Decrypt() should have failed with invalid data")
			}
		})
	}
}

func TestProcessor_EncryptDecrypt_Concurrent(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	processor, err := New(key)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	data := []byte("concurrent test data")
	numGoroutines := 10

	// Channel to collect results
	results := make(chan error, numGoroutines)

	// Start multiple goroutines doing encrypt/decrypt
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Each goroutine does multiple rounds
			for range 5 {
				encrypted, err := processor.Encrypt(data)
				if err != nil {
					results <- err
					return
				}

				decrypted, err := processor.Decrypt(encrypted)
				if err != nil {
					results <- err
					return
				}

				if !bytes.Equal(data, decrypted) {
					results <- err
					return
				}
			}
			results <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent test failed: %v", err)
		}
	}
}

func BenchmarkProcessor_Encrypt(b *testing.B) {
	key := make([]byte, 32)
	rand.Read(key)

	processor, err := New(key)
	if err != nil {
		b.Fatalf("failed to create processor: %v", err)
	}

	data := []byte("benchmark data for encryption performance testing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.Encrypt(data)
		if err != nil {
			b.Fatalf("Encrypt failed: %v", err)
		}
	}
}

func BenchmarkProcessor_Decrypt(b *testing.B) {
	key := make([]byte, 32)
	rand.Read(key)

	processor, err := New(key)
	if err != nil {
		b.Fatalf("failed to create processor: %v", err)
	}

	data := []byte("benchmark data for decryption performance testing")
	encrypted, err := processor.Encrypt(data)
	if err != nil {
		b.Fatalf("failed to encrypt test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.Decrypt(encrypted)
		if err != nil {
			b.Fatalf("Decrypt failed: %v", err)
		}
	}
}

func BenchmarkProcessor_EncryptDecrypt(b *testing.B) {
	key := make([]byte, 32)
	rand.Read(key)

	processor, err := New(key)
	if err != nil {
		b.Fatalf("failed to create processor: %v", err)
	}

	data := []byte("benchmark data for full encrypt/decrypt cycle performance testing")

	for b.Loop() {
		encrypted, err := processor.Encrypt(data)
		if err != nil {
			b.Fatalf("Encrypt failed: %v", err)
		}

		_, err = processor.Decrypt(encrypted)
		if err != nil {
			b.Fatalf("Decrypt failed: %v", err)
		}
	}
}

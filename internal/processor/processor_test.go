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
	}{
		{
			name:    "valid key length",
			keyLen:  64,
			wantErr: false,
		},
		{
			name:    "key too short",
			keyLen:  32,
			wantErr: true,
		},
		{
			name:    "key minimum length",
			keyLen:  64,
			wantErr: false,
		},
		{
			name:    "key longer than minimum",
			keyLen:  128,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLen)
			if _, err := rand.Read(key); err != nil {
				t.Fatalf("failed to generate random key: %v", err)
			}

			processor, err := New(key)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if processor != nil {
					t.Error("expected nil processor when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if processor == nil {
					t.Error("expected non-nil processor")
				}
			}
		})
	}
}

func TestProcessor_EncryptDecrypt(t *testing.T) {
	// Generate a valid key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
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
			data: []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."),
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "repeated pattern",
			data: bytes.Repeat([]byte("test"), 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := processor.Encrypt(tt.data)
			if err != nil {
				t.Fatalf("encryption failed: %v", err)
			}

			if len(encrypted) == 0 && len(tt.data) > 0 {
				t.Error("encrypted data should not be empty for non-empty input")
			}

			// Encrypted data should be different from original (unless original is empty)
			if len(tt.data) > 0 && bytes.Equal(encrypted, tt.data) {
				t.Error("encrypted data should differ from original data")
			}

			// Test decryption
			decrypted, err := processor.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("decryption failed: %v", err)
			}

			// Decrypted data should match original
			if !bytes.Equal(decrypted, tt.data) {
				t.Errorf("decrypted data does not match original\noriginal: %v\ndecrypted: %v", tt.data, decrypted)
			}
		})
	}
}

func TestProcessor_EncryptDecrypt_LargeData(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	processor, err := New(key)
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	// Test with 1MB of data
	largeData := make([]byte, 1024*1024)
	if _, err := rand.Read(largeData); err != nil {
		t.Fatalf("failed to generate large random data: %v", err)
	}

	encrypted, err := processor.Encrypt(largeData)
	if err != nil {
		t.Fatalf("encryption of large data failed: %v", err)
	}

	decrypted, err := processor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decryption of large data failed: %v", err)
	}

	if !bytes.Equal(decrypted, largeData) {
		t.Error("decrypted large data does not match original")
	}
}

func TestProcessor_Decrypt_InvalidData(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
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
			data: []byte("invalid encrypted data"),
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "random bytes",
			data: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.Decrypt(tt.data)
			if err == nil {
				t.Error("expected decryption to fail with invalid data")
			}
		})
	}
}

func TestProcessor_MultipleInstances(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	// Create two processors with the same key
	processor1, err := New(key)
	if err != nil {
		t.Fatalf("failed to create first processor: %v", err)
	}

	processor2, err := New(key)
	if err != nil {
		t.Fatalf("failed to create second processor: %v", err)
	}

	data := []byte("test data for multiple instances")

	// Encrypt with first processor
	encrypted, err := processor1.Encrypt(data)
	if err != nil {
		t.Fatalf("encryption with first processor failed: %v", err)
	}

	// Decrypt with second processor
	decrypted, err := processor2.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("decryption with second processor failed: %v", err)
	}

	if !bytes.Equal(decrypted, data) {
		t.Error("data encrypted by one processor should be decryptable by another with same key")
	}
}

func BenchmarkProcessor_Encrypt(b *testing.B) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		b.Fatalf("failed to generate random key: %v", err)
	}

	processor, err := New(key)
	if err != nil {
		b.Fatalf("failed to create processor: %v", err)
	}

	data := make([]byte, 1024) // 1KB of data
	if _, err := rand.Read(data); err != nil {
		b.Fatalf("failed to generate random data: %v", err)
	}

	for b.Loop() {
		_, err := processor.Encrypt(data)
		if err != nil {
			b.Fatalf("encryption failed: %v", err)
		}
	}
}

func BenchmarkProcessor_Decrypt(b *testing.B) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		b.Fatalf("failed to generate random key: %v", err)
	}

	processor, err := New(key)
	if err != nil {
		b.Fatalf("failed to create processor: %v", err)
	}

	data := make([]byte, 1024) // 1KB of data
	if _, err := rand.Read(data); err != nil {
		b.Fatalf("failed to generate random data: %v", err)
	}

	encrypted, err := processor.Encrypt(data)
	if err != nil {
		b.Fatalf("encryption failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.Decrypt(encrypted)
		if err != nil {
			b.Fatalf("decryption failed: %v", err)
		}
	}
}

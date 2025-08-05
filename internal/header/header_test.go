package header

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"testing"
)

func TestNewHeader(t *testing.T) {
	key := make([]byte, 32)
	salt := make([]byte, SaltSize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(salt)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		salt         []byte
		originalSize uint64
		key          []byte
		wantErr      bool
		errType      error
	}{
		{
			name:         "valid header",
			salt:         salt,
			originalSize: 1024,
			key:          key,
			wantErr:      false,
		},
		{
			name:         "invalid salt size",
			salt:         make([]byte, SaltSize-1),
			originalSize: 1024,
			key:          key,
			wantErr:      true,
			errType:      ErrInvalidSalt,
		},
		{
			name:         "empty key",
			salt:         salt,
			originalSize: 1024,
			key:          []byte{},
			wantErr:      true,
		},
		{
			name:         "weak salt (all zeros)",
			salt:         make([]byte, SaltSize),
			originalSize: 1024,
			key:          key,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.salt, tt.originalSize, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("New() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestHeaderSerialization(t *testing.T) {
	key := make([]byte, 32)
	salt := make([]byte, SaltSize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(salt)
	if err != nil {
		t.Fatal(err)
	}

	originalSize := uint64(1024)
	hdr, err := New(salt, originalSize, key)
	if err != nil {
		t.Fatal(err)
	}

	// Test WriteTo/ReadFrom roundtrip
	var buf bytes.Buffer
	_, err = hdr.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	readHdr, _, err := ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// Verify all fields match
	if !bytes.Equal(hdr.Salt(), readHdr.Salt()) {
		t.Error("salt mismatch after roundtrip")
	}
	if hdr.OriginalSize() != readHdr.OriginalSize() {
		t.Error("originalSize mismatch after roundtrip")
	}
	if !bytes.Equal(hdr.Nonce(), readHdr.Nonce()) {
		t.Error("nonce mismatch after roundtrip")
	}
	if err := readHdr.VerifyKey(key); err != nil {
		t.Errorf("verify failed after roundtrip: %v", err)
	}
}

func TestHeaderTampering(t *testing.T) {
	key := make([]byte, 32)
	salt := make([]byte, SaltSize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(salt)
	if err != nil {
		t.Fatal(err)
	}

	hdr, err := New(salt, 1024, key)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_, err = hdr.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with the data in various ways
	tests := []struct {
		name    string
		tamper  func([]byte)
		wantErr error
	}{
		{
			name: "corrupt magic bytes",
			tamper: func(b []byte) {
				b[0] = 'X'
			},
			wantErr: ErrInvalidMagic,
		},
		{
			name: "corrupt salt",
			tamper: func(b []byte) {
				b[len(MagicBytes)] = 0xFF
			},
			wantErr: ErrChecksumMismatch,
		},
		{
			name: "corrupt size",
			tamper: func(b []byte) {
				offset := len(MagicBytes) + SaltSize
				b[offset] = 0xFF
			},
			wantErr: ErrChecksumMismatch,
		},
		{
			name: "corrupt nonce",
			tamper: func(b []byte) {
				offset := len(MagicBytes) + SaltSize + OriginalSizeBytes
				b[offset] = 0xFF
			},
			wantErr: ErrChecksumMismatch,
		},
		{
			name: "corrupt integrity hash",
			tamper: func(b []byte) {
				offset := len(MagicBytes) + SaltSize + OriginalSizeBytes + NonceSize
				b[offset] = 0xFF
			},
			wantErr: ErrChecksumMismatch,
		},
		{
			name: "corrupt auth tag",
			tamper: func(b []byte) {
				offset := len(MagicBytes) + SaltSize + OriginalSizeBytes + NonceSize + IntegritySize
				b[offset] = 0xFF
			},
			wantErr: ErrChecksumMismatch,
		},
		{
			name: "corrupt checksum",
			tamper: func(b []byte) {
				b[len(b)-1] = 0xFF
			},
			wantErr: ErrChecksumMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buf.Bytes()
			copyData := make([]byte, len(data))
			copy(copyData, data)

			tt.tamper(copyData)

			_, _, err := ReadFrom(bytes.NewReader(copyData))
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestHeaderVerification(t *testing.T) {
	key := make([]byte, 32)
	wrongKey := make([]byte, 32)
	salt := make([]byte, SaltSize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(wrongKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(salt)
	if err != nil {
		t.Fatal(err)
	}

	hdr, err := New(salt, 1024, key)
	if err != nil {
		t.Fatal(err)
	}

	// Test with correct key
	if err := hdr.VerifyKey(key); err != nil {
		t.Errorf("verification failed with correct key: %v", err)
	}

	// Test with wrong key
	if err := hdr.VerifyKey(wrongKey); !errors.Is(err, ErrAuthFailure) {
		t.Errorf("expected ErrAuthFailure with wrong key, got %v", err)
	}

	// Test with nil key
	if err := hdr.VerifyKey(nil); err == nil {
		t.Error("expected error with nil key")
	}
}

func TestIncompleteRead(t *testing.T) {
	key := make([]byte, 32)
	salt := make([]byte, SaltSize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(salt)
	if err != nil {
		t.Fatal(err)
	}

	hdr, err := New(salt, 1024, key)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_, err = hdr.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// Test with partial data
	partialData := buf.Bytes()[:TotalSize-1]
	_, _, err = ReadFrom(bytes.NewReader(partialData))
	if !errors.Is(err, ErrIncompleteRead) {
		t.Errorf("expected ErrIncompleteRead, got %v", err)
	}
}

func TestWeakSaltDetection(t *testing.T) {
	tests := []struct {
		name   string
		salt   []byte
		isWeak bool
	}{
		{
			name:   "all zeros",
			salt:   make([]byte, SaltSize),
			isWeak: true,
		},
		{
			name:   "repeating pattern",
			salt:   bytes.Repeat([]byte{0x01, 0x02, 0x03, 0x04}, SaltSize/4),
			isWeak: true,
		},
		{
			name:   "random salt",
			salt:   func() []byte { b := make([]byte, SaltSize); rand.Read(b); return b }(),
			isWeak: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isWeakSalt(tt.salt) != tt.isWeak {
				t.Errorf("isWeakSalt() = %v, want %v", !tt.isWeak, tt.isWeak)
			}
		})
	}
}

func TestHeaderSize(t *testing.T) {
	if TotalSize != 128 {
		t.Errorf("expected TotalSize to be 128 bytes, got %d", TotalSize)
	}
}

type errorWriter struct{}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrShortWrite
}

func TestWriteError(t *testing.T) {
	key := make([]byte, 32)
	salt := make([]byte, SaltSize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rand.Read(salt)
	if err != nil {
		t.Fatal(err)
	}

	hdr, err := New(salt, 1024, key)
	if err != nil {
		t.Fatal(err)
	}

	_, err = hdr.WriteTo(&errorWriter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Check that it's either ErrIncompleteWrite or the underlying io.ErrShortWrite
	if !errors.Is(err, ErrIncompleteWrite) && !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("expected ErrIncompleteWrite or io.ErrShortWrite, got %v", err)
	}
}

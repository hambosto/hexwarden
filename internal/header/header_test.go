package header

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestNewHeader_ValidInput(t *testing.T) {
	salt := make([]byte, SaltSize)
	_, _ = rand.Read(salt)
	key := []byte("this is a very secure key with enough length")
	originalSize := uint64(123456789)

	h, err := New(salt, originalSize, key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := h.Validate(); err != nil {
		t.Errorf("expected valid header, got error: %v", err)
	}
	if !h.VerifyPassword(key) {
		t.Error("expected password verification to succeed")
	}
}

func TestNewHeader_InvalidSalt(t *testing.T) {
	key := []byte("secure key")
	salt := []byte("short")
	_, err := New(salt, 0, key)
	if err == nil {
		t.Fatal("expected error for invalid salt size, got nil")
	}
}

func TestHeader_WriteAndRead(t *testing.T) {
	salt := make([]byte, SaltSize)
	_, _ = rand.Read(salt)
	key := []byte("another secure key with enough length")
	originalSize := uint64(987654321)

	h, err := New(salt, originalSize, key)
	if err != nil {
		t.Fatalf("unexpected error creating header: %v", err)
	}

	var buf bytes.Buffer
	if err := h.Write(&buf); err != nil {
		t.Fatalf("unexpected error writing header: %v", err)
	}

	readHeader, err := Read(&buf)
	if err != nil {
		t.Fatalf("unexpected error reading header: %v", err)
	}

	if !bytes.Equal(readHeader.Salt, h.Salt) {
		t.Error("salt mismatch")
	}
	if readHeader.OriginalSize != h.OriginalSize {
		t.Errorf("original size mismatch: got %d, want %d", readHeader.OriginalSize, h.OriginalSize)
	}
	if !bytes.Equal(readHeader.VerificationHash, h.VerificationHash) {
		t.Error("verification hash mismatch")
	}
	if !readHeader.VerifyPassword(key) {
		t.Error("password verification failed after read")
	}
}

func TestVerifyPassword_Failure(t *testing.T) {
	salt := make([]byte, SaltSize)
	_, _ = rand.Read(salt)
	key := []byte("correct key")
	wrongKey := []byte("wrong key")

	h, err := New(salt, 1000, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if h.VerifyPassword(wrongKey) {
		t.Error("expected password verification to fail with wrong key")
	}
}

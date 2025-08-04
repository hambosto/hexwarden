package kdf

import (
	"bytes"
	"errors"
	"testing"
)

func TestDeriveKey_Valid(t *testing.T) {
	password := []byte("correct horse battery staple")
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	key1, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}
	if len(key1) != int(keyLength) {
		t.Errorf("unexpected key length: got %d, want %d", len(key1), keyLength)
	}

	// Deterministic output for same password and salt
	key2, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}
	if !bytes.Equal(key1, key2) {
		t.Error("keys derived from the same input should be equal")
	}
}

func TestDeriveKey_EmptyPassword(t *testing.T) {
	salt := make([]byte, saltLength)
	_, err := DeriveKey([]byte(""), salt)
	if err != ErrEmptyPassword {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}

func TestDeriveKey_InvalidSaltLength(t *testing.T) {
	password := []byte("test-password")
	invalidSalt := make([]byte, saltLength-1) // too short
	_, err := DeriveKey(password, invalidSalt)
	if err == nil || !errors.Is(err, ErrInvalidSalt) {
		t.Errorf("expected ErrInvalidSalt, got %v", err)
	}
}

func TestGenerateSalt_Unique(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("GenerateSalt returned the same value twice; expected unique salts")
	}
}

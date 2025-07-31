package kdf

import (
	"bytes"
	"errors"
	"testing"
)

func TestDeriveKey_Success(t *testing.T) {
	password := []byte("securepassword")
	salt := make([]byte, SaltLength())
	for i := range salt {
		salt[i] = byte(i)
	}

	key, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(key) != KeyLength() {
		t.Errorf("expected key length %d, got %d", KeyLength(), len(key))
	}
}

func TestDeriveKey_EmptyPassword(t *testing.T) {
	salt := make([]byte, SaltLength())
	_, err := DeriveKey([]byte(""), salt)
	if err == nil || !errors.Is(err, ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}

func TestDeriveKey_InvalidSalt(t *testing.T) {
	password := []byte("test")

	shortSalt := make([]byte, SaltLength()-1)
	_, err := DeriveKey(password, shortSalt)
	if err == nil || !errors.Is(err, ErrInvalidSalt) {
		t.Errorf("expected ErrInvalidSalt for short salt, got %v", err)
	}

	longSalt := make([]byte, SaltLength()+1)
	_, err = DeriveKey(password, longSalt)
	if err == nil || !errors.Is(err, ErrInvalidSalt) {
		t.Errorf("expected ErrInvalidSalt for long salt, got %v", err)
	}
}

func TestDeriveKey_Consistency(t *testing.T) {
	password := []byte("reproducible")
	salt := make([]byte, SaltLength())
	for i := range salt {
		salt[i] = byte(i + 1)
	}

	key1, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatal(err)
	}

	key2, err := DeriveKey(password, salt)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("derived keys with same input do not match")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt failed: %v", err)
	}

	if len(salt1) != SaltLength() || len(salt2) != SaltLength() {
		t.Errorf("expected salt length %d, got %d and %d", SaltLength(), len(salt1), len(salt2))
	}

	if bytes.Equal(salt1, salt2) {
		t.Error("two generated salts should not be equal")
	}
}

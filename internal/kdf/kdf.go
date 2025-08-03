// Package kdf provides a simple, secure key derivation function using scrypt.
package kdf

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

var (
	ErrEmptyPassword = errors.New("kdf: password cannot be empty")
	ErrInvalidSalt   = errors.New("kdf: invalid salt length")
)

const (
	// Heavy scrypt parameters for maximum security
	scryptN    = 1048576 // 2^20, CPU/memory cost
	scryptR    = 8       // block size
	scryptP    = 1       // parallelization
	keyLength  = 32      // 256-bit key
	saltLength = 32      // 256-bit salt
)

// DeriveKey derives a cryptographic key from password and salt using scrypt.
// The salt must be exactly 32 bytes. Returns a 32-byte key.
func DeriveKey(password, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, ErrEmptyPassword
	}

	if len(salt) != saltLength {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidSalt, saltLength, len(salt))
	}

	key, err := scrypt.Key(password, salt, scryptN, scryptR, scryptP, keyLength)
	if err != nil {
		return nil, fmt.Errorf("kdf: key derivation failed: %w", err)
	}

	return key, nil
}

// GenerateSalt creates a cryptographically secure 32-byte random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("kdf: failed to generate salt: %w", err)
	}
	return salt, nil
}

// SaltLength returns the required salt length in bytes.
func SaltLength() int {
	return saltLength
}

// KeyLength returns the derived key length in bytes.
func KeyLength() int {
	return keyLength
}

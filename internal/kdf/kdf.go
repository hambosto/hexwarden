// Package kdf provides a secure key derivation function using Argon2id.
package kdf

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// Errors returned by the package.
var (
	ErrEmptyPassword = errors.New("kdf: password cannot be empty")
	ErrInvalidSalt   = errors.New("kdf: invalid salt length")
)

// Recommended Argon2id parameters based on OWASP guidance:
// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
const (
	// Number of iterations (time cost).
	argonTime uint32 = 3

	// Memory usage in KiB (64 MiB).
	argonMemory uint32 = 64 * 1024

	// Number of threads (parallelism).
	argonThreads uint8 = 4

	// Desired length of the derived key in bytes.
	keyLength uint32 = 32

	// Length of the salt in bytes.
	saltLength = 32
)

// DeriveKey derives a key from the given password and salt using Argon2id.
func DeriveKey(password, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, ErrEmptyPassword
	}
	if len(salt) != saltLength {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidSalt, saltLength, len(salt))
	}

	key := argon2.IDKey(password, salt, argonTime, argonMemory, argonThreads, keyLength)
	return key, nil
}

// GenerateSalt generates a new cryptographically secure random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("kdf: failed to generate salt: %w", err)
	}
	return salt, nil
}

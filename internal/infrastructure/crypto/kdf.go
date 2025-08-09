package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/utils"
)

// DeriveKey derives a key from the given password and salt using Argon2id
func DeriveKey(password, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, constants.ErrEmptyPassword
	}
	if len(salt) != constants.SaltSize {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", constants.ErrInvalidSalt, constants.SaltSize, len(salt))
	}

	key := argon2.IDKey(
		password,
		salt,
		constants.ArgonTime,
		constants.ArgonMemory,
		constants.ArgonThreads,
		uint32(constants.KeySize),
	)
	return key, nil
}

// GenerateSalt generates a new cryptographically secure random salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, constants.SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("%w: %v", constants.ErrSaltGeneration, err)
	}
	return salt, nil
}

// ValidateSalt checks if the salt is valid and secure
func ValidateSalt(salt []byte) error {
	if len(salt) != constants.SaltSize {
		return fmt.Errorf("%w: expected %d bytes, got %d", constants.ErrInvalidSalt, constants.SaltSize, len(salt))
	}

	if isWeakSalt(salt) {
		return fmt.Errorf("weak salt detected - use cryptographically random salt")
	}

	return nil
}

// isWeakSalt checks for weak salt patterns that may indicate non-random sources
func isWeakSalt(salt []byte) bool {
	if len(salt) != constants.SaltSize {
		return true
	}

	// Check for all zero bytes
	allZero := true
	for _, b := range salt {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return true
	}

	// Check for repeating 4-byte patterns
	if len(salt) >= 4 {
		pattern := salt[:4]
		for i := 4; i < len(salt); i += 4 {
			end := utils.MinInt(i+4, len(salt))
			if !bytesEqual(pattern, salt[i:end]) {
				return false
			}
		}
		return true
	}

	return false
}

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

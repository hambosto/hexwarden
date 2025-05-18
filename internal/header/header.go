package header

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

const (
	SaltSize          = 32
	OriginalSizeBytes = 8
	NonceSize         = 12
	VerificationSize  = 32
)

type Header struct {
	Salt             []byte
	OriginalSize     uint64
	Nonce            []byte
	VerificationHash []byte
}

func NewHeader(salt []byte, originalSize uint64, nonce []byte, key []byte) (Header, error) {
	verification := generateVerificationHash(salt, key)

	h := Header{
		Salt:             salt,
		OriginalSize:     originalSize,
		Nonce:            nonce,
		VerificationHash: verification,
	}

	if err := h.Validate(); err != nil {
		return Header{}, err
	}

	return h, nil
}

func (h Header) Validate() error {
	if len(h.Salt) != SaltSize {
		return fmt.Errorf("invalid salt size: got %d, want %d", len(h.Salt), SaltSize)
	}

	if len(h.Nonce) != NonceSize {
		return fmt.Errorf("invalid cipher nonce size: got %d, want %d", len(h.Nonce), NonceSize)
	}

	if len(h.VerificationHash) != VerificationSize {
		return fmt.Errorf("invalid verification hash size: got %d, want %d", len(h.VerificationHash), VerificationSize)
	}

	return nil
}

func (h Header) VerifyPassword(key []byte) bool {
	expectedHash := generateVerificationHash(h.Salt, key)

	return hmac.Equal(h.VerificationHash, expectedHash)
}

func generateVerificationHash(salt, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(salt)
	return mac.Sum(nil)
}

package header

import (
	"fmt"
)

const (
	SaltSize          = 32
	OriginalSizeBytes = 8
	NonceSize         = 12
)

type Header struct {
	Salt         []byte
	OriginalSize uint64
	Nonce        []byte
}

func NewHeader(salt []byte, originalSize uint64, Nonce []byte) (Header, error) {
	h := Header{
		Salt:         salt,
		OriginalSize: originalSize,
		Nonce:        Nonce,
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

	return nil
}

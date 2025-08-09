package utils

import (
	"github.com/hambosto/hexwarden/internal/constants"
)

// Padder handles PKCS7 padding and unpadding operations
type Padder struct {
	blockSize int
}

// NewPadder creates a new padder with the specified block size
func NewPadder(blockSize int) (*Padder, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, constants.ErrPaddingFailed
	}

	return &Padder{
		blockSize: blockSize,
	}, nil
}

// NewDefaultPadder creates a new padder with default block size
func NewDefaultPadder() (*Padder, error) {
	return NewPadder(constants.PaddingSize)
}

// Pad applies PKCS7 padding to the input data
func (p *Padder) Pad(data []byte) ([]byte, error) {
	if data == nil {
		return nil, constants.ErrPaddingFailed
	}

	padding := p.blockSize - (len(data) % p.blockSize)
	padText := make([]byte, padding)

	for i := range padText {
		padText[i] = byte(padding)
	}

	return append(data, padText...), nil
}

// Unpad removes PKCS7 padding from the input data
func (p *Padder) Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, constants.ErrUnpaddingFailed
	}

	if len(data)%p.blockSize != 0 {
		return nil, constants.ErrUnpaddingFailed
	}

	padding := int(data[len(data)-1])

	if padding == 0 || padding > p.blockSize {
		return nil, constants.ErrUnpaddingFailed
	}

	if padding > len(data) {
		return nil, constants.ErrUnpaddingFailed
	}

	// Verify padding bytes
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, constants.ErrUnpaddingFailed
		}
	}

	return data[:len(data)-padding], nil
}

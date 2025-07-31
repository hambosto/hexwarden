// Package padding provides PKCS#7 padding functionality with length headers.
package padding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	// ErrDataTooLarge is returned when data exceeds uint32 maximum size.
	ErrDataTooLarge = errors.New("data too large to encode length as uint32")

	// ErrInvalidBlockSize is returned when block size is not between 1 and 255.
	ErrInvalidBlockSize = errors.New("invalid block size: must be between 1 and 255")

	// ErrInvalidDataLength is returned when data length is not a multiple of block size.
	ErrInvalidDataLength = errors.New("invalid data length: must be multiple of block size")

	// ErrInvalidPadding is returned when PKCS#7 padding is malformed.
	ErrInvalidPadding = errors.New("invalid PKCS#7 padding")
)

// Padder provides PKCS#7 padding with length headers.
type Padder struct {
	blockSize int
}

// NewPKCS7 creates a new PKCS#7 padder with the specified block size.
// Block size must be between 1 and 255 bytes inclusive.
func NewPKCS7(blockSize int) (*Padder, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidBlockSize, blockSize)
	}
	return &Padder{blockSize: blockSize}, nil
}

// Pad adds a 4-byte big-endian length header followed by PKCS#7 padding to data.
// The length header contains the original data length as uint32.
func (p *Padder) Pad(data []byte) ([]byte, error) {
	const maxUint32 = 1<<32 - 1

	if len(data) > maxUint32 {
		return nil, fmt.Errorf("%w: %d bytes", ErrDataTooLarge, len(data))
	}

	// Create length header
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(len(data)))

	// Combine header and data
	combined := make([]byte, 0, len(header)+len(data)+p.blockSize)
	combined = append(combined, header...)
	combined = append(combined, data...)

	// Calculate padding length
	paddingLen := p.blockSize - (len(combined) % p.blockSize)
	if paddingLen == 0 {
		paddingLen = p.blockSize
	}

	// Add PKCS#7 padding
	padding := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)
	return append(combined, padding...), nil
}

// Unpad removes PKCS#7 padding and extracts the original data.
// Returns the original data without the length header and padding.
func (p *Padder) Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data)%p.blockSize != 0 {
		return nil, fmt.Errorf("%w: got %d, block size %d", ErrInvalidDataLength, len(data), p.blockSize)
	}

	// Validate and extract padding
	paddingLen := int(data[len(data)-1])
	if paddingLen == 0 || paddingLen > p.blockSize || paddingLen > len(data) {
		return nil, fmt.Errorf("%w: padding length %d", ErrInvalidPadding, paddingLen)
	}

	// Verify all padding bytes are correct
	paddingStart := len(data) - paddingLen
	for i := paddingStart; i < len(data); i++ {
		if data[i] != byte(paddingLen) {
			return nil, fmt.Errorf("%w: incorrect padding byte at position %d",
				ErrInvalidPadding, i)
		}
	}

	// Remove padding to get header + original data
	withoutPadding := data[:paddingStart]

	// Must have at least 4 bytes for the length header
	if len(withoutPadding) < 4 {
		return nil, fmt.Errorf("%w: insufficient data for length header", ErrInvalidPadding)
	}

	// Extract and validate original data length
	originalLen := binary.BigEndian.Uint32(withoutPadding[:4])
	if int(originalLen) != len(withoutPadding)-4 {
		return nil, fmt.Errorf("%w: length header mismatch", ErrInvalidPadding)
	}

	// Return original data (without header)
	return withoutPadding[4:], nil
}

// BlockSize returns the block size used by this padder.
func (p *Padder) BlockSize() int {
	return p.blockSize
}

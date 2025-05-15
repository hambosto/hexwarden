package processor

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/hambosto/hexwarden/internal/compression"
)

func (c *ChunkProcessor) encrypt(chunk []byte) ([]byte, error) {
	compressedData, err := compression.CompressData(chunk)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	compressedSize := len(compressedData)
	if compressedSize < 0 || compressedSize > math.MaxUint32 {
		return nil, fmt.Errorf("compressed data size too large: %d", compressedSize)
	}

	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(compressedSize))
	fullPayload := append(sizeHeader, compressedData...)

	// Pad to 16-byte boundary
	alignedSize := (len(fullPayload) + 15) & ^15
	paddedPayload := make([]byte, alignedSize)
	copy(paddedPayload, fullPayload)

	encrypted, err := c.Cipher.Encrypt(paddedPayload)
	if err != nil {
		return nil, fmt.Errorf("cipher encryption failed: %w", err)
	}

	encoded, err := c.Encoding.Encode(encrypted)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	return encoded, nil
}

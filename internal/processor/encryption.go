package processor

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/padding"
)

func (c *ChunkProcessor) Encryption(chunk []byte) ([]byte, error) {
	compressedData, err := c.Compression.Compress(chunk)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	paddedData, err := padding.PadPKCS7(compressedData, 16)
	if err != nil {
		return nil, fmt.Errorf("padding failed: %w", err)
	}

	encrypted, err := c.Cipher.Encrypt(paddedData)
	if err != nil {
		return nil, fmt.Errorf("cipher encryption failed: %w", err)
	}

	encoded, err := c.Encoding.Encode(encrypted)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	return encoded, nil
}

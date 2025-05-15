package processor

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/compression"
)

func (c *ChunkProcessor) decrypt(chunk []byte) ([]byte, error) {
	decodedData, err := c.Encoding.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	decrypted, err := c.Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("cipher decryption failed: %w", err)
	}

	decompressedData, err := compression.DecompressData(decrypted)
	if err != nil {
		return nil, fmt.Errorf("zlib decompression failed: %w", err)
	}

	return decompressedData, nil
}

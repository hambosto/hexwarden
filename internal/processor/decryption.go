package processor

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/padding"
)

func (c *ChunkProcessor) Decryption(chunk []byte) ([]byte, error) {
	decodedData, err := c.Encoding.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	decrypted, err := c.Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("cipher decryption failed: %w", err)
	}

	unpaddedData, err := padding.UnpadPKCS7(decrypted, 16)
	if err != nil {
		return nil, fmt.Errorf("padding failed: %w", err)
	}

	decompressedData, err := c.Compression.Decompress(unpaddedData)
	if err != nil {
		return nil, fmt.Errorf("zlib decompression failed: %w", err)
	}

	return decompressedData, nil
}

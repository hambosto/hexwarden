package processor

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/cipher"
	"github.com/hambosto/hexwarden/internal/compression"
	"github.com/hambosto/hexwarden/internal/encoding"
)

type ChunkProcessor struct {
	Cipher      *cipher.Cipher
	Encoding    *encoding.Encoding
	Compression *compression.ZlibCompressor
}

func NewChunkProcessor(key []byte) (*ChunkProcessor, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("encryption key must be at least 64 bytes long")
	}

	cipher, err := cipher.NewCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	encoding, err := encoding.NewEncoding(4, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	compression, err := compression.NewZlibCompressor(compression.LevelBestSpeed)
	if err != nil {
		return nil, fmt.Errorf("failed to create compression: %w", err)
	}

	return &ChunkProcessor{
		Cipher:      cipher,
		Encoding:    encoding,
		Compression: compression,
	}, nil
}

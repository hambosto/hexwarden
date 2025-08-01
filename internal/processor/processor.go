package processor

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/cipher"
	"github.com/hambosto/hexwarden/internal/compression"
	"github.com/hambosto/hexwarden/internal/encoding"
	"github.com/hambosto/hexwarden/internal/padding"
)

const (
	minKeyLength = 64
	cipherKeyLen = 32
	encodingBase = 4
	encodingPow  = 10
	paddingSize  = 16
)

// Processor handles encryption/decryption operations with compression, padding, and encoding.
type Processor struct {
	cipher      *cipher.Cipher
	encoder     *encoding.Encoder
	compression *compression.Compression
	padder      *padding.Padder
}

// New creates a new Processor with the provided encryption key.
// The key must be at least 64 bytes long.
func New(key []byte) (*Processor, error) {
	if len(key) < minKeyLength {
		return nil, fmt.Errorf("encryption key must be at least %d bytes long", minKeyLength)
	}

	cipher, err := cipher.New(key[:cipherKeyLen])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	encoder, err := encoding.New(encodingBase, encodingPow)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	compression, err := compression.New(compression.LevelBestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed to create compression: %w", err)
	}

	padder, err := padding.New(paddingSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create padder: %w", err)
	}

	return &Processor{
		cipher:      cipher,
		encoder:     encoder,
		compression: compression,
		padder:      padder,
	}, nil
}

// Encrypt compresses, pads, encrypts, and encodes the input data.
func (p *Processor) Encrypt(data []byte) ([]byte, error) {
	compressed, err := p.compression.Compress(data)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	padded, err := p.padder.Pad(compressed)
	if err != nil {
		return nil, fmt.Errorf("padding failed: %w", err)
	}

	encrypted, err := p.cipher.Encrypt(padded)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	encoded, err := p.encoder.Encode(encrypted)
	if err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return encoded, nil
}

// Decrypt decodes, decrypts, unpads, and decompresses the input data.
func (p *Processor) Decrypt(data []byte) ([]byte, error) {
	decoded, err := p.encoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}

	decrypted, err := p.cipher.Decrypt(decoded)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	unpadded, err := p.padder.Unpad(decrypted)
	if err != nil {
		return nil, fmt.Errorf("unpadding failed: %w", err)
	}

	decompressed, err := p.compression.Decompress(unpadded)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return decompressed, nil
}

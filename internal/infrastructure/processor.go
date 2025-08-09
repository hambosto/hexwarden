package infrastructure

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/compression"
	"github.com/hambosto/hexwarden/internal/infrastructure/crypto"
	"github.com/hambosto/hexwarden/internal/infrastructure/encoding"
	"github.com/hambosto/hexwarden/internal/infrastructure/utils"
)

// Processor handles encryption/decryption operations with compression, padding, and encoding
type Processor struct {
	cipher     *crypto.AESCipher
	encoder    *encoding.Encoder
	compressor *compression.Compressor
	padder     *utils.Padder
}

// NewProcessor creates a new processor with the provided encryption key
func NewProcessor(key []byte) (*Processor, error) {
	if len(key) < constants.KeySize {
		return nil, fmt.Errorf("encryption key must be at least %d bytes long", constants.KeySize)
	}

	cipher, err := crypto.NewAESCipher(key[:constants.KeySize])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	encoder, err := encoding.NewDefaultEncoder()
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	compressor, err := compression.NewDefaultCompressor()
	if err != nil {
		return nil, fmt.Errorf("failed to create compressor: %w", err)
	}

	padder, err := utils.NewDefaultPadder()
	if err != nil {
		return nil, fmt.Errorf("failed to create padder: %w", err)
	}

	return &Processor{
		cipher:     cipher,
		encoder:    encoder,
		compressor: compressor,
		padder:     padder,
	}, nil
}

// Encrypt compresses, pads, encrypts, and encodes the input data
func (p *Processor) Encrypt(data []byte) ([]byte, error) {
	// Step 1: Compress the data
	compressed, err := p.compressor.Compress(data)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Step 2: Pad the compressed data
	padded, err := p.padder.Pad(compressed)
	if err != nil {
		return nil, fmt.Errorf("padding failed: %w", err)
	}

	// Step 3: Encrypt the padded data
	encrypted, err := p.cipher.Encrypt(padded)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	// Step 4: Encode the encrypted data with Reed-Solomon
	encoded, err := p.encoder.Encode(encrypted)
	if err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return encoded, nil
}

// Decrypt decodes, decrypts, unpads, and decompresses the input data
func (p *Processor) Decrypt(data []byte) ([]byte, error) {
	// Step 1: Decode the Reed-Solomon encoded data
	decoded, err := p.encoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}

	// Step 2: Decrypt the decoded data
	decrypted, err := p.cipher.Decrypt(decoded)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Step 3: Remove padding from decrypted data
	unpadded, err := p.padder.Unpad(decrypted)
	if err != nil {
		return nil, fmt.Errorf("unpadding failed: %w", err)
	}

	// Step 4: Decompress the unpadded data
	decompressed, err := p.compressor.Decompress(unpadded)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return decompressed, nil
}

package processor

import (
	"fmt"
)

func (p *Processor) Encrypt(data []byte) ([]byte, error) {
	compressed, err := p.Compressor.Compress(data)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	padded, err := p.Padding.Pad(compressed)
	if err != nil {
		return nil, fmt.Errorf("padding failed: %w", err)
	}

	encrypted, err := p.Cipher.Encrypt(padded)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	encoded, err := p.Encoder.Encode(encrypted)
	if err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return encoded, nil
}

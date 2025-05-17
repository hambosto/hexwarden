package processor

import (
	"fmt"
)

func (p *Processor) Decrypt(data []byte) ([]byte, error) {
	decoded, err := p.Encoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}

	decrypted, err := p.Cipher.Decrypt(decoded)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	unpadded, err := p.Padding.Unpad(decrypted)
	if err != nil {
		return nil, fmt.Errorf("unpadding failed: %w", err)
	}

	decompressed, err := p.Compressor.Decompress(unpadded)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return decompressed, nil
}

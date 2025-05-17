package processor

import (
	"fmt"

	"github.com/hambosto/hexwarden/internal/cipher"
	"github.com/hambosto/hexwarden/internal/compression"
	"github.com/hambosto/hexwarden/internal/encoding"
	"github.com/hambosto/hexwarden/internal/padding"
)

type Processor struct {
	Cipher     *cipher.Cipher
	Encoder    *encoding.Encoding
	Compressor *compression.ZlibCompressor
	Padding    *padding.Padding
}

func NewProcessor(key []byte) (*Processor, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("encryption key must be at least 64 bytes long")
	}

	cipher, err := cipher.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}

	encoder, err := encoding.NewEncoding(4, 10)
	if err != nil {
		return nil, err
	}

	compressor, err := compression.NewZlibCompressor(compression.LevelBestSpeed)
	if err != nil {
		return nil, err
	}

	padding, err := padding.NewPKCS7(16)
	if err != nil {
		return nil, err
	}

	return &Processor{
		Cipher:     cipher,
		Encoder:    encoder,
		Compressor: compressor,
		Padding:    padding,
	}, nil
}

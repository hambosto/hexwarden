package encoding

import (
	"errors"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

const (
	headerSize = 4
	maxDataLen = 1 << 30
)

var (
	ErrInvalidDataShards   = errors.New("data shards must be positive")
	ErrInvalidParityShards = errors.New("parity shards must be positive")
	ErrDataSize            = errors.New("invalid data size")
	ErrEncodedDataSize     = errors.New("invalid encoded data size")
	ErrCorruptedData       = errors.New("corrupted data")
)

type Encoding struct {
	dataShards   int
	parityShards int
	encoder      reedsolomon.Encoder
}

func NewEncoding(dataShards, parityShards int) (*Encoding, error) {
	if dataShards <= 0 {
		return nil, ErrInvalidDataShards
	}
	if parityShards <= 0 {
		return nil, ErrInvalidParityShards
	}

	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed-solomon encoder: %w", err)
	}

	return &Encoding{
		dataShards:   dataShards,
		parityShards: parityShards,
		encoder:      enc,
	}, nil
}

func (e *Encoding) Encode(data []byte) ([]byte, error) {
	if err := e.validate(data); err != nil {
		return nil, err
	}

	shards := e.splitIntoShards(data)
	if err := e.encoder.Encode(shards); err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return e.combineShards(shards), nil
}

func (e *Encoding) Decode(encoded []byte) ([]byte, error) {
	totalShards := e.dataShards + e.parityShards

	if len(encoded) == 0 || len(encoded)%totalShards != 0 {
		return nil, ErrEncodedDataSize
	}

	shards := e.splitEncodedData(encoded)

	if err := e.encoder.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("reconstruction failed: %w", err)
	}

	return e.extractData(shards)
}

func (e *Encoding) validate(data []byte) error {
	if len(data) == 0 || len(data) > maxDataLen {
		return fmt.Errorf("%w: must be between 1 qand %d bytes", ErrDataSize, maxDataLen)
	}
	return nil
}

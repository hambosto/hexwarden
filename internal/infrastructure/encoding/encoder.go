package encoding

import (
	"fmt"

	"github.com/klauspost/reedsolomon"

	"github.com/hambosto/hexwarden/internal/constants"
)

// Encoder handles Reed-Solomon encoding and decoding operations
type Encoder struct {
	dataShards   int
	parityShards int
	encoder      reedsolomon.Encoder
}

// NewEncoder creates a new Reed-Solomon encoder with the specified number of data and parity shards
func NewEncoder(dataShards, parityShards int) (*Encoder, error) {
	if dataShards <= 0 {
		return nil, constants.ErrEncodingFailed
	}
	if parityShards <= 0 {
		return nil, constants.ErrEncodingFailed
	}

	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed-solomon encoder: %w", err)
	}

	return &Encoder{
		dataShards:   dataShards,
		parityShards: parityShards,
		encoder:      enc,
	}, nil
}

// NewDefaultEncoder creates a new encoder with default Reed-Solomon parameters
func NewDefaultEncoder() (*Encoder, error) {
	return NewEncoder(constants.DataShards, constants.ParityShards)
}

// Encode encodes the input data using Reed-Solomon encoding
func (e *Encoder) Encode(data []byte) ([]byte, error) {
	if !e.isValidDataSize(data) {
		return nil, fmt.Errorf("%w: must be between 1 and %d bytes", constants.ErrEncodingFailed, constants.MaxDataLen)
	}

	shards := e.splitIntoShards(data)
	if err := e.encoder.Encode(shards); err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return e.combineShards(shards), nil
}

// Decode decodes the Reed-Solomon encoded data
func (e *Encoder) Decode(encoded []byte) ([]byte, error) {
	totalShards := e.dataShards + e.parityShards

	if len(encoded) == 0 || len(encoded)%totalShards != 0 {
		return nil, constants.ErrDecodingFailed
	}

	shards := e.splitEncodedData(encoded)

	if err := e.encoder.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("reconstruction failed: %w", err)
	}

	return e.extractData(shards)
}

// splitIntoShards splits input data into shards for encoding
func (e *Encoder) splitIntoShards(data []byte) [][]byte {
	totalShards := e.dataShards + e.parityShards
	shardSize := (len(data) + e.dataShards - 1) / e.dataShards

	shards := make([][]byte, totalShards)
	for i := range shards {
		shards[i] = make([]byte, shardSize)
	}

	for i, b := range data {
		shardIndex := i / shardSize
		posInShard := i % shardSize
		shards[shardIndex][posInShard] = b
	}

	return shards
}

// splitEncodedData splits encoded data back into individual shards
func (e *Encoder) splitEncodedData(data []byte) [][]byte {
	totalShards := e.dataShards + e.parityShards
	shardSize := len(data) / totalShards

	shards := make([][]byte, totalShards)
	for i := range shards {
		start := i * shardSize
		end := (i + 1) * shardSize
		shards[i] = data[start:end]
	}

	return shards
}

// combineShards combines all shards into a single byte slice
func (e *Encoder) combineShards(shards [][]byte) []byte {
	if len(shards) == 0 {
		return nil
	}

	shardSize := len(shards[0])
	result := make([]byte, shardSize*len(shards))

	for i, shard := range shards {
		copy(result[i*shardSize:], shard)
	}

	return result
}

// extractData extracts the original data from the data shards
func (e *Encoder) extractData(shards [][]byte) ([]byte, error) {
	if len(shards) < e.dataShards {
		return nil, constants.ErrDecodingFailed
	}

	shardSize := len(shards[0])
	combined := make([]byte, 0, shardSize*e.dataShards)

	for i := 0; i < e.dataShards; i++ {
		combined = append(combined, shards[i]...)
	}

	return combined, nil
}

// isValidDataSize validates the input data size
func (e *Encoder) isValidDataSize(data []byte) bool {
	return len(data) != 0 && len(data) <= constants.MaxDataLen
}

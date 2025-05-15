package encoding

import (
	"encoding/binary"
	"fmt"
	"math"
)

func (e *Encoding) splitIntoShards(data []byte) [][]byte {
	totalShards := e.dataShards + e.parityShards

	shardSize := (len(data) + e.dataShards - 1) / e.dataShards
	if shardSize%e.dataShards != 0 {
		shardSize = ((shardSize + e.dataShards - 1) / e.dataShards) * e.dataShards
	}

	shards := make([][]byte, totalShards)
	for i := range shards {
		shards[i] = make([]byte, shardSize)
	}

	for i := range data {
		shardIndex := i / shardSize
		posInShard := i % shardSize
		shards[shardIndex][posInShard] = data[i]
	}

	return shards
}

func (e *Encoding) splitEncodedData(data []byte) [][]byte {
	totalShards := e.dataShards + e.parityShards
	shardSize := len(data) / totalShards

	shards := make([][]byte, totalShards)
	for i := range shards {
		shards[i] = data[i*shardSize : (i+1)*shardSize]
	}

	return shards
}

func (e *Encoding) combineShards(shards [][]byte) []byte {
	shardSize := len(shards[0])
	totalShards := e.dataShards + e.parityShards

	result := make([]byte, shardSize*totalShards)
	for i, shard := range shards {
		copy(result[i*shardSize:], shard)
	}

	return result
}

func (e *Encoding) extractData(shards [][]byte) ([]byte, error) {
	shardSize := len(shards[0])

	combined := make([]byte, shardSize*e.dataShards)
	for i := range e.dataShards {
		copy(combined[i*shardSize:], shards[i])
	}

	if len(combined) < headerSize {
		return nil, fmt.Errorf("%w: too short", ErrCorruptedData)
	}

	originalSize := binary.BigEndian.Uint32(combined[:headerSize])
	combinedLen := len(combined) - headerSize
	if combinedLen < 0 || (originalSize > uint32(math.MaxInt32) && uint64(originalSize) > uint64(combinedLen)) {
		return nil, fmt.Errorf("%w: invalid size header", ErrCorruptedData)
	}

	return combined[headerSize : headerSize+originalSize], nil
}

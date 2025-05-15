package encoding

import (
	"encoding/binary"
	"fmt"
	"math"
)

func (e *Encoding) appendHeader(data []byte) ([]byte, error) {
	dataLen := len(data)
	// Check if data length exceeds the maximum allowed size for uint32
	if dataLen > math.MaxUint32 {
		return nil, fmt.Errorf("invalid data length: %d exceeds 4GB", dataLen)
	}

	// Safely convert the length to uint32
	dataLenUint32 := uint32(dataLen)

	// Create the result slice with the appropriate size (header size + data size)
	result := make([]byte, headerSize+dataLenUint32)

	// Write the size of the data in big-endian format
	binary.BigEndian.PutUint32(result, dataLenUint32)

	// Copy the actual data after the header
	copy(result[headerSize:], data)

	return result, nil
}

package padding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func PadPKCS7(data []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, fmt.Errorf("invalid block size: %d (must be between 1 and 255)", blockSize)
	}

	dataLen := len(data)
	if dataLen > math.MaxUint32 {
		return nil, fmt.Errorf("data too large to encode length as uint32: %d bytes", dataLen)
	}

	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(dataLen))

	combined := append(sizeHeader, data...)

	paddingLen := blockSize - (len(combined) % blockSize)
	if paddingLen == 0 {
		paddingLen = blockSize
	}

	padding := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)
	return append(combined, padding...), nil
}

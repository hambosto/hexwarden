package padding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func (p *Padding) Pad(data []byte) ([]byte, error) {
	dataLen := len(data)
	if dataLen > math.MaxUint32 {
		return nil, fmt.Errorf("data too large to encode length as uint32: %d bytes", dataLen)
	}

	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(dataLen))

	combined := append(sizeHeader, data...)

	paddingLen := p.BlockSize - (len(combined) % p.BlockSize)
	if paddingLen == 0 {
		paddingLen = p.BlockSize
	}

	padding := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)
	return append(combined, padding...), nil
}

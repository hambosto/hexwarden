package padding

import (
	"fmt"
)

func (p *PKCS7) Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data)%p.BlockSize != 0 {
		return nil, fmt.Errorf("invalid data length: %d (must be multiple of block size %d)", len(data), p.BlockSize)
	}

	paddingLen := int(data[len(data)-1])
	if paddingLen == 0 || paddingLen > p.BlockSize || paddingLen > len(data) {
		return nil, fmt.Errorf("invalid padding length: %d", paddingLen)
	}

	for i := len(data) - paddingLen; i < len(data); i++ {
		if data[i] != byte(paddingLen) {
			return nil, fmt.Errorf("invalid PKCS#7 padding at byte %d", i)
		}
	}

	return data[:len(data)-paddingLen], nil
}

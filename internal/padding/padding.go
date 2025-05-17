package padding

import (
	"fmt"
)

type PKCS7 struct {
	BlockSize int
}

func NewPKCS7(blockSize int) (*PKCS7, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, fmt.Errorf("invalid block size: %d (must be between 1 and 255)", blockSize)
	}
	return &PKCS7{BlockSize: blockSize}, nil
}

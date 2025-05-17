package padding

import (
	"fmt"
)

type Padding struct {
	BlockSize int
}

func NewPKCS7(blockSize int) (*Padding, error) {
	if blockSize <= 0 || blockSize > 255 {
		return nil, fmt.Errorf("invalid block size: %d (must be between 1 and 255)", blockSize)
	}
	return &Padding{BlockSize: blockSize}, nil
}

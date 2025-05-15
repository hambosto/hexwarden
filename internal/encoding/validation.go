package encoding

import (
	"fmt"
)

func (e *Encoding) validate(data []byte) error {
	if len(data) == 0 || len(data) > maxDataLen {
		return fmt.Errorf("%w: must be between 1 and %d bytes", ErrDataSize, maxDataLen)
	}
	return nil
}

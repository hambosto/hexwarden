package cipher

import (
	"fmt"
)

func (c *Cipher) SetNonce(nonce []byte) error {
	if len(nonce) != 12 {
		return fmt.Errorf("invalid nonce size: %d bytes", len(nonce))
	}
	c.nonce = make([]byte, len(nonce))
	copy(c.nonce, nonce)
	return nil
}

func (c *Cipher) GetNonce() []byte {
	nonce := make([]byte, len(c.nonce))
	copy(nonce, c.nonce)
	return nonce
}

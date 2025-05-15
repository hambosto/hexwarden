package header

import (
	"encoding/binary"
	"fmt"
	"io"
)

func ReadHeader(r io.Reader) (Header, error) {
	var h Header

	salt, err := read(r, SaltSize)
	if err != nil {
		return Header{}, fmt.Errorf("reading salt: %w", err)
	}
	h.Salt = salt

	sizeBuffer, err := read(r, OriginalSizeBytes)
	if err != nil {
		return Header{}, fmt.Errorf("reading original size: %w", err)
	}
	h.OriginalSize = binary.BigEndian.Uint64(sizeBuffer)

	nonce, err := read(r, NonceSize)
	if err != nil {
		return Header{}, fmt.Errorf("reading cipher nonce: %w", err)
	}
	h.Nonce = nonce

	if err := h.Validate(); err != nil {
		return Header{}, err
	}

	return h, nil
}

func read(r io.Reader, n int) ([]byte, error) {
	buffer := make([]byte, n)
	if _, err := io.ReadFull(r, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
}

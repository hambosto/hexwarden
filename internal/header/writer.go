package header

import (
	"encoding/binary"
	"fmt"
	"io"
)

func WriteHeader(w io.Writer, h Header) error {
	if err := h.Validate(); err != nil {
		return err
	}

	if err := write(w, h.Salt); err != nil {
		return fmt.Errorf("writing salt: %w", err)
	}

	sizeBuffer := make([]byte, OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuffer, h.OriginalSize)
	if err := write(w, sizeBuffer); err != nil {
		return fmt.Errorf("writing original size: %w", err)
	}

	if err := write(w, h.Nonce); err != nil {
		return fmt.Errorf("writing cipher nonce: %w", err)
	}

	return nil
}

func write(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("short write: wrote %d of %d bytes", n, len(data))
	}
	return nil
}

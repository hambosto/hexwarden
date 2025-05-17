package compression

import (
	"bytes"
	"compress/zlib"
	"fmt"
)

func (c *ZlibCompressor) Compress(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	w, err := zlib.NewWriterLevel(&buffer, c.level.getCompressionLevel())
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer: %w", err)
	}

	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data to zlib writer: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %w", err)
	}

	return buffer.Bytes(), nil
}

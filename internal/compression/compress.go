package compression

import (
	"bytes"
	"compress/zlib"
	"fmt"
)

func CompressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	// Create a zlib writer with best compression
	w, err := zlib.NewWriterLevel(&buf, zlib.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer: %w", err)
	}

	// Write data to the compressor
	if _, err := w.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data to zlib compressor: %w", err)
	}

	// Close the writer to flush any pending data
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %w", err)
	}

	return buf.Bytes(), nil
}

package compression

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
)

func (c *ZlibCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data: insufficient bytes for size header")
	}

	compressedSize := binary.BigEndian.Uint32(data[:4])
	if int(compressedSize) > len(data)-4 {
		return nil, fmt.Errorf("invalid compressed data size: expected %d, got %d bytes available", compressedSize, len(data)-4)
	}

	compressedData := data[4 : 4+compressedSize]

	r, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer r.Close()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("failed to read data from zlib reader: %w", err)
	}

	return buffer.Bytes(), nil
}

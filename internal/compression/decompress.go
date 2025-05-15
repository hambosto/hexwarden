package compression

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

const maxDecompressedSize = 1 << 30 // 1GB

func DecompressData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data: insufficient bytes for size header")
	}

	compressedSize := binary.BigEndian.Uint32(data[:4])
	if int(compressedSize) > len(data)-4 {
		return nil, fmt.Errorf("invalid compressed data size: expected %d, got %d bytes available",
			compressedSize, len(data)-4)
	}

	compressedData := data[4 : 4+compressedSize]

	return decompress(compressedData)
}

func decompress(compressedData []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer r.Close()

	limitedReader := io.LimitReader(r, maxDecompressedSize)

	var buf bytes.Buffer
	bytesRead, err := io.Copy(&buf, limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data with zlib: %w", err)
	}

	if bytesRead >= maxDecompressedSize {
		return nil, fmt.Errorf("decompressed data exceeds maximum allowed size of %d bytes", maxDecompressedSize)
	}

	return buf.Bytes(), nil
}

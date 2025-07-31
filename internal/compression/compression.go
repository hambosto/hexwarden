package compression

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// Compression handles zlib compression and decompression using best compression
type Compression struct{}

// New creates a new Compressor
func New() *Compression {
	return &Compression{}
}

// Compress compresses the input data using zlib with best compression
func (c *Compression) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	writer, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses zlib compressed data
func (c *Compression) Decompress(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short: need at least 4 bytes for size header, got %d", len(data))
	}

	// Read the compressed size from the header
	compressedSize := binary.BigEndian.Uint32(data[:4])
	if int(compressedSize) > len(data)-4 {
		return nil, fmt.Errorf("invalid compressed size: header indicates %d bytes, but only %d available", compressedSize, len(data)-4)
	}

	compressedData := data[4 : 4+compressedSize]
	reader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Read all decompressed data
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

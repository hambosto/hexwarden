package compression

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// Compression levels
const (
	LevelNoCompression      = zlib.NoCompression
	LevelBestSpeed          = zlib.BestSpeed
	LevelBestCompression    = zlib.BestCompression
	LevelDefaultCompression = zlib.DefaultCompression
	LevelHuffmanOnly        = zlib.HuffmanOnly
)

// Compression handles zlib compression and decompression
type Compression struct {
	level int
}

// New creates a new Compressor with the specified compression level
func New(level int) (*Compression, error) {
	// Validate compression level
	if level < zlib.HuffmanOnly || level > zlib.BestCompression {
		return nil, fmt.Errorf("invalid compression level: %d (valid range: %d to %d)",
			level, zlib.HuffmanOnly, zlib.BestCompression)
	}

	return &Compression{level: level}, nil
}

// Compress compresses the input data using zlib with the configured compression level
// The output includes a 4-byte header with the compressed data size
func (c *Compression) Compress(data []byte) ([]byte, error) {
	var compressBuf bytes.Buffer

	writer, err := zlib.NewWriterLevel(&compressBuf, c.level)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	compressedData := compressBuf.Bytes()

	// Create final buffer with size header + compressed data
	var finalBuf bytes.Buffer

	// Write the compressed size as a 4-byte big-endian header
	if err := binary.Write(&finalBuf, binary.BigEndian, uint32(len(compressedData))); err != nil {
		return nil, fmt.Errorf("failed to write size header: %w", err)
	}

	// Write the compressed data
	if _, err := finalBuf.Write(compressedData); err != nil {
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	return finalBuf.Bytes(), nil
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

	// Read all decompressed data
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if err := reader.Close(); err != nil {
		return nil, err
	}

	return decompressed, nil
}

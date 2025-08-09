package compression

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/hambosto/hexwarden/internal/constants"
)

// Compressor handles data compression and decompression using gzip
type Compressor struct {
	level int
}

// NewCompressor creates a new compressor with the specified compression level
func NewCompressor(level constants.CompressionLevel) (*Compressor, error) {
	// Validate compression level
	if level < constants.LevelNoCompression || level > constants.LevelBestCompression {
		level = constants.LevelDefaultCompression
	}

	return &Compressor{
		level: int(level),
	}, nil
}

// NewDefaultCompressor creates a new compressor with default compression level
func NewDefaultCompressor() (*Compressor, error) {
	return NewCompressor(constants.LevelDefaultCompression)
}

// Compress compresses the input data using gzip
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, c.level)
	if err != nil {
		return nil, constants.ErrCompressionFailed
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, constants.ErrCompressionFailed
	}

	if err := writer.Close(); err != nil {
		return nil, constants.ErrCompressionFailed
	}

	return buf.Bytes(), nil
}

// Decompress decompresses the input data using gzip
func (c *Compressor) Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, constants.ErrDecompressionFailed
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return nil, constants.ErrDecompressionFailed
	}

	return buf.Bytes(), nil
}

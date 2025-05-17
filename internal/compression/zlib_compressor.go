package compression

import (
	"compress/zlib"
	"fmt"
)

type CompressionLevel int

const (
	LevelNoCompression CompressionLevel = iota
	LevelBestSpeed
	LevelBestCompression
	LevelDefault
	LevelHuffmanOnly
)

func (c CompressionLevel) Valid() bool {
	return c >= LevelNoCompression && c <= LevelHuffmanOnly
}

func (c CompressionLevel) getCompressionLevel() int {
	switch c {
	case LevelNoCompression:
		return zlib.NoCompression
	case LevelBestSpeed:
		return zlib.BestSpeed
	case LevelBestCompression:
		return zlib.BestCompression
	case LevelHuffmanOnly:
		return zlib.HuffmanOnly
	default:
		return zlib.DefaultCompression
	}
}

type ZlibCompressor struct {
	level CompressionLevel
}

func NewZlibCompressor(level CompressionLevel) (*ZlibCompressor, error) {
	if !level.Valid() {
		return nil, fmt.Errorf("invalid compression level: %d", level)
	}
	return &ZlibCompressor{level: level}, nil
}

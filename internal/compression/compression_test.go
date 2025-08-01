package compression

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestCompressionRoundTrip(t *testing.T) {
	comp, err := New(LevelBestCompression)
	if err != nil {
		t.Fatalf("failed to create compressor: %v", err)
	}

	input := []byte("this is some test data that should compress well well well!")

	compressedData, err := comp.Compress(input)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	// No need to manually add header - Compress now includes it
	decompressed, err := comp.Decompress(compressedData)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("decompressed data does not match original\nGot:  %q\nWant: %q", decompressed, input)
	}
}

func TestCompressionLevels(t *testing.T) {
	testData := []byte("this is some test data that should compress well with different levels!")

	levels := []struct {
		name  string
		level int
	}{
		{"BestCompression", LevelBestCompression},
		{"BestSpeed", LevelBestSpeed},
		{"DefaultCompression", LevelDefaultCompression},
		{"NoCompression", LevelNoCompression},
		{"HuffmanOnly", LevelHuffmanOnly},
	}

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			comp, err := New(tc.level)
			if err != nil {
				t.Fatalf("failed to create compressor with level %s: %v", tc.name, err)
			}

			compressed, err := comp.Compress(testData)
			if err != nil {
				t.Fatalf("compression failed with level %s: %v", tc.name, err)
			}

			decompressed, err := comp.Decompress(compressed)
			if err != nil {
				t.Fatalf("decompression failed with level %s: %v", tc.name, err)
			}

			if !bytes.Equal(testData, decompressed) {
				t.Errorf("data mismatch with level %s\nGot:  %q\nWant: %q", tc.name, decompressed, testData)
			}
		})
	}
}

func TestNewWithInvalidLevel(t *testing.T) {
	invalidLevels := []int{-3, 10, 100}

	for _, level := range invalidLevels {
		t.Run("InvalidLevel", func(t *testing.T) {
			_, err := New(level)
			if err == nil {
				t.Fatalf("expected error for invalid compression level %d, got nil", level)
			}
		})
	}
}

func TestDecompressInvalidHeader(t *testing.T) {
	comp, err := New(LevelBestCompression)
	if err != nil {
		t.Fatalf("failed to create compressor: %v", err)
	}

	// Too short
	data := []byte{0x01, 0x02}
	_, err = comp.Decompress(data)
	if err == nil {
		t.Fatal("expected error due to short header, got nil")
	}

	// Invalid compressed size (too large)
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint32(1000)); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	buf.Write([]byte{0x00, 0x01}) // actual data < 1000
	_, err = comp.Decompress(buf.Bytes())
	if err == nil {
		t.Fatal("expected error due to invalid compressed size, got nil")
	}
}

func TestDecompressCorruptedData(t *testing.T) {
	comp, err := New(LevelBestCompression)
	if err != nil {
		t.Fatalf("failed to create compressor: %v", err)
	}

	// Proper header, but invalid compressed data
	corruptData := []byte("this is not valid zlib")
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(corruptData))); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	buf.Write(corruptData)

	_, err = comp.Decompress(buf.Bytes())
	if err == nil {
		t.Fatal("expected error due to corrupted zlib data, got nil")
	}
}

func TestEmptyData(t *testing.T) {
	comp, err := New(LevelBestCompression)
	if err != nil {
		t.Fatalf("failed to create compressor: %v", err)
	}

	// Test empty input
	empty := []byte{}
	compressed, err := comp.Compress(empty)
	if err != nil {
		t.Fatalf("compression of empty data failed: %v", err)
	}

	decompressed, err := comp.Decompress(compressed)
	if err != nil {
		t.Fatalf("decompression of empty data failed: %v", err)
	}

	if !bytes.Equal(empty, decompressed) {
		t.Errorf("decompressed empty data does not match\nGot:  %q\nWant: %q", decompressed, empty)
	}
}

func TestLargeData(t *testing.T) {
	comp, err := New(LevelBestCompression)
	if err != nil {
		t.Fatalf("failed to create compressor: %v", err)
	}

	// Generate large test data
	largeData := make([]byte, 100000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	compressed, err := comp.Compress(largeData)
	if err != nil {
		t.Fatalf("compression of large data failed: %v", err)
	}

	decompressed, err := comp.Decompress(compressed)
	if err != nil {
		t.Fatalf("decompression of large data failed: %v", err)
	}

	if !bytes.Equal(largeData, decompressed) {
		t.Error("decompressed large data does not match original")
	}
}

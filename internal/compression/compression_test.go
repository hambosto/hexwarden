package compression

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestCompressionRoundTrip(t *testing.T) {
	comp := New()
	input := []byte("this is some test data that should compress well well well!")

	compressedData, err := comp.Compress(input)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	// Simulate wrapping compressed data with size header
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(compressedData))); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}
	buf.Write(compressedData)

	decompressed, err := comp.Decompress(buf.Bytes())
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}

	if !bytes.Equal(input, decompressed) {
		t.Errorf("decompressed data does not match original\nGot:  %q\nWant: %q", decompressed, input)
	}
}

func TestDecompressInvalidHeader(t *testing.T) {
	comp := New()

	// Too short
	data := []byte{0x01, 0x02}
	_, err := comp.Decompress(data)
	if err == nil {
		t.Fatal("expected error due to short header, got nil")
	}

	// Invalid compressed size (too large)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(1000)) // bogus size
	buf.Write([]byte{0x00, 0x01})                     // actual data < 1000
	_, err = comp.Decompress(buf.Bytes())
	if err == nil {
		t.Fatal("expected error due to invalid compressed size, got nil")
	}
}

func TestDecompressCorruptedData(t *testing.T) {
	comp := New()

	// Proper header, but invalid compressed data
	corruptData := []byte("this is not valid zlib")
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(len(corruptData)))
	buf.Write(corruptData)

	_, err := comp.Decompress(buf.Bytes())
	if err == nil {
		t.Fatal("expected error due to corrupted zlib data, got nil")
	}
}

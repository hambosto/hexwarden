package padding

import (
	"bytes"
	"errors"
	"testing"
)

func TestPKCS7(t *testing.T) {
	_, err := New(0)
	if !errors.Is(err, ErrInvalidBlockSize) {
		t.Errorf("expected ErrInvalidBlockSize, got %v", err)
	}

	_, err = New(300)
	if !errors.Is(err, ErrInvalidBlockSize) {
		t.Errorf("expected ErrInvalidBlockSize, got %v", err)
	}

	p, err := New(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.BlockSize() != 16 {
		t.Errorf("expected block size 16, got %d", p.BlockSize())
	}
}

func TestPadUnpad(t *testing.T) {
	p, _ := New(16)
	original := []byte("hello world")

	padded, err := p.Pad(original)
	if err != nil {
		t.Fatalf("padding failed: %v", err)
	}

	unpadded, err := p.Unpad(padded)
	if err != nil {
		t.Fatalf("unpadding failed: %v", err)
	}

	if !bytes.Equal(original, unpadded) {
		t.Errorf("unpadded data mismatch\nGot:  %q\nWant: %q", unpadded, original)
	}
}

func TestPadExactBlockSize(t *testing.T) {
	p, _ := New(8)
	original := []byte("1234567") // 7 bytes
	padded, err := p.Pad(original)
	if err != nil {
		t.Fatal(err)
	}

	// The result should be aligned to 8-byte block, including header and padding
	if len(padded)%p.BlockSize() != 0 {
		t.Errorf("expected padded length to be multiple of block size")
	}

	unpadded, err := p.Unpad(padded)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(original, unpadded) {
		t.Errorf("unpadded does not match original: got %q, want %q", unpadded, original)
	}
}

func TestPadTooLarge(t *testing.T) {
	p, _ := New(16)

	// 1 more than uint32 max
	tooLarge := make([]byte, 1<<32)
	_, err := p.Pad(tooLarge)
	if !errors.Is(err, ErrDataTooLarge) {
		t.Errorf("expected ErrDataTooLarge, got %v", err)
	}
}

func TestUnpadInvalidLength(t *testing.T) {
	p, _ := New(8)
	_, err := p.Unpad([]byte("notmod")) // 7 bytes, not divisible by block size
	if !errors.Is(err, ErrInvalidDataLength) {
		t.Errorf("expected ErrInvalidDataLength, got %v", err)
	}
}

func TestUnpadInvalidPadding(t *testing.T) {
	p, _ := New(8)

	data := []byte("12345678abcdefgh") // padded correctly
	padded, _ := p.Pad(data)
	padded[len(padded)-1] = 0 // invalidate padding

	_, err := p.Unpad(padded)
	if !errors.Is(err, ErrInvalidPadding) {
		t.Errorf("expected ErrInvalidPadding, got %v", err)
	}
}

func TestUnpadShortHeader(t *testing.T) {
	p, _ := New(8)

	data := []byte("abcd") // 4 bytes
	padded, _ := p.Pad(data)

	// Corrupt padding length
	padded[len(padded)-1] = 20

	_, err := p.Unpad(padded)
	if !errors.Is(err, ErrInvalidPadding) {
		t.Errorf("expected ErrInvalidPadding, got %v", err)
	}
}

func TestUnpadHeaderMismatch(t *testing.T) {
	p, _ := New(8)

	original := []byte("data with header")
	padded, _ := p.Pad(original)

	// Corrupt header length field
	padded[0] = 0xFF
	padded[1] = 0xFF
	padded[2] = 0xFF
	padded[3] = 0xFF

	_, err := p.Unpad(padded)
	if !errors.Is(err, ErrInvalidPadding) {
		t.Errorf("expected ErrInvalidPadding (header mismatch), got %v", err)
	}
}

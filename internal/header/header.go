// Package header provides functionality for reading and writing encrypted file headers.
package header

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Header field sizes in bytes.
const (
	SaltSize          = 32
	OriginalSizeBytes = 8
	NonceSize         = 12
	VerificationSize  = 32
)

// Common errors.
var (
	ErrInvalidSaltSize         = errors.New("invalid salt size")
	ErrInvalidNonceSize        = errors.New("invalid nonce size")
	ErrInvalidVerificationSize = errors.New("invalid verification hash size")
	ErrShortWrite              = errors.New("short write")
)

// Header represents an encrypted file header containing metadata and verification information.
type Header struct {
	Salt             []byte
	OriginalSize     uint64
	VerificationHash []byte
}

// New creates a new Header with the provided parameters and validates it.
func New(salt []byte, originalSize uint64, key []byte) (*Header, error) {
	verification := getVerificationHash(salt, key)

	h := &Header{
		Salt:             salt,
		OriginalSize:     originalSize,
		VerificationHash: verification,
	}

	if err := h.Validate(); err != nil {
		return nil, err
	}

	return h, nil
}

// Validate checks that all header fields have the correct sizes.
func (h *Header) Validate() error {
	if len(h.Salt) != SaltSize {
		return fmt.Errorf("%w: got %d, want %d", ErrInvalidSaltSize, len(h.Salt), SaltSize)
	}

	if len(h.VerificationHash) != VerificationSize {
		return fmt.Errorf("%w: got %d, want %d", ErrInvalidVerificationSize, len(h.VerificationHash), VerificationSize)
	}

	return nil
}

// VerifyPassword checks if the provided key matches the stored verification hash.
func (h *Header) VerifyPassword(key []byte) bool {
	expectedHash := getVerificationHash(h.Salt, key)
	return hmac.Equal(h.VerificationHash, expectedHash)
}

// Read reads a Header from the provided io.Reader.
func Read(r io.Reader) (*Header, error) {
	var h Header

	salt, err := readExact(r, SaltSize)
	if err != nil {
		return nil, fmt.Errorf("reading salt: %w", err)
	}
	h.Salt = salt

	sizeBuffer, err := readExact(r, OriginalSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("reading original size: %w", err)
	}
	h.OriginalSize = binary.BigEndian.Uint64(sizeBuffer)

	verification, err := readExact(r, VerificationSize)
	if err != nil {
		return nil, fmt.Errorf("reading verification hash: %w", err)
	}
	h.VerificationHash = verification

	if err := h.Validate(); err != nil {
		return nil, err
	}

	return &h, nil
}

// Write writes the Header to the provided io.Writer.
func (h *Header) Write(w io.Writer) error {
	if err := h.Validate(); err != nil {
		return err
	}

	if err := writeAll(w, h.Salt); err != nil {
		return fmt.Errorf("writing salt: %w", err)
	}

	sizeBuffer := make([]byte, OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuffer, h.OriginalSize)
	if err := writeAll(w, sizeBuffer); err != nil {
		return fmt.Errorf("writing original size: %w", err)
	}

	if err := writeAll(w, h.VerificationHash); err != nil {
		return fmt.Errorf("writing verification hash: %w", err)
	}

	return nil
}

// getVerificationHash computes HMAC-SHA256 of salt using the provided key.
func getVerificationHash(salt, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(salt)
	return mac.Sum(nil)
}

// readExact reads exactly n bytes from r.
func readExact(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(r, buf)
	return buf, err
}

// writeAll writes all data to w, ensuring complete writes.
func writeAll(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("%w: wrote %d of %d bytes", ErrShortWrite, n, len(data))
	}
	return nil
}

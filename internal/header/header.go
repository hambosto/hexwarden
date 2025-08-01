package header

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Field sizes in bytes.
const (
	SaltSize         = 32
	OriginalSizeSize = 8
	NonceSize        = 12
	VerificationSize = 32
	TotalSize        = SaltSize + OriginalSizeSize + NonceSize + VerificationSize // 84 bytes
)

// Errors returned by this package.
var (
	ErrInvalidSaltSize         = fmt.Errorf("salt must be %d bytes", SaltSize)
	ErrInvalidNonceSize        = fmt.Errorf("nonce must be %d bytes", NonceSize)
	ErrInvalidVerificationSize = fmt.Errorf("verification hash must be %d bytes", VerificationSize)
	ErrShortWrite              = errors.New("incomplete write operation")
	ErrInvalidPassword         = errors.New("password verification failed")
)

// Header represents an encrypted file header containing cryptographic metadata.
type Header struct {
	salt         [SaltSize]byte
	originalSize uint64
	nonce        [NonceSize]byte
	verification [VerificationSize]byte
}

// Config holds parameters for creating a new header.
type Config struct {
	Salt         []byte
	OriginalSize uint64
	Nonce        []byte
	Key          []byte
}

// New creates and validates a new Header from the provided configuration.
func New(cfg Config) (*Header, error) {
	h := &Header{
		originalSize: cfg.OriginalSize,
	}

	if err := h.setSalt(cfg.Salt); err != nil {
		return nil, err
	}

	if err := h.setNonce(cfg.Nonce); err != nil {
		return nil, err
	}

	h.setVerification(cfg.Key)
	return h, nil
}

// Read reads and validates a Header from the provided reader.
func Read(r io.Reader) (*Header, error) {
	buf := make([]byte, TotalSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	h := &Header{}
	offset := 0

	// Parse salt
	copy(h.salt[:], buf[offset:offset+SaltSize])
	offset += SaltSize

	// Parse original size
	h.originalSize = binary.BigEndian.Uint64(buf[offset : offset+OriginalSizeSize])
	offset += OriginalSizeSize

	// Parse nonce
	copy(h.nonce[:], buf[offset:offset+NonceSize])
	offset += NonceSize

	// Parse verification hash
	copy(h.verification[:], buf[offset:offset+VerificationSize])

	return h, nil
}

// Write writes the header to the provided writer.
func (h *Header) Write(w io.Writer) error {
	buf := make([]byte, 0, TotalSize)

	// Append salt
	buf = append(buf, h.salt[:]...)

	// Append original size
	sizeBytes := make([]byte, OriginalSizeSize)
	binary.BigEndian.PutUint64(sizeBytes, h.originalSize)
	buf = append(buf, sizeBytes...)

	// Append nonce
	buf = append(buf, h.nonce[:]...)

	// Append verification hash
	buf = append(buf, h.verification[:]...)

	n, err := w.Write(buf)
	if err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	if n != TotalSize {
		return fmt.Errorf("%w: wrote %d of %d bytes", ErrShortWrite, n, TotalSize)
	}

	return nil
}

// VerifyPassword checks if the provided key produces the stored verification hash.
func (h *Header) VerifyPassword(key []byte) error {
	expected := h.computeVerification(key)
	if !hmac.Equal(h.verification[:], expected) {
		return ErrInvalidPassword
	}
	return nil
}

// Salt returns a copy of the header's salt.
func (h *Header) Salt() []byte {
	salt := make([]byte, SaltSize)
	copy(salt, h.salt[:])
	return salt
}

// OriginalSize returns the original file size stored in the header.
func (h *Header) OriginalSize() uint64 {
	return h.originalSize
}

// Nonce returns a copy of the header's nonce.
func (h *Header) Nonce() []byte {
	nonce := make([]byte, NonceSize)
	copy(nonce, h.nonce[:])
	return nonce
}

// VerificationHash returns a copy of the header's verification hash.
func (h *Header) VerificationHash() []byte {
	hash := make([]byte, VerificationSize)
	copy(hash, h.verification[:])
	return hash
}

// setSalt validates and sets the salt field.
func (h *Header) setSalt(salt []byte) error {
	if len(salt) != SaltSize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidSaltSize, len(salt))
	}
	copy(h.salt[:], salt)
	return nil
}

// setNonce validates and sets the nonce field.
func (h *Header) setNonce(nonce []byte) error {
	if len(nonce) != NonceSize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidNonceSize, len(nonce))
	}
	copy(h.nonce[:], nonce)
	return nil
}

// setVerification computes and sets the verification hash.
func (h *Header) setVerification(key []byte) {
	verification := h.computeVerification(key)
	copy(h.verification[:], verification)
}

// computeVerification calculates HMAC-SHA256 of the salt using the provided key.
func (h *Header) computeVerification(key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(h.salt[:])
	return mac.Sum(nil)
}

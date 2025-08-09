package crypto

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"

	"github.com/hambosto/hexwarden/internal/constants"
)

// Header represents the metadata prepended to an encrypted file with tamper protection
type Header struct {
	salt          []byte // 32 bytes: cryptographically random salt for KDF
	originalSize  uint64 // 8 bytes: size of original plaintext
	nonce         []byte // 16 bytes: nonce for AEAD
	integrityHash []byte // 32 bytes: SHA-256 hash over [Magic, Salt, Size, Nonce]
	authTag       []byte // 32 bytes: HMAC-SHA256 over [Magic, Salt, Size, Nonce, IntegrityHash] with key
}

// NewHeader creates a new, fully-hardened header
func NewHeader(salt []byte, originalSize uint64, key []byte) (*Header, error) {
	if err := validateHeaderInputs(salt, key); err != nil {
		return nil, err
	}

	// Generate cryptographically random nonce
	nonce := make([]byte, constants.NonceSizeBytes)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	header := &Header{
		salt:         append([]byte(nil), salt...), // Defensive copy
		originalSize: originalSize,
		nonce:        nonce,
	}

	// Compute integrity hash and auth tag with the provided key
	if err := header.computeProtection(key); err != nil {
		return nil, fmt.Errorf("failed to compute protection: %w", err)
	}

	return header, nil
}

// Salt returns a copy of the header's salt
func (h *Header) Salt() []byte {
	return append([]byte(nil), h.salt...)
}

// OriginalSize returns the stored original plaintext size
func (h *Header) OriginalSize() uint64 {
	return h.originalSize
}

// Nonce returns a copy of the header's nonce
func (h *Header) Nonce() []byte {
	return append([]byte(nil), h.nonce...)
}

// VerifyKey validates the key against the header, checking full cryptographic integrity and authentication
func (h *Header) VerifyKey(key []byte) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	expectedAuth := h.computeAuthTag(key)
	if !hmac.Equal(h.authTag, expectedAuth) {
		return constants.ErrAuthFailure
	}

	expectedIntegrity := h.computeIntegrityHash()
	if !bytes.Equal(h.integrityHash, expectedIntegrity) {
		return constants.ErrIntegrityFailure
	}

	return nil
}

// WriteTo writes the serialized header to the given writer
func (h *Header) WriteTo(w io.Writer) (int64, error) {
	if err := h.validate(); err != nil {
		return 0, err
	}

	// Serialize header into a buffer
	buf := make([]byte, 0, constants.TotalHeaderSize)
	buf = h.marshal(buf)

	n, err := w.Write(buf)
	if err != nil {
		return int64(n), fmt.Errorf("write header: %w", err)
	}
	if n != len(buf) {
		return int64(n), fmt.Errorf("%w: wrote %d of %d bytes", constants.ErrIncompleteWrite, n, len(buf))
	}

	return int64(n), nil
}

// Write is a shorthand for WriteTo, discarding the number of bytes written
func (h *Header) Write(w io.Writer) error {
	_, err := h.WriteTo(w)
	return err
}

// ReadHeader reads and parses a header from a reader
func ReadHeader(r io.Reader) (*Header, error) {
	buf := make([]byte, constants.TotalHeaderSize)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", constants.ErrIncompleteRead, err)
	}

	header, err := unmarshalHeader(buf)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// validateHeaderInputs ensures all cryptographic inputs are safe and strong
func validateHeaderInputs(salt []byte, key []byte) error {
	if len(salt) != constants.SaltSizeBytes {
		return fmt.Errorf("%w: got %d bytes", constants.ErrInvalidSalt, len(salt))
	}
	if len(key) == 0 {
		return fmt.Errorf("key cannot be nil or empty")
	}
	if err := ValidateSalt(salt); err != nil {
		return err
	}
	return nil
}

// validate checks the internal consistency of the header fields before serialization
func (h *Header) validate() error {
	if len(h.salt) != constants.SaltSizeBytes {
		return fmt.Errorf("%w: got %d bytes", constants.ErrInvalidSalt, len(h.salt))
	}
	if len(h.nonce) != constants.NonceSizeBytes {
		return fmt.Errorf("%w: got %d bytes", constants.ErrInvalidNonce, len(h.nonce))
	}
	if len(h.integrityHash) != constants.IntegritySize {
		return fmt.Errorf("%w: got %d bytes", constants.ErrInvalidIntegrity, len(h.integrityHash))
	}
	if len(h.authTag) != constants.AuthSize {
		return fmt.Errorf("%w: got %d bytes", constants.ErrInvalidAuth, len(h.authTag))
	}
	return nil
}

// computeProtection calculates both the integrity hash and authentication tag
func (h *Header) computeProtection(key []byte) error {
	h.integrityHash = h.computeIntegrityHash()
	h.authTag = h.computeAuthTag(key)
	return nil
}

// computeIntegrityHash returns the SHA-256 hash of the header's critical fields
func (h *Header) computeIntegrityHash() []byte {
	hasher := sha256.New()
	hasher.Write([]byte(constants.MagicBytes))
	hasher.Write(h.salt)

	sizeBuf := make([]byte, constants.OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuf, h.originalSize)
	hasher.Write(sizeBuf)

	hasher.Write(h.nonce)

	return hasher.Sum(nil)
}

// computeAuthTag computes the HMAC-SHA256 authentication tag over the header fields and integrity hash
func (h *Header) computeAuthTag(key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(constants.MagicBytes))
	mac.Write(h.salt)

	sizeBuf := make([]byte, constants.OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuf, h.originalSize)
	mac.Write(sizeBuf)

	mac.Write(h.nonce)
	mac.Write(h.integrityHash)

	return mac.Sum(nil)
}

// marshal serializes the header fields in order into the given buffer
func (h *Header) marshal(buf []byte) []byte {
	buf = append(buf, constants.MagicBytes...)
	buf = append(buf, h.salt...)

	sizeBuf := make([]byte, constants.OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuf, h.originalSize)
	buf = append(buf, sizeBuf...)

	buf = append(buf, h.nonce...)
	buf = append(buf, h.integrityHash...)
	buf = append(buf, h.authTag...)

	// Compute CRC32 checksum of everything except magic bytes
	checksum := crc32.ChecksumIEEE(buf[len(constants.MagicBytes):])
	checksumBuf := make([]byte, constants.ChecksumSize)
	binary.BigEndian.PutUint32(checksumBuf, checksum)
	buf = append(buf, checksumBuf...)

	return buf
}

// unmarshalHeader deserializes and validates a header from the given byte slice
func unmarshalHeader(data []byte) (*Header, error) {
	if len(data) != constants.TotalHeaderSize {
		return nil, fmt.Errorf("invalid header size: got %d, expected %d", len(data), constants.TotalHeaderSize)
	}

	// Check magic bytes
	if subtle.ConstantTimeCompare(data[:len(constants.MagicBytes)], []byte(constants.MagicBytes)) != 1 {
		return nil, constants.ErrInvalidMagic
	}

	offset := len(constants.MagicBytes)

	// Parse salt
	salt := make([]byte, constants.SaltSizeBytes)
	copy(salt, data[offset:offset+constants.SaltSizeBytes])
	offset += constants.SaltSizeBytes

	// Parse original size
	originalSize := binary.BigEndian.Uint64(data[offset : offset+constants.OriginalSizeBytes])
	offset += constants.OriginalSizeBytes

	// Parse nonce
	nonce := make([]byte, constants.NonceSizeBytes)
	copy(nonce, data[offset:offset+constants.NonceSizeBytes])
	offset += constants.NonceSizeBytes

	// Parse integrity hash
	integrityHash := make([]byte, constants.IntegritySize)
	copy(integrityHash, data[offset:offset+constants.IntegritySize])
	offset += constants.IntegritySize

	// Parse auth tag
	authTag := make([]byte, constants.AuthSize)
	copy(authTag, data[offset:offset+constants.AuthSize])
	offset += constants.AuthSize

	// Validate CRC32 checksum
	checksumData := data[len(constants.MagicBytes):offset]
	calculatedChecksum := crc32.ChecksumIEEE(checksumData)

	storedBytes := data[offset : offset+constants.ChecksumSize]
	calculatedBytes := make([]byte, constants.ChecksumSize)
	binary.BigEndian.PutUint32(calculatedBytes, calculatedChecksum)

	if subtle.ConstantTimeCompare(storedBytes, calculatedBytes) != 1 {
		return nil, constants.ErrChecksumMismatch
	}

	header := &Header{
		salt:          salt,
		originalSize:  originalSize,
		nonce:         nonce,
		integrityHash: integrityHash,
		authTag:       authTag,
	}

	// Verify integrity hash for tamper detection
	if !bytes.Equal(header.integrityHash, header.computeIntegrityHash()) {
		return nil, constants.ErrTampering
	}

	return header, nil
}

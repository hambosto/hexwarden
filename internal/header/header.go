// Package header provides functionality for reading and writing encrypted file headers
// with comprehensive tampering protection and cryptographic integrity verification.
//
// The Header struct encapsulates metadata critical for secure file decryption, including:
// - Cryptographically random salt (for key derivation)
// - Original plaintext size (for verification and padding removal)
// - Nonce (for AEAD ciphers)
// - Integrity hash (SHA-256 over critical fields, for tamper detection)
// - Authentication tag (HMAC-SHA256, keyed, for key correctness & tampering)
// - CRC32 checksum (for accidental corruption/tampering detection)
// The header is fixed at 128 bytes and must be written & parsed atomically.
package header

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
)

// Header format constants: these define the header layout and field sizes.
const (
	MagicBytes        = "HWX2" // Identifies file type and version
	SaltSize          = 32     // Salt for KDF, must be securely random
	OriginalSizeBytes = 8      // Size of original plaintext, uint64
	NonceSize         = 16     // Nonce for AEAD encryption
	IntegritySize     = 32     // SHA-256 hash of header fields (except auth & checksum)
	AuthSize          = 32     // HMAC-SHA256 authentication tag
	ChecksumSize      = 4      // CRC32 checksum of header (except magic)
	TotalSize         = 128    // the fixed byte size of the header.
)

// Header errors with detailed context for all failure cases.
var (
	ErrInvalidMagic     = fmt.Errorf("invalid magic bytes, expected %q", MagicBytes)
	ErrInvalidSalt      = fmt.Errorf("invalid salt size, expected %d bytes", SaltSize)
	ErrInvalidNonce     = fmt.Errorf("invalid nonce size, expected %d bytes", NonceSize)
	ErrInvalidIntegrity = fmt.Errorf("invalid integrity hash size, expected %d bytes", IntegritySize)
	ErrInvalidAuth      = fmt.Errorf("invalid authentication tag size, expected %d bytes", AuthSize)
	ErrChecksumMismatch = fmt.Errorf("header checksum verification failed - possible tampering")
	ErrIntegrityFailure = fmt.Errorf("header integrity verification failed - data corrupted")
	ErrAuthFailure      = fmt.Errorf("header authentication failed - invalid key or tampering")
	ErrIncompleteWrite  = fmt.Errorf("incomplete header write")
	ErrIncompleteRead   = fmt.Errorf("incomplete header read")
	ErrTampering        = fmt.Errorf("header tampering detected")
)

// Header represents the metadata prepended to an encrypted file with tamper protection.
// All fields except authTag and integrityHash are public metadata; integrityHash and authTag are cryptographic protections.
type Header struct {
	salt          []byte // 32 bytes: cryptographically random salt for KDF
	originalSize  uint64 // 8 bytes: size of original plaintext
	nonce         []byte // 16 bytes: nonce for AEAD
	integrityHash []byte // 32 bytes: SHA-256 hash over [Magic, Salt, Size, Nonce]
	authTag       []byte // 32 bytes: HMAC-SHA256 over [Magic, Salt, Size, Nonce, IntegrityHash] with key
}

// New creates a new, fully-hardened header.
// It generates a random nonce and computes the integrity hash and authentication tag.
// The salt must be generated securely by the caller.
func New(salt []byte, originalSize uint64, key []byte) (*Header, error) {
	if err := validateInputs(salt, key); err != nil {
		return nil, err
	}

	// Generate cryptographically random nonce.
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	header := &Header{
		salt:         append([]byte(nil), salt...), // Defensive copy
		originalSize: originalSize,
		nonce:        nonce,
	}

	// Compute integrity hash and auth tag with the provided key.
	if err := header.computeProtection(key); err != nil {
		return nil, fmt.Errorf("failed to compute protection: %w", err)
	}

	return header, nil
}

// Salt returns a copy of the header's salt.
func (h *Header) Salt() []byte {
	return append([]byte(nil), h.salt...)
}

// OriginalSize returns the stored original plaintext size.
func (h *Header) OriginalSize() uint64 {
	return h.originalSize
}

// Nonce returns a copy of the header's nonce.
func (h *Header) Nonce() []byte {
	return append([]byte(nil), h.nonce...)
}

// VerifyKey validates the key against the header, checking full cryptographic integrity and authentication.
// Returns ErrAuthFailure if the key is wrong or tampering occurred.
func (h *Header) VerifyKey(key []byte) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}

	expectedAuth := h.computeAuthTag(key)
	if !hmac.Equal(h.authTag, expectedAuth) {
		return ErrAuthFailure
	}

	expectedIntegrity := h.computeIntegrityHash()
	if !bytes.Equal(h.integrityHash, expectedIntegrity) {
		return ErrIntegrityFailure
	}

	return nil
}

// WriteTo writes the serialized header to the given writer, returning the number of bytes written.
// Returns ErrIncompleteWrite on short writes.
func (h *Header) WriteTo(w io.Writer) (int64, error) {
	if err := h.validate(); err != nil {
		return 0, err
	}

	// Serialize header into a buffer.
	buf := make([]byte, 0, TotalSize)
	buf = h.marshal(buf)

	n, err := w.Write(buf)
	if err != nil {
		return int64(n), fmt.Errorf("write header: %w", err)
	}
	if n != len(buf) {
		return int64(n), fmt.Errorf("%w: wrote %d of %d bytes", ErrIncompleteWrite, n, len(buf))
	}

	return int64(n), nil
}

// Write is a shorthand for WriteTo, discarding the number of bytes written.
func (h *Header) Write(w io.Writer) error {
	_, err := h.WriteTo(w)
	return err
}

// ReadFrom reads and parses a header from a reader, returning the header and the number of bytes read.
// Returns ErrIncompleteRead on short reads or if parsing fails.
func ReadFrom(r io.Reader) (*Header, int64, error) {
	buf := make([]byte, TotalSize)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, int64(n), fmt.Errorf("%w: %v", ErrIncompleteRead, err)
	}

	header, err := unmarshal(buf)
	if err != nil {
		return nil, int64(n), err
	}

	return header, int64(n), nil
}

// Read is a shorthand for ReadFrom, discarding the number of bytes read.
func Read(r io.Reader) (*Header, error) {
	header, _, err := ReadFrom(r)
	return header, err
}

// validateInputs ensures all cryptographic inputs are safe and strong.
// It checks salt size, key presence, and salt randomness.
func validateInputs(salt []byte, key []byte) error {
	if len(salt) != SaltSize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidSalt, len(salt))
	}
	if len(key) == 0 {
		return fmt.Errorf("key cannot be nil or empty")
	}
	if isWeakSalt(salt) {
		return fmt.Errorf("weak salt detected - use cryptographically random salt")
	}
	return nil
}

// validate checks the internal consistency of the header fields before serialization.
func (h *Header) validate() error {
	if len(h.salt) != SaltSize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidSalt, len(h.salt))
	}
	if len(h.nonce) != NonceSize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidNonce, len(h.nonce))
	}
	if len(h.integrityHash) != IntegritySize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidIntegrity, len(h.integrityHash))
	}
	if len(h.authTag) != AuthSize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidAuth, len(h.authTag))
	}
	return nil
}

// computeProtection calculates both the integrity hash and authentication tag.
// This must be called after setting salt, originalSize, and nonce.
func (h *Header) computeProtection(key []byte) error {
	h.integrityHash = h.computeIntegrityHash()
	h.authTag = h.computeAuthTag(key)
	return nil
}

// computeIntegrityHash returns the SHA-256 hash of the header's critical fields.
// Order: MagicBytes, Salt, OriginalSize, Nonce.
func (h *Header) computeIntegrityHash() []byte {
	hasher := sha256.New()
	hasher.Write([]byte(MagicBytes))
	hasher.Write(h.salt)

	sizeBuf := make([]byte, OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuf, h.originalSize)
	hasher.Write(sizeBuf)

	hasher.Write(h.nonce)

	return hasher.Sum(nil)
}

// computeAuthTag computes the HMAC-SHA256 authentication tag over the header fields and integrity hash.
func (h *Header) computeAuthTag(key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(MagicBytes))
	mac.Write(h.salt)

	sizeBuf := make([]byte, OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuf, h.originalSize)
	mac.Write(sizeBuf)

	mac.Write(h.nonce)
	mac.Write(h.integrityHash)

	return mac.Sum(nil)
}

// marshal serializes the header fields in order into the given buffer, returning the result.
//
// Layout:
//
//	MagicBytes | salt | originalSize | nonce | integrityHash | authTag | checksum
func (h *Header) marshal(buf []byte) []byte {
	buf = append(buf, MagicBytes...)
	buf = append(buf, h.salt...)

	sizeBuf := make([]byte, OriginalSizeBytes)
	binary.BigEndian.PutUint64(sizeBuf, h.originalSize)
	buf = append(buf, sizeBuf...)

	buf = append(buf, h.nonce...)
	buf = append(buf, h.integrityHash...)
	buf = append(buf, h.authTag...)

	// Compute CRC32 checksum of everything except magic bytes:
	checksum := crc32.ChecksumIEEE(buf[len(MagicBytes):])
	checksumBuf := make([]byte, ChecksumSize)
	binary.BigEndian.PutUint32(checksumBuf, checksum)
	buf = append(buf, checksumBuf...)

	return buf
}

// unmarshal deserializes and validates a header from the given byte slice.
// Checks magic bytes, field sizes, CRC32 checksum, and integrity hash.
func unmarshal(data []byte) (*Header, error) {
	if len(data) != TotalSize {
		return nil, fmt.Errorf("invalid header size: got %d, expected %d", len(data), TotalSize)
	}
	// Check magic bytes
	if subtle.ConstantTimeCompare(data[:len(MagicBytes)], []byte(MagicBytes)) != 1 {
		return nil, ErrInvalidMagic
	}

	offset := len(MagicBytes)

	// Parse salt (32 bytes)
	salt := make([]byte, SaltSize)
	copy(salt, data[offset:offset+SaltSize])
	offset += SaltSize

	// Parse original size (8 bytes)
	originalSize := binary.BigEndian.Uint64(data[offset : offset+OriginalSizeBytes])
	offset += OriginalSizeBytes

	// Parse nonce (16 bytes)
	nonce := make([]byte, NonceSize)
	copy(nonce, data[offset:offset+NonceSize])
	offset += NonceSize

	// Parse integrity hash (32 bytes)
	integrityHash := make([]byte, IntegritySize)
	copy(integrityHash, data[offset:offset+IntegritySize])
	offset += IntegritySize

	// Parse auth tag (32 bytes)
	authTag := make([]byte, AuthSize)
	copy(authTag, data[offset:offset+AuthSize])
	offset += AuthSize

	// Validate CRC32 checksum (4 bytes)
	checksumData := data[len(MagicBytes):offset]
	calculatedChecksum := crc32.ChecksumIEEE(checksumData)

	storedBytes := data[offset : offset+ChecksumSize]
	calculatedBytes := make([]byte, ChecksumSize)
	binary.BigEndian.PutUint32(calculatedBytes, calculatedChecksum)

	if subtle.ConstantTimeCompare(storedBytes, calculatedBytes) != 1 {
		return nil, ErrChecksumMismatch
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
		return nil, ErrTampering
	}

	return header, nil
}

// isWeakSalt checks for weak salt patterns that may indicate non-random sources.
// (e.g., all zero bytes or repeating byte patterns)
func isWeakSalt(salt []byte) bool {
	if len(salt) != SaltSize {
		return true
	}
	allZero := true
	for _, b := range salt {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return true
	}
	// Check for repeating 4-byte patterns (very weak!)
	if len(salt) >= 4 {
		pattern := salt[:4]
		for i := 4; i < len(salt); i += 4 {
			end := min(i+4, len(salt))
			if !bytes.Equal(pattern, salt[i:end]) {
				return false
			}
		}
		return true
	}
	return false
}

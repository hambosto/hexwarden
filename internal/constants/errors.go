package constants

import "errors"

// Infrastructure Layer Errors
var (
	ErrInvalidKey          = errors.New("invalid encryption key")
	ErrInvalidKeySize      = errors.New("AES key must be 16, 24, or 32 bytes")
	ErrEmptyPlaintext      = errors.New("plaintext cannot be empty")
	ErrEmptyCiphertext     = errors.New("ciphertext cannot be empty")
	ErrEncryptionFailed    = errors.New("encryption operation failed")
	ErrDecryptionFailed    = errors.New("decryption operation failed")
	ErrCompressionFailed   = errors.New("compression operation failed")
	ErrDecompressionFailed = errors.New("decompression operation failed")
	ErrEncodingFailed      = errors.New("encoding operation failed")
	ErrDecodingFailed      = errors.New("decoding operation failed")
	ErrPaddingFailed       = errors.New("padding operation failed")
	ErrUnpaddingFailed     = errors.New("unpadding operation failed")
)

// KDF Errors
var (
	ErrEmptyPassword  = errors.New("password cannot be empty")
	ErrInvalidSalt    = errors.New("invalid salt length")
	ErrSaltGeneration = errors.New("failed to generate salt")
)

// Header Errors
var (
	ErrInvalidMagic     = errors.New("invalid magic bytes")
	ErrInvalidHeader    = errors.New("invalid header format")
	ErrInvalidNonce     = errors.New("invalid nonce size")
	ErrInvalidIntegrity = errors.New("invalid integrity hash size")
	ErrInvalidAuth      = errors.New("invalid authentication tag size")
	ErrChecksumMismatch = errors.New("header checksum verification failed")
	ErrIntegrityFailure = errors.New("header integrity verification failed")
	ErrAuthFailure      = errors.New("header authentication failed")
	ErrIncompleteWrite  = errors.New("incomplete header write")
	ErrIncompleteRead   = errors.New("incomplete header read")
	ErrTampering        = errors.New("header tampering detected")
)

// Data Layer Errors
var (
	ErrFileNotFound       = errors.New("file not found")
	ErrFileExists         = errors.New("file already exists")
	ErrFileEmpty          = errors.New("file is empty")
	ErrInvalidPath        = errors.New("invalid file path")
	ErrFileCreateFailed   = errors.New("failed to create file")
	ErrFileOpenFailed     = errors.New("failed to open file")
	ErrFileReadFailed     = errors.New("failed to read file")
	ErrFileWriteFailed    = errors.New("failed to write file")
	ErrSecureDeleteFailed = errors.New("secure deletion failed")
)

// Stream Processing Errors
var (
	ErrNilStream     = errors.New("input and output streams must not be nil")
	ErrCanceled      = errors.New("operation was canceled")
	ErrChunkTooLarge = errors.New("chunk size exceeds maximum allowed")
)

// Business Layer Errors
var (
	ErrPasswordMismatch = errors.New("passwords do not match")
)

// Presentation Layer Errors
var (
	ErrUserCanceled     = errors.New("operation canceled by user")
	ErrNoFilesAvailable = errors.New("no files available for selection")
	ErrPromptFailed     = errors.New("user prompt failed")
)

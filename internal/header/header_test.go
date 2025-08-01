package header

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"
)

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("failed to generate random bytes: %v", err)
	}
	return b
}

func TestNewHeader_Success(t *testing.T) {
	salt := randomBytes(t, SaltSize)
	nonce := randomBytes(t, NonceSize)
	key := randomBytes(t, 32)
	size := uint64(123456)

	cfg := Config{
		Salt:         salt,
		OriginalSize: size,
		Nonce:        nonce,
		Key:          key,
	}

	hdr, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := hdr.OriginalSize(); got != size {
		t.Errorf("original size mismatch: got %d, want %d", got, size)
	}

	if !bytes.Equal(hdr.Salt(), salt) {
		t.Errorf("salt mismatch")
	}

	if !bytes.Equal(hdr.Nonce(), nonce) {
		t.Errorf("nonce mismatch")
	}

	if err := hdr.VerifyPassword(key); err != nil {
		t.Errorf("VerifyPassword failed: %v", err)
	}
}

func TestNewHeader_InvalidSalt(t *testing.T) {
	cfg := Config{
		Salt:         []byte("short"),
		OriginalSize: 0,
		Nonce:        randomBytes(t, NonceSize),
		Key:          randomBytes(t, 32),
	}

	_, err := New(cfg)
	if err == nil || !errors.Is(err, ErrInvalidSaltSize) {
		t.Errorf("expected salt size error, got %v", err)
	}
}

func TestNewHeader_InvalidNonce(t *testing.T) {
	cfg := Config{
		Salt:         randomBytes(t, SaltSize),
		OriginalSize: 0,
		Nonce:        []byte("short"),
		Key:          randomBytes(t, 32),
	}

	_, err := New(cfg)
	if err == nil || !errors.Is(err, ErrInvalidNonceSize) {
		t.Errorf("expected nonce size error, got %v", err)
	}
}

func TestHeader_WriteRead(t *testing.T) {
	salt := randomBytes(t, SaltSize)
	nonce := randomBytes(t, NonceSize)
	key := randomBytes(t, 32)
	size := uint64(987654321)

	cfg := Config{
		Salt:         salt,
		OriginalSize: size,
		Nonce:        nonce,
		Key:          key,
	}

	hdr, err := New(cfg)
	if err != nil {
		t.Fatalf("New header failed: %v", err)
	}

	buf := new(bytes.Buffer)
	if err := hdr.Write(buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	readHdr, err := Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if got := readHdr.OriginalSize(); got != size {
		t.Errorf("original size mismatch: got %d, want %d", got, size)
	}

	if !bytes.Equal(readHdr.Salt(), salt) {
		t.Errorf("salt mismatch after read")
	}

	if !bytes.Equal(readHdr.Nonce(), nonce) {
		t.Errorf("nonce mismatch after read")
	}

	if !bytes.Equal(readHdr.VerificationHash(), hdr.VerificationHash()) {
		t.Errorf("verification hash mismatch after read")
	}

	if err := readHdr.VerifyPassword(key); err != nil {
		t.Errorf("VerifyPassword failed after read: %v", err)
	}
}

func TestVerifyPassword_Failure(t *testing.T) {
	salt := randomBytes(t, SaltSize)
	nonce := randomBytes(t, NonceSize)
	correctKey := randomBytes(t, 32)
	wrongKey := randomBytes(t, 32)

	cfg := Config{
		Salt:         salt,
		OriginalSize: 0,
		Nonce:        nonce,
		Key:          correctKey,
	}

	hdr, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := hdr.VerifyPassword(wrongKey); !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestHeader_Write_ShortWrite(t *testing.T) {
	hdr := &Header{}
	w := &limitedWriter{max: TotalSize - 1}
	err := hdr.Write(w)
	if err == nil || !errors.Is(err, ErrShortWrite) {
		t.Errorf("expected short write error, got: %v", err)
	}
}

// Helper to simulate a short write.
type limitedWriter struct {
	buf bytes.Buffer
	max int
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > lw.max {
		p = p[:lw.max]
	}
	return lw.buf.Write(p)
}

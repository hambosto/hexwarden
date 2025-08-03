package cipher

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes AES-128
	c, err := New(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	plaintext := []byte("secret message")
	ciphertext, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if bytes.Equal(plaintext, ciphertext) {
		t.Errorf("ciphertext should not be equal to plaintext")
	}

	decrypted, err := c.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted text does not match original\nGot:  %q\nWant: %q", decrypted, plaintext)
	}
}

func TestEncryptEmpty(t *testing.T) {
	key := make([]byte, 32)
	c, _ := New(key)

	_, err := c.Encrypt([]byte{})
	if err != ErrEmptyPlaintext {
		t.Errorf("expected ErrEmptyPlaintext, got %v", err)
	}
}

func TestDecryptEmpty(t *testing.T) {
	key := make([]byte, 32)
	c, _ := New(key)

	_, err := c.Decrypt([]byte{})
	if err != ErrEmptyCiphertext {
		t.Errorf("expected ErrEmptyCiphertext, got %v", err)
	}
}

func TestInvalidKeySize(t *testing.T) {
	_, err := New([]byte("shortkey"))
	if err != ErrInvalidKeySize {
		t.Errorf("expected ErrInvalidKeySize, got %v", err)
	}
}

func TestDecryptTamperedData(t *testing.T) {
	key := make([]byte, 16)
	c, _ := New(key)

	plaintext := []byte("secure")
	ciphertext, _ := c.Encrypt(plaintext)

	// Tamper with ciphertext
	ciphertext[len(ciphertext)-1] ^= 0xFF

	_, err := c.Decrypt(ciphertext)
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed due to tampering, got %v", err)
	}
}

func TestDecryptShortCiphertext(t *testing.T) {
	key := make([]byte, 16)
	c, _ := New(key)

	// Too short for nonce
	_, err := c.Decrypt([]byte("123"))
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed for short ciphertext, got %v", err)
	}
}

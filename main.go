package main

import (
	"fmt"
	"log"

	"github.com/hambosto/hexwarden/internal/processor"
)

func main() {
	// Original plaintext data
	originalData := []byte("Hello, this is a secret message that will be compressed, padded, encrypted, and encoded!")

	// Example 64-byte key (in a real scenario, this should be securely generated)
	key := make([]byte, 64)
	for i := range key {
		key[i] = byte(i)
	}

	// Create a new Processor
	proc, err := processor.New(key)
	if err != nil {
		log.Fatalf("Failed to initialize processor: %v", err)
	}

	// Encrypt the data
	encrypted, err := proc.Encrypt(originalData)
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}

	// Decrypt the data
	decrypted, err := proc.Decrypt(encrypted)
	if err != nil {
		log.Fatalf("Decryption failed: %v", err)
	}

	// Print results
	fmt.Println("Original Data : ", string(originalData))
	fmt.Println("Encrypted Data: ", string(encrypted))
	fmt.Println("Decrypted Data: ", string(decrypted))
}

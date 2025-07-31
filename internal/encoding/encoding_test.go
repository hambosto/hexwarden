package encoding

import (
	"bytes"
	"errors"
	"testing"
)

func TestNewEncoder(t *testing.T) {
	_, err := NewEncoder(0, 2)
	if !errors.Is(err, ErrInvalidDataShards) {
		t.Errorf("expected ErrInvalidDataShards, got %v", err)
	}

	_, err = NewEncoder(4, 0)
	if !errors.Is(err, ErrInvalidParityShards) {
		t.Errorf("expected ErrInvalidParityShards, got %v", err)
	}

	encoder, err := NewEncoder(4, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if encoder.dataShards != 4 || encoder.parityShards != 2 {
		t.Errorf("unexpected shard config: %+v", encoder)
	}
}

func TestEncodeDecode(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog")

	encoder, err := NewEncoder(5, 3)
	if err != nil {
		t.Fatalf("failed to create encoder: %v", err)
	}

	encoded, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("failed to encode data: %v", err)
	}

	decoded, err := encoder.Decode(encoded)
	if err != nil {
		t.Fatalf("failed to decode data: %v", err)
	}

	// Trim to original length since padding might be added
	decoded = decoded[:len(data)]

	if !bytes.Equal(data, decoded) {
		t.Errorf("decoded data does not match original\nGot:  %q\nWant: %q", decoded, data)
	}
}

func TestEncodeInvalidSize(t *testing.T) {
	encoder, err := NewEncoder(2, 2)
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, maxdataShards+1) // too big
	_, err = encoder.Encode(data)
	if !errors.Is(err, ErrDataSize) {
		t.Errorf("expected ErrDataSize, got %v", err)
	}
}

func TestDecodeCorruptData(t *testing.T) {
	encoder, err := NewEncoder(4, 2)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("hello world")
	encoded, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("encoding failed: %v", err)
	}

	shards := encoder.splitEncodedData(encoded)

	// Corrupt 3 data shards (exceeding parityShards = 2)
	shards[0] = nil
	shards[1] = nil
	shards[2] = nil

	corrupted := encoder.combineShards(shards)
	_, err = encoder.Decode(corrupted)

	if err == nil {
		t.Fatal("expected decode to fail due to excessive corruption, but got nil error")
	}
}

func TestDecodeInsufficientShards(t *testing.T) {
	encoder, _ := NewEncoder(4, 2)

	data := []byte("1234567890")
	encoded, _ := encoder.Encode(data)

	// Simulate corruption by zeroing out too many shards
	shards := encoder.splitEncodedData(encoded)
	shards[0] = nil
	shards[1] = nil
	shards[2] = nil

	flat := encoder.combineShards(shards)
	_, err := encoder.Decode(flat)
	if err == nil {
		t.Errorf("expected decoding to fail due to corrupted shards")
	}
}

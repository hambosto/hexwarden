# HexWarden Test Suite

This directory contains comprehensive unit tests for the HexWarden file encryption/decryption tool.

## Structure

The test structure mirrors the internal package structure:

- `crypto/` - Tests for cryptographic operations (AES, KDF, header)
- `compression/` - Tests for compression functionality
- `encoding/` - Tests for encoding operations
- `utils/` - Tests for utility functions (padding, helpers)
- `business/` - Tests for business operations (encryptor, decryptor)
- `data/` - Tests for data layer (files, streaming)
- `presentation/` - Tests for presentation layer components
- `helpers/` - Test helper utilities and mocks
- `testdata/` - Test data files and fixtures

## Running Tests

```bash
# Run all tests
go test ./tests/...

# Run tests with coverage
go test -cover ./tests/...

# Run tests with verbose output
go test -v ./tests/...

# Run specific test package
go test ./tests/crypto/...
```

## Test Conventions

- Test files follow the `*_test.go` naming convention
- Test functions start with `Test`
- Benchmark functions start with `Benchmark`
- Example functions start with `Example`
- Use table-driven tests for multiple test cases
- Include both positive and negative test cases
- Test error conditions and edge cases
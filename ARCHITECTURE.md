# Hexwarden v1.1 - Refactored Architecture

## Overview

Hexwarden has been completely refactored with a clean layered architecture that eliminates configuration files, removes dependency injection complexity, and maintains all performance features while providing better maintainability and clarity.

## Architecture Principles

### ✅ **No External Configuration**
- All configuration embedded as constants in code
- No JSON, YAML, ENV files, or external config dependencies
- Compile-time configuration ensures consistency

### ✅ **Layered Architecture**
- Clear separation of concerns across 4 distinct layers
- Each layer has single responsibility and well-defined boundaries
- Dependencies flow downward only (no circular dependencies)

### ✅ **Simplified Dependencies**
- No dependency injection framework
- Direct imports and function calls
- Easier to understand and debug

### ✅ **Maintained Performance**
- All concurrent processing features preserved
- Reed-Solomon error correction maintained
- Streaming and progress tracking intact
- Worker pools and buffering optimizations kept

## Layer Structure

```
┌─────────────────────────────────────────┐
│           Presentation Layer            │
│  ┌─────────────────┐ ┌─────────────────┐│
│  │   CLI Interface │ │  UI Components  ││
│  │                 │ │  - Prompts      ││
│  │                 │ │  - Progress     ││
│  │                 │ │  - Terminal     ││
│  └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│           Business Logic Layer          │
│  ┌─────────────────┐ ┌─────────────────┐│
│  │   Operations    │ │   Workflows     ││
│  │  - Encryptor    │ │  - Validation   ││
│  │  - Decryptor    │ │  - Security     ││
│  └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│            Data Access Layer            │
│  ┌─────────────────┐ ┌─────────────────┐│
│  │  File Operations│ │   Streaming     ││
│  │  - Manager      │ │  - Processor    ││
│  │  - Finder       │ │  - Pool         ││
│  │                 │ │  - Buffer       ││
│  └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│          Infrastructure Layer           │
│  ┌─────────┐ ┌─────────┐ ┌─────────────┐│
│  │ Crypto  │ │Compress │ │  Encoding   ││
│  │- AES    │ │- gzip   │ │- Reed-Sol   ││
│  │- KDF    │ │         │ │             ││
│  │- Header │ │         │ │             ││
│  └─────────┘ └─────────┘ └─────────────┘│
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│              Constants                  │
│  - Configuration Values                 │
│  - Error Definitions                    │
│  - Type Definitions                     │
└─────────────────────────────────────────┘
```

## Directory Structure

```
hexwarden/
├── cmd/hexwarden/main.go                    # Simple entry point
├── internal/
│   ├── constants/                           # All configuration
│   │   ├── config.go                        # App & processing config
│   │   ├── errors.go                        # Error definitions
│   │   └── types.go                         # Type definitions
│   ├── infrastructure/                      # Low-level components
│   │   ├── crypto/
│   │   │   ├── aes.go                       # AES-256-GCM encryption
│   │   │   ├── kdf.go                       # Argon2id key derivation
│   │   │   └── header.go                    # File header management
│   │   ├── compression/
│   │   │   └── compressor.go                # gzip compression
│   │   ├── encoding/
│   │   │   └── encoder.go                   # Reed-Solomon encoding
│   │   ├── utils/
│   │   │   └── padding.go                   # PKCS7 padding
│   │   └── processor.go                     # Combined processor
│   ├── data/                                # Data access
│   │   ├── files/
│   │   │   ├── manager.go                   # File operations
│   │   │   └── finder.go                    # File discovery
│   │   └── streaming/
│   │       ├── processor.go                 # Stream processing
│   │       ├── pool.go                      # Worker pools
│   │       └── buffer.go                    # Result buffering
│   ├── business/                            # Business logic
│   │   └── operations/
│   │       ├── encryptor.go                 # Encryption workflows
│   │       └── decryptor.go                 # Decryption workflows
│   └── presentation/                        # User interface
│       ├── cli/
│       │   └── interface.go                 # Main CLI interface
│       └── ui/
│           ├── prompts.go                   # User prompts
│           ├── progress.go                  # Progress tracking
│           └── terminal.go                  # Terminal operations
```

## Key Features Maintained

### 🔒 **Security Features**
- **AES-256-GCM**: Authenticated encryption with integrity protection
- **Argon2id**: Modern password-based key derivation
- **Reed-Solomon**: Error correction codes for data integrity
- **Secure Random**: Cryptographically secure nonce and salt generation
- **Secure Deletion**: Multiple-pass overwrite for sensitive data

### ⚡ **Performance Features**
- **Concurrent Processing**: Multi-threaded encryption/decryption
- **Streaming**: Memory-efficient processing of large files
- **Worker Pools**: Configurable concurrency levels
- **Progress Tracking**: Real-time progress with ETA
- **Buffering**: Ordered result processing with minimal memory usage

### 🛡️ **Reliability Features**
- **Error Correction**: Reed-Solomon codes can recover from data corruption
- **Header Integrity**: Multiple layers of tamper detection
- **Input Validation**: Comprehensive validation at all layers
- **Graceful Degradation**: Robust error handling and recovery

## Configuration

All configuration is now embedded in [`internal/constants/config.go`](internal/constants/config.go):

```go
// Processing Configuration
const (
    DefaultChunkSize = 1 * 1024 * 1024 // 1MB chunks
    MaxConcurrency   = 8               // Max worker threads
    QueueSize        = 100             // Task queue buffer size
    OverwritePasses  = 3               // Secure deletion passes
)

// Reed-Solomon Configuration
const (
    DataShards   = 4  // Number of data shards
    ParityShards = 10 // Number of parity shards
)

// Security Configuration
const (
    SaltSize   = 32  // Argon2id salt size
    KeySize    = 32  // AES-256 key size
    NonceSize  = 16  // AES-GCM nonce size
)
```

## Usage

The refactored application maintains the same user interface:

```bash
# Build the application
go build -o hexwarden cmd/hexwarden/main.go

# Run the application
./hexwarden
```

## Benefits of New Architecture

### 🎯 **Simplified Codebase**
- **No Config Files**: Everything embedded in code
- **No DI Framework**: Direct imports and instantiation
- **Clear Dependencies**: Easy to follow code flow
- **Single Responsibility**: Each component has one job

### 🔧 **Better Maintainability**
- **Layer Isolation**: Changes in one layer don't affect others
- **Centralized Configuration**: All settings in one place
- **Consistent Error Handling**: Standardized error types
- **Clear Interfaces**: Well-defined component boundaries

### 🚀 **Enhanced Performance**
- **Compile-time Optimization**: No runtime configuration parsing
- **Reduced Memory Usage**: No configuration objects in memory
- **Faster Startup**: No configuration file loading
- **Better Caching**: Constants can be optimized by compiler

### 🧪 **Improved Testability**
- **Layer Testing**: Each layer can be tested independently
- **Mock-friendly**: Easy to create test doubles
- **Predictable Behavior**: No external configuration variables
- **Isolated Components**: Clear testing boundaries

## Migration from v1.x

The refactored version maintains full compatibility:

- **Same CLI Interface**: No changes to user experience
- **Same File Format**: Encrypted files remain compatible
- **Same Security**: All cryptographic features preserved
- **Same Performance**: All optimizations maintained

## Development Guidelines

### Adding New Features
1. **Identify Layer**: Determine which layer the feature belongs to
2. **Update Constants**: Add any new configuration to constants package
3. **Implement Bottom-Up**: Start with infrastructure, work up to presentation
4. **Maintain Boundaries**: Don't skip layers or create circular dependencies

### Error Handling
- All errors defined in [`internal/constants/errors.go`](internal/constants/errors.go)
- Use wrapped errors with context: `fmt.Errorf("operation failed: %w", err)`
- Handle errors at appropriate layer boundaries

### Testing Strategy
- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test layer interactions
- **End-to-End Tests**: Test complete workflows
- **Performance Tests**: Verify concurrent processing works correctly

## Technical Specifications

### Encryption Pipeline
```
Input File → Compression → Padding → AES-256-GCM → Reed-Solomon → Output File
```

### Decryption Pipeline
```
Input File → Reed-Solomon → AES-256-GCM → Unpadding → Decompression → Output File
```

### Concurrent Processing
- **Chunk Size**: 1MB default (configurable)
- **Worker Threads**: CPU count default (configurable)
- **Queue Size**: 100 tasks default (configurable)
- **Memory Usage**: Bounded by queue size and chunk size

### File Format
```
┌─────────────┬──────────────────────────────────────┐
│   Header    │            Encrypted Data            │
│  (128 bytes)│         (Variable Length)            │
└─────────────┴──────────────────────────────────────┘
```

## Conclusion

The refactored Hexwarden v1.1 provides:
- **Cleaner Architecture**: Layered design with clear separation
- **No External Config**: Everything embedded in code
- **Maintained Performance**: All concurrent features preserved
- **Better Maintainability**: Easier to understand and modify
- **Enhanced Security**: All cryptographic features intact

This refactoring achieves the goal of eliminating configuration files while maintaining all functionality and improving code organization.
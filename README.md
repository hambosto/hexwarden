<div align="center">
  <h1>Hexwarden v1.1</h1>
  <p>
    <a href="https://goreportcard.com/report/github.com/hambosto/hexwarden"><img src="https://goreportcard.com/badge/github.com/hambosto/hexwarden" alt="Go Report Card"></a>
    <a href="https://github.com/hambosto/hexwarden/releases"><img src="https://img.shields.io/github/v/release/hambosto/hexwarden" alt="Latest Release"></a>
    <a href="https://github.com/hambosto/hexwarden/blob/main/LICENSE"><img src="https://img.shields.io/github/license/hambosto/hexwarden" alt="License"></a>
  </p>
</div>

**Hexwarden v1.1** is a completely refactored command-line file encryption tool that makes secure file protection simple and reliable. Built with a clean layered architecture, it uses strong encryption algorithms and maintains high performance through concurrent processing.

## 🚀 What's New in v1.1

- **🏗️ Complete Refactoring**: Clean 4-layer architecture with clear separation of concerns
- **⚙️ No Configuration Files**: All settings embedded in [`internal/constants/config.go`](internal/constants/config.go) - no JSON, YAML, or ENV files
- **🔧 Simplified Dependencies**: Removed dependency injection complexity while maintaining functionality
- **⚡ Maintained Performance**: All concurrent processing and streaming features preserved
- **🛡️ Enhanced Security**: Reed-Solomon error correction with AES-256-GCM encryption
- **🖥️ Dual Interface**: Auto-detecting CLI and interactive modes in single executable
- **🚀 Modern Go**: Built with Go 1.24.5 and latest dependency versions

## Features

- **🔐 Strong Encryption**: AES-256-GCM authenticated encryption with integrity protection
- **🛠️ Error Correction**: Reed-Solomon codes protect against data corruption
- **🔑 Secure Key Derivation**: Argon2id password-based key derivation function
- **🚀 High Performance**: Concurrent processing with configurable worker pools
- **📊 Progress Tracking**: Real-time progress bars with ETA and transfer rates
- **🗑️ Secure Deletion**: Optional cryptographically secure file overwriting
- **🌐 Cross-Platform**: Works on Windows, macOS, and Linux
- **📱 User-Friendly**: Interactive CLI with intuitive prompts
- **🖥️ Automation Ready**: Full command-line interface for scripting
- **🔓 Open Source**: Fully auditable codebase

## Architecture

Hexwarden v1.1 uses a clean 4-layer architecture:

```
┌─────────────────────────────────────────┐
│           Presentation Layer            │  ← CLI Interface, Interactive UI, Progress
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│           Business Logic Layer          │  ← Encryption/Decryption Workflows
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│            Data Access Layer            │  ← File Operations, Streaming
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│          Infrastructure Layer           │  ← Crypto, Compression, Encoding
└─────────────────────────────────────────┘
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed technical documentation.

## Getting Started

### Prerequisites

- **Go**: Version 1.24.5 or higher (only if building from source)

### Installation

**From Releases (Recommended)**

1. Go to the [Releases](https://github.com/hambosto/hexwarden/releases) page
2. Download the latest binary for your operating system
3. (On macOS/Linux) Make the file executable:
   ```bash
   chmod +x hexwarden
   ```

**From Source**

1. Clone the repository:
   ```bash
   git clone https://github.com/hambosto/hexwarden.git
   ```
2. Navigate to the project directory:
   ```bash
   cd hexwarden
   ```
3. Build the project:
   ```bash
   go build -o hexwarden
   ```

## Usage

Hexwarden v1.1 supports both interactive and command-line modes:

### Interactive Mode (Default)

Run Hexwarden without arguments for the interactive experience:

```bash
./hexwarden
```

1. Choose whether you want to **Encrypt** or **Decrypt** a file
2. Select the file you want to process from the list
3. Enter a strong password to secure your file

### Command-Line Mode

Use Hexwarden in scripts and automation with the CLI interface:

**Encrypt a file:**
```bash
./hexwarden encrypt -i document.txt -o document.txt.hex
./hexwarden encrypt -i document.txt -p mypassword --delete-source
```

**Decrypt a file:**
```bash
./hexwarden decrypt -i document.txt.hex -o document.txt
./hexwarden decrypt -i document.txt.hex -p mypassword
```

**Get help:**
```bash
./hexwarden --help
./hexwarden encrypt --help
./hexwarden decrypt --help
```

### CLI Options

**Encrypt Command:**
- `-i, --input`: Input file to encrypt (required)
- `-o, --output`: Output encrypted file (default: input + .hex)
- `-p, --password`: Encryption password (will prompt if not provided)
- `--delete-source`: Delete source file after encryption
- `--secure-delete`: Use secure deletion (slower but unrecoverable)

**Decrypt Command:**
- `-i, --input`: Input file to decrypt (required)
- `-o, --output`: Output decrypted file (default: remove .hex extension)
- `-p, --password`: Decryption password (will prompt if not provided)
- `--delete-source`: Delete source file after decryption
- `--secure-delete`: Use secure deletion (slower but unrecoverable)

### Entry Points

Hexwarden provides a single main entry point that auto-detects the mode:

- **`./hexwarden`**: Main executable (auto-detects mode based on arguments)
  - With arguments: Uses CLI mode via [`internal/presentation/cli`](internal/presentation/cli)
  - Without arguments: Uses interactive mode via [`internal/presentation/interactive`](internal/presentation/interactive)

The [`cmd/`](cmd) directory contains lightweight wrapper functions that delegate to the internal presentation layer implementations.

## How It Works

Hexwarden v1.1 uses a sophisticated multi-stage processing pipeline:

### Encryption Pipeline
```
Input File → Compression → Padding → AES-256-GCM → Reed-Solomon → Output File
```

### Decryption Pipeline
```
Input File → Reed-Solomon → AES-256-GCM → Unpadding → Decompression → Output File
```

### Key Components

1. **🗜️ Compression**: Files are compressed using gzip to reduce size
2. **🔐 Encryption**: Data is encrypted using AES-256-GCM for confidentiality and integrity
3. **🛡️ Error Correction**: Reed-Solomon codes add redundancy to protect against corruption
4. **🔑 Key Derivation**: Argon2id transforms passwords into strong encryption keys
5. **📋 Header Protection**: File metadata is protected with multiple integrity checks
6. **⚡ Concurrent Processing**: Large files are processed in parallel chunks

## Security

Hexwarden v1.1 is designed with security as the top priority:

### Cryptographic Features
- **AES-256-GCM**: Industry-standard authenticated encryption
- **Argon2id**: Modern, secure password-based key derivation
- **Reed-Solomon**: Error correction codes for data integrity
- **Secure Random**: Cryptographically secure nonce and salt generation
- **Header Protection**: Multiple layers of tamper detection

### Security Best Practices
- Use strong, unique passwords (minimum 8 characters recommended)
- Keep your passwords safe and don't share them
- Verify file integrity after encryption/decryption
- Use secure deletion for sensitive source files

## Performance

Hexwarden v1.1 is optimized for performance:

- **⚡ Concurrent Processing**: Multi-threaded encryption/decryption
- **🔄 Streaming**: Memory-efficient processing of large files
- **📊 Progress Tracking**: Real-time progress with ETA
- **🎯 Configurable**: Adjustable chunk sizes and worker counts
- **💾 Memory Efficient**: Bounded memory usage regardless of file size

### Performance Configuration

All performance settings are embedded in [`internal/constants/config.go`](internal/constants/config.go):

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
```

## File Format

Encrypted files use a secure format with integrity protection:

```
┌─────────────┬──────────────────────────────────────┐
│   Header    │            Encrypted Data            │
│  (128 bytes)│         (Variable Length)            │
└─────────────┴──────────────────────────────────────┘
```

The header contains:
- Magic bytes for file type identification
- Salt for key derivation
- Original file size
- Nonce for encryption
- Integrity hash (SHA-256)
- Authentication tag (HMAC-SHA256)
- CRC32 checksum

## Error Recovery

Reed-Solomon error correction provides robust protection:

- **Data Shards**: 4 (configurable)
- **Parity Shards**: 10 (configurable)
- **Recovery Capability**: Can recover from up to 10 corrupted data segments
- **Automatic Detection**: Corruption is detected and corrected transparently

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/hambosto/hexwarden.git
cd hexwarden

# Build the main application (supports both modes)
go build -o hexwarden

# Run tests
go test ./...

# Run with race detection
go run -race main.go
```

### Project Structure

```
hexwarden/
├── main.go                         # Main entry point (auto-detects mode)
├── cmd/
│   ├── cli/cli.go                  # CLI wrapper
│   └── interactive/interactive.go  # Interactive wrapper
├── internal/
│   ├── constants/                  # Configuration and constants
│   │   ├── config.go               # App & processing config
│   │   ├── errors.go               # Error definitions
│   │   └── types.go                # Type definitions
│   ├── infrastructure/             # Low-level components
│   │   ├── crypto/                 # Cryptographic operations
│   │   ├── compression/            # Data compression
│   │   ├── encoding/               # Reed-Solomon encoding
│   │   ├── utils/                  # Utility functions
│   │   └── processor.go            # Combined processor
│   ├── data/                       # Data access layer
│   │   ├── files/                  # File operations
│   │   └── streaming/              # Stream processing
│   ├── business/                   # Business logic layer
│   │   └── operations/             # Encryption/decryption operations
│   └── presentation/               # Presentation layer
│       ├── cli/                    # Command-line interface
│       ├── interactive/            # Interactive mode
│       └── ui/                     # UI components
├── ARCHITECTURE.md                 # Detailed technical documentation
└── README.md                       # This file
```

## Future Compatibility

Hexwarden v1.1 is designed for future compatibility:

- **✅ Same File Format**: Existing encrypted files work without changes
- **✅ Same Security**: All cryptographic features preserved
- **✅ Same Performance**: All optimizations maintained
- **✅ Enhanced Interface**: Dual-mode operation (CLI and interactive)
- **✅ Enhanced Architecture**: Clean layered architecture with no external configuration

## Contributing

We welcome contributions! Here's how you can help:

1. **🐛 Report Bugs**: Open an [issue](https://github.com/hambosto/hexwarden/issues) with details
2. **💡 Suggest Features**: Share your ideas for improvements
3. **🔧 Submit Code**: Fork the repo and submit a pull request
4. **📖 Improve Docs**: Help make the documentation better
5. **🧪 Add Tests**: Increase test coverage

### Development Guidelines

- Follow the 4-layer architecture principles (Presentation → Business → Data → Infrastructure)
- Add new configuration to [`internal/constants/config.go`](internal/constants/config.go)
- Maintain backward compatibility with existing encrypted files
- Include tests for new features
- Update documentation as needed
- Dependencies flow downward only (no circular dependencies)

## License

Hexwarden is open-source software licensed under the [MIT License](LICENSE).

## Acknowledgments

- **Reed-Solomon**: Uses [klauspost/reedsolomon](https://github.com/klauspost/reedsolomon) v1.12.5 for error correction
- **Progress Bars**: Uses [schollz/progressbar](https://github.com/schollz/progressbar) v3.18.0 for progress tracking
- **CLI Prompts**: Uses [AlecAivazis/survey](https://github.com/AlecAivazis/survey) v2.3.7 for interactive prompts
- **CLI Framework**: Uses [spf13/cobra](https://github.com/spf13/cobra) v1.9.1 for command-line interface
- **Terminal Control**: Uses [inancgumus/screen](https://github.com/inancgumus/screen) for terminal operations
- **Cryptography**: Built on Go's standard crypto libraries and [golang.org/x/crypto](https://golang.org/x/crypto) v0.40.0

---

<div align="center">
  <p><strong>Hexwarden v1.1 - Secure, Fast, Simple</strong></p>
  <p>Made with ❤️ for secure file encryption</p>
</div>
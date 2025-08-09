package constants

// Application Configuration
const (
	AppName       = "hexwarden"
	AppVersion    = "1.0"
	FileExtension = ".hex"
)

// Processing Configuration
const (
	DefaultChunkSize = 1 * 1024 * 1024 // 1MB chunks
	MaxConcurrency   = 8               // Max worker threads
	QueueSize        = 100             // Task queue buffer size
	OverwritePasses  = 3               // Secure deletion passes
)

// Cryptographic Configuration
const (
	SaltSize = 32 // Argon2id salt size
	KeySize  = 32 // AES-256 key size
)

// File Processing Configuration
var (
	ExcludedDirs = []string{
		"vendor/", "node_modules/", ".git", ".github",
		".vscode/", "build/", "dist/", "target/",
	}

	ExcludedExts = []string{
		".go", "go.mod", "go.sum", ".nix", ".gitignore",
		".exe", ".dll", ".so", ".dylib",
	}
)

// Argon2id Parameters
const (
	ArgonTime    uint32 = 3         // Time cost
	ArgonMemory  uint32 = 64 * 1024 // Memory cost (64MB)
	ArgonThreads uint8  = 4         // Parallelism
)

// Header Format Constants
const (
	MagicBytes        = "HWX2" // File type identifier
	SaltSizeBytes     = 32     // Salt for KDF
	OriginalSizeBytes = 8      // Size of original plaintext
	NonceSizeBytes    = 16     // Nonce for AEAD encryption
	IntegritySize     = 32     // SHA-256 hash size
	AuthSize          = 32     // HMAC-SHA256 tag size
	ChecksumSize      = 4      // CRC32 checksum size
	TotalHeaderSize   = 128    // Fixed header size
)

// Stream Processing Constants
const (
	ChunkHeaderSize = 4 // Size of chunk length header in bytes
)

// Reed-Solomon Encoding Configuration
const (
	DataShards   = 4       // Number of data shards
	ParityShards = 10      // Number of parity shards for error correction
	MaxDataLen   = 1 << 30 // Maximum data length (1GB)
)

// Padding Configuration
const (
	PaddingSize = 16 // Block size for padding
)

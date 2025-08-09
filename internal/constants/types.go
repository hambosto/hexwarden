package constants

// ProcessorMode represents the operation type for file processing
type ProcessorMode string

const (
	// ModeEncrypt indicates files should be encrypted
	ModeEncrypt ProcessorMode = "Encrypt"
	// ModeDecrypt indicates files should be decrypted
	ModeDecrypt ProcessorMode = "Decrypt"
)

// DeleteOption represents different file deletion methods
type DeleteOption string

const (
	// DeleteStandard performs normal file deletion (faster, but recoverable)
	DeleteStandard DeleteOption = "Normal Delete (faster, but recoverable)"
	// DeleteSecure performs secure file deletion (slower, but unrecoverable)
	DeleteSecure DeleteOption = "Secure Delete (slower, but unrecoverable)"
)

// Processing represents the stream processing operation type
type Processing int

const (
	// Encryption indicates data should be encrypted
	Encryption Processing = iota
	// Decryption indicates data should be decrypted
	Decryption
)

// CompressionLevel represents compression levels
type CompressionLevel int

const (
	// LevelNoCompression disables compression
	LevelNoCompression CompressionLevel = 0
	// LevelBestSpeed provides fastest compression
	LevelBestSpeed CompressionLevel = 1
	// LevelDefaultCompression provides balanced compression
	LevelDefaultCompression CompressionLevel = 6
	// LevelBestCompression provides maximum compression
	LevelBestCompression CompressionLevel = 9
)

// Task represents a processing task for concurrent operations
type Task struct {
	Data  []byte
	Index uint64
}

// TaskResult represents the result of a processed task
type TaskResult struct {
	Index uint64
	Data  []byte
	Size  int
	Err   error
}

// FileInfo represents file information for processing
type FileInfo struct {
	Path        string
	Size        int64
	IsEncrypted bool
	IsEligible  bool
}

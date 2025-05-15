package kdf

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrInvalidSaltLength = errors.New("salt length doesn't match configuration")
	ErrInvalidParameters = errors.New("invalid parameters")
)

type Parameters struct {
	MemoryMB    uint32
	Iterations  uint32
	Parallelism uint8
	KeyBytes    uint32
	SaltBytes   uint32
}

func DefaultParameters() Parameters {
	return Parameters{
		MemoryMB:    64, // 64MB
		Iterations:  4,  // 4 iterations
		Parallelism: 4,  // 4 threads
		KeyBytes:    64, // 64 byte key
		SaltBytes:   32, // 32 byte salt
	}
}

func MinimumParameters() Parameters {
	return Parameters{
		MemoryMB:    8,  // 8MB minimum
		Iterations:  1,  // At least 1 iteration
		Parallelism: 1,  // At least 1 thread
		KeyBytes:    16, // At least 16 byte key
		SaltBytes:   16, // At least 16 byte salt
	}
}

func (p Parameters) Validate() error {
	min := MinimumParameters()

	if p.MemoryMB < min.MemoryMB {
		return fmt.Errorf("%w: memory must be at least %d MB", ErrInvalidParameters, min.MemoryMB)
	}
	if p.Iterations < min.Iterations {
		return fmt.Errorf("%w: iterations must be at least %d", ErrInvalidParameters, min.Iterations)
	}
	if p.Parallelism < min.Parallelism {
		return fmt.Errorf("%w: parallelism must be at least %d", ErrInvalidParameters, min.Parallelism)
	}
	if p.KeyBytes < min.KeyBytes {
		return fmt.Errorf("%w: key length must be at least %d bytes", ErrInvalidParameters, min.KeyBytes)
	}
	if p.SaltBytes < min.SaltBytes {
		return fmt.Errorf("%w: salt length must be at least %d bytes", ErrInvalidParameters, min.SaltBytes)
	}

	return nil
}

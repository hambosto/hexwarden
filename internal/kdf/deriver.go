package kdf

import (
	"crypto/rand"
	"fmt"
	"math"

	"golang.org/x/crypto/argon2"
)

type Deriver interface {
	DeriveKey(password, salt []byte) ([]byte, error)
	GenerateSalt() ([]byte, error)
	GetParameters() Parameters
}

type ArgonDeriver struct {
	params Parameters
}

func NewDeriver(params *Parameters) (Deriver, error) {
	p := DefaultParameters()
	if params != nil {
		p = *params
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return &ArgonDeriver{
		params: p,
	}, nil
}

func (d *ArgonDeriver) DeriveKey(password, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, ErrEmptyPassword
	}

	saltLen := len(salt)
	if saltLen < 0 || saltLen > math.MaxUint32 {
		return nil, fmt.Errorf("%w: salt length out of range: %d",
			ErrInvalidSaltLength, saltLen)
	}

	saltLenUint32 := uint32(saltLen)
	if saltLenUint32 != d.params.SaltBytes {
		return nil, fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidSaltLength,
			d.params.SaltBytes,
			saltLen,
		)
	}

	key := argon2.IDKey(
		password,
		salt,
		d.params.Iterations,
		d.params.MemoryMB*1024, // Convert MB to KB
		d.params.Parallelism,
		d.params.KeyBytes,
	)

	return key, nil
}

func (d *ArgonDeriver) GenerateSalt() ([]byte, error) {
	salt := make([]byte, d.params.SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}

	return salt, nil
}

func (d *ArgonDeriver) GetParameters() Parameters {
	return d.params
}

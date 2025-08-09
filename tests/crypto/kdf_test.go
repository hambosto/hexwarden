package crypto

import (
	"bytes"
	"testing"

	"github.com/hambosto/hexwarden/internal/constants"
	"github.com/hambosto/hexwarden/internal/infrastructure/crypto"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestDeriveKey(t *testing.T) {
	testData := helpers.NewTestData()

	tests := []struct {
		name        string
		password    []byte
		salt        []byte
		expectError bool
		expectedErr error
	}{
		{
			name:        "Valid password and salt",
			password:    []byte(testData.TestPassword),
			salt:        testData.ValidSalt,
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    []byte{},
			salt:        testData.ValidSalt,
			expectError: true,
			expectedErr: constants.ErrEmptyPassword,
		},
		{
			name:        "Nil password",
			password:    nil,
			salt:        testData.ValidSalt,
			expectError: true,
			expectedErr: constants.ErrEmptyPassword,
		},
		{
			name:        "Invalid salt size - too short",
			password:    []byte(testData.TestPassword),
			salt:        []byte("short"),
			expectError: true,
			expectedErr: constants.ErrInvalidSalt,
		},
		{
			name:        "Invalid salt size - too long",
			password:    []byte(testData.TestPassword),
			salt:        make([]byte, constants.SaltSize+1),
			expectError: true,
			expectedErr: constants.ErrInvalidSalt,
		},
		{
			name:        "Nil salt",
			password:    []byte(testData.TestPassword),
			salt:        nil,
			expectError: true,
			expectedErr: constants.ErrInvalidSalt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := crypto.DeriveKey(tt.password, tt.salt)

			if tt.expectError {
				helpers.AssertError(t, err, nil)
				if key != nil {
					t.Error("Expected key to be nil when error occurs")
				}
			} else {
				helpers.AssertNoError(t, err)
				if key == nil {
					t.Error("Expected key to be non-nil when no error occurs")
				}

				// Key should be the correct size
				helpers.AssertEqual(t, constants.KeySize, len(key))

				// Key should not be all zeros
				allZero := true
				for _, b := range key {
					if b != 0 {
						allZero = false
						break
					}
				}
				if allZero {
					t.Error("Derived key should not be all zeros")
				}
			}
		})
	}
}

func TestDeriveKey_Deterministic(t *testing.T) {
	testData := helpers.NewTestData()
	password := []byte(testData.TestPassword)
	salt := testData.ValidSalt

	// Derive the same key multiple times
	keys := make([][]byte, 5)
	for i := range 5 {
		key, err := crypto.DeriveKey(password, salt)
		helpers.AssertNoError(t, err)
		keys[i] = key
	}

	// All keys should be identical (deterministic)
	for i := 1; i < len(keys); i++ {
		helpers.AssertBytesEqual(t, keys[0], keys[i])
	}
}

func TestDeriveKey_DifferentPasswords(t *testing.T) {
	testData := helpers.NewTestData()
	salt := testData.ValidSalt

	passwords := []string{
		"password1",
		"password2",
		"different-password",
		"P@ssw0rd!",
		"very-long-password-with-special-characters-123456789",
	}

	keys := make([][]byte, len(passwords))
	for i, password := range passwords {
		key, err := crypto.DeriveKey([]byte(password), salt)
		helpers.AssertNoError(t, err)
		keys[i] = key
	}

	// All keys should be different
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			helpers.AssertBytesNotEqual(t, keys[i], keys[j])
		}
	}
}

func TestDeriveKey_DifferentSalts(t *testing.T) {
	testData := helpers.NewTestData()
	password := []byte(testData.TestPassword)

	// Generate different salts
	salts := make([][]byte, 5)
	for i := range 5 {
		salt, err := crypto.GenerateSalt()
		helpers.AssertNoError(t, err)
		salts[i] = salt
	}

	// Derive keys with different salts
	keys := make([][]byte, len(salts))
	for i, salt := range salts {
		key, err := crypto.DeriveKey(password, salt)
		helpers.AssertNoError(t, err)
		keys[i] = key
	}

	// All keys should be different
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			helpers.AssertBytesNotEqual(t, keys[i], keys[j])
		}
	}
}

func TestGenerateSalt(t *testing.T) {
	// Generate multiple salts
	salts := make([][]byte, 10)
	for i := range 10 {
		salt, err := crypto.GenerateSalt()
		helpers.AssertNoError(t, err)
		salts[i] = salt

		// Salt should be the correct size
		helpers.AssertEqual(t, constants.SaltSize, len(salt))

		// Salt should not be all zeros
		allZero := true
		for _, b := range salt {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Error("Generated salt should not be all zeros")
		}
	}

	// All salts should be different (extremely high probability)
	for i := range salts {
		for j := i + 1; j < len(salts); j++ {
			helpers.AssertBytesNotEqual(t, salts[i], salts[j])
		}
	}
}

func TestValidateSalt(t *testing.T) {
	testData := helpers.NewTestData()

	tests := []struct {
		name        string
		salt        []byte
		expectError bool
	}{
		{
			name:        "Valid salt",
			salt:        testData.ValidSalt,
			expectError: false,
		},
		{
			name:        "Weak salt - all zeros",
			salt:        testData.WeakSalt,
			expectError: true,
		},
		{
			name:        "Invalid salt size - too short",
			salt:        []byte("short"),
			expectError: true,
		},
		{
			name:        "Invalid salt size - too long",
			salt:        make([]byte, constants.SaltSize+1),
			expectError: true,
		},
		{
			name:        "Nil salt",
			salt:        nil,
			expectError: true,
		},
		{
			name:        "Empty salt",
			salt:        []byte{},
			expectError: true,
		},
		{
			name:        "Weak salt - repeating pattern",
			salt:        createRepeatingPatternSalt(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := crypto.ValidateSalt(tt.salt)

			if tt.expectError {
				helpers.AssertError(t, err, nil)
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestValidateSalt_GeneratedSaltsAreValid(t *testing.T) {
	// Generate multiple salts and ensure they all pass validation
	// Reduced from 100 to 10 iterations to reduce resource usage
	for range 10 {
		salt, err := crypto.GenerateSalt()
		helpers.AssertNoError(t, err)

		err = crypto.ValidateSalt(salt)
		helpers.AssertNoError(t, err)
	}
}

func TestDeriveKey_PasswordStrengthVariations(t *testing.T) {
	testData := helpers.NewTestData()
	salt := testData.ValidSalt

	passwords := []struct {
		name     string
		password string
	}{
		{"Short password", "abc"},
		{"Medium password", "password123"},
		{"Long password", "this-is-a-very-long-password-with-many-characters"},
		{"Special characters", "P@ssw0rd!#$%^&*()"},
		{"Unicode characters", "пароль密码パスワード"},
		{"Numbers only", "1234567890"},
		{"Mixed case", "MiXeD-CaSe-PaSsWoRd"},
	}

	keys := make([][]byte, len(passwords))
	for i, p := range passwords {
		t.Run(p.name, func(t *testing.T) {
			key, err := crypto.DeriveKey([]byte(p.password), salt)
			helpers.AssertNoError(t, err)
			helpers.AssertEqual(t, constants.KeySize, len(key))
			keys[i] = key
		})
	}

	// All keys should be different
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			helpers.AssertBytesNotEqual(t, keys[i], keys[j])
		}
	}
}

// BenchmarkDeriveKey benchmarks key derivation performance
func BenchmarkDeriveKey(b *testing.B) {
	testData := helpers.NewTestData()
	password := []byte(testData.TestPassword)
	salt := testData.ValidSalt

	for b.Loop() {
		_, err := crypto.DeriveKey(password, salt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateSalt benchmarks salt generation performance
func BenchmarkGenerateSalt(b *testing.B) {
	for b.Loop() {
		_, err := crypto.GenerateSalt()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateSalt benchmarks salt validation performance
func BenchmarkValidateSalt(b *testing.B) {
	testData := helpers.NewTestData()
	salt := testData.ValidSalt

	for b.Loop() {
		err := crypto.ValidateSalt(salt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// createRepeatingPatternSalt creates a salt with a repeating 4-byte pattern
func createRepeatingPatternSalt() []byte {
	salt := make([]byte, constants.SaltSize)
	pattern := []byte{0xAB, 0xCD, 0xEF, 0x12}

	for i := 0; i < len(salt); i += 4 {
		copy(salt[i:], pattern)
	}

	return salt
}

// TestDeriveKey_EdgeCases tests edge cases for key derivation
func TestDeriveKey_EdgeCases(t *testing.T) {
	testData := helpers.NewTestData()

	t.Run("Very long password", func(t *testing.T) {
		// Reduced from 10000 to 1000 bytes to reduce memory usage
		longPassword := bytes.Repeat([]byte("a"), 1000)
		key, err := crypto.DeriveKey(longPassword, testData.ValidSalt)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, constants.KeySize, len(key))
	})

	t.Run("Single character password", func(t *testing.T) {
		key, err := crypto.DeriveKey([]byte("a"), testData.ValidSalt)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, constants.KeySize, len(key))
	})

	t.Run("Binary password", func(t *testing.T) {
		binaryPassword := make([]byte, 256)
		for i := 0; i < 256; i++ {
			binaryPassword[i] = byte(i)
		}
		key, err := crypto.DeriveKey(binaryPassword, testData.ValidSalt)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, constants.KeySize, len(key))
	})
}

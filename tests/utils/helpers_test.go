package utils

import (
	"fmt"
	"testing"

	"github.com/hambosto/hexwarden/internal/infrastructure/utils"
	"github.com/hambosto/hexwarden/tests/helpers"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "Zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "Single byte",
			bytes:    1,
			expected: "1 B",
		},
		{
			name:     "Less than 1KB",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "Exactly 1KB",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "1.5 KB",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "Exactly 1MB",
			bytes:    1024 * 1024,
			expected: "1.0 MB",
		},
		{
			name:     "2.5 MB",
			bytes:    1024*1024*2 + 1024*512,
			expected: "2.5 MB",
		},
		{
			name:     "Exactly 1GB",
			bytes:    1024 * 1024 * 1024,
			expected: "1.0 GB",
		},
		{
			name:     "1.2 GB",
			bytes:    1024*1024*1024 + 1024*1024*200,
			expected: "1.2 GB",
		},
		{
			name:     "Exactly 1TB",
			bytes:    1024 * 1024 * 1024 * 1024,
			expected: "1.0 TB",
		},
		{
			name:     "Large value",
			bytes:    1024*1024*1024*1024*5 + 1024*1024*1024*512,
			expected: "5.5 TB",
		},
		{
			name:     "Very large value (PB)",
			bytes:    1024 * 1024 * 1024 * 1024 * 1024 * 2,
			expected: "2.0 PB",
		},
		{
			name:     "Extremely large value (EB)",
			bytes:    1024 * 1024 * 1024 * 1024 * 1024 * 1024 * 3,
			expected: "3.0 EB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.FormatBytes(tt.bytes)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestFormatBytes_EdgeCases(t *testing.T) {
	t.Run("Negative bytes", func(t *testing.T) {
		result := utils.FormatBytes(-1)
		// Should handle negative values gracefully
		if result == "" {
			t.Error("FormatBytes should not return empty string for negative values")
		}
	})

	t.Run("Maximum int64", func(t *testing.T) {
		result := utils.FormatBytes(9223372036854775807) // Max int64
		if result == "" {
			t.Error("FormatBytes should handle maximum int64 value")
		}
		t.Logf("Max int64 formatted as: %s", result)
	})
}

func TestMinInt64(t *testing.T) {
	tests := []struct {
		name     string
		a        int64
		b        int64
		expected int64
	}{
		{
			name:     "First smaller",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "Second smaller",
			a:        15,
			b:        8,
			expected: 8,
		},
		{
			name:     "Equal values",
			a:        7,
			b:        7,
			expected: 7,
		},
		{
			name:     "Zero and positive",
			a:        0,
			b:        5,
			expected: 0,
		},
		{
			name:     "Negative and positive",
			a:        -3,
			b:        2,
			expected: -3,
		},
		{
			name:     "Both negative",
			a:        -10,
			b:        -5,
			expected: -10,
		},
		{
			name:     "Large values",
			a:        1000000000,
			b:        999999999,
			expected: 999999999,
		},
		{
			name:     "Maximum values",
			a:        9223372036854775807, // Max int64
			b:        9223372036854775806,
			expected: 9223372036854775806,
		},
		{
			name:     "Minimum values",
			a:        -9223372036854775808, // Min int64
			b:        -9223372036854775807,
			expected: -9223372036854775808,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.MinInt64(tt.a, tt.b)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "First smaller",
			a:        3,
			b:        8,
			expected: 3,
		},
		{
			name:     "Second smaller",
			a:        12,
			b:        4,
			expected: 4,
		},
		{
			name:     "Equal values",
			a:        6,
			b:        6,
			expected: 6,
		},
		{
			name:     "Zero and positive",
			a:        0,
			b:        3,
			expected: 0,
		},
		{
			name:     "Negative and positive",
			a:        -2,
			b:        1,
			expected: -2,
		},
		{
			name:     "Both negative",
			a:        -7,
			b:        -3,
			expected: -7,
		},
		{
			name:     "Large values",
			a:        1000000,
			b:        999999,
			expected: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.MinInt(tt.a, tt.b)
			helpers.AssertEqual(t, tt.expected, result)
		})
	}
}

func TestMinInt_EdgeCases(t *testing.T) {
	t.Run("Maximum int values", func(t *testing.T) {
		// Test with large int values (platform dependent)
		a := 2147483647 // Typical max int32
		b := 2147483646
		result := utils.MinInt(a, b)
		helpers.AssertEqual(t, b, result)
	})

	t.Run("Minimum int values", func(t *testing.T) {
		// Test with small int values (platform dependent)
		a := -2147483648 // Typical min int32
		b := -2147483647
		result := utils.MinInt(a, b)
		helpers.AssertEqual(t, a, result)
	})
}

// BenchmarkFormatBytes benchmarks the FormatBytes function
func BenchmarkFormatBytes(b *testing.B) {
	testValues := []int64{
		0,
		1024,
		1024 * 1024,
		1024 * 1024 * 1024,
		1024 * 1024 * 1024 * 1024,
	}

	for _, value := range testValues {
		b.Run(fmt.Sprintf("Value_%d", value), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				utils.FormatBytes(value)
			}
		})
	}
}

// BenchmarkMinInt64 benchmarks the MinInt64 function
func BenchmarkMinInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utils.MinInt64(12345, 67890)
	}
}

// BenchmarkMinInt benchmarks the MinInt function
func BenchmarkMinInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utils.MinInt(123, 456)
	}
}

// TestFormatBytes_Consistency tests that formatting is consistent
func TestFormatBytes_Consistency(t *testing.T) {
	// Test that the same input always produces the same output
	testValue := int64(1536) // 1.5 KB
	expected := utils.FormatBytes(testValue)

	for i := 0; i < 100; i++ {
		result := utils.FormatBytes(testValue)
		helpers.AssertEqual(t, expected, result)
	}
}

// TestMinFunctions_Commutativity tests that min functions are commutative
func TestMinFunctions_Commutativity(t *testing.T) {
	t.Run("MinInt64 commutativity", func(t *testing.T) {
		a, b := int64(42), int64(24)
		result1 := utils.MinInt64(a, b)
		result2 := utils.MinInt64(b, a)
		helpers.AssertEqual(t, result1, result2)
	})

	t.Run("MinInt commutativity", func(t *testing.T) {
		a, b := 42, 24
		result1 := utils.MinInt(a, b)
		result2 := utils.MinInt(b, a)
		helpers.AssertEqual(t, result1, result2)
	})
}

// TestMinFunctions_Transitivity tests transitivity property
func TestMinFunctions_Transitivity(t *testing.T) {
	t.Run("MinInt64 transitivity", func(t *testing.T) {
		a, b, c := int64(10), int64(5), int64(15)

		// If min(a,b) = b and min(b,c) = b, then min(a,c) should be min(a,b)
		minAB := utils.MinInt64(a, b)
		minBC := utils.MinInt64(b, c)
		minAC := utils.MinInt64(a, c)

		if minAB == b && minBC == b {
			helpers.AssertEqual(t, minAB, utils.MinInt64(minAC, b))
		}
	})

	t.Run("MinInt transitivity", func(t *testing.T) {
		a, b, c := 10, 5, 15

		minAB := utils.MinInt(a, b)
		minBC := utils.MinInt(b, c)
		minAC := utils.MinInt(a, c)

		if minAB == b && minBC == b {
			helpers.AssertEqual(t, minAB, utils.MinInt(minAC, b))
		}
	})
}

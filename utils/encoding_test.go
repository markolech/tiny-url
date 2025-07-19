package utils

import (
	"testing"
)

func TestEncodeBase62(t *testing.T) {
	testCases := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "a"},
		{35, "z"},
		{36, "A"},
		{61, "Z"},
		{62, "10"},
		{63, "11"},
		{124, "20"},
		{1000, "g8"},
		{3844, "100"},   // 62^2
		{238328, "1000"}, // 62^3
		{1000000000, "15FTGg"}, // 1 billion
	}

	for _, tc := range testCases {
		result := EncodeBase62(tc.input)
		if result != tc.expected {
			t.Errorf("EncodeBase62(%d) = %s; expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestDecodeBase62(t *testing.T) {
	testCases := []struct {
		input    string
		expected uint64
	}{
		{"0", 0},
		{"1", 1},
		{"9", 9},
		{"a", 10},
		{"z", 35},
		{"A", 36},
		{"Z", 61},
		{"10", 62},
		{"11", 63},
		{"20", 124},
		{"g8", 1000},
		{"100", 3844},
		{"1000", 238328},
		{"15FTGg", 1000000000},
	}

	for _, tc := range testCases {
		result := DecodeBase62(tc.input)
		if result != tc.expected {
			t.Errorf("DecodeBase62(%s) = %d; expected %d", tc.input, result, tc.expected)
		}
	}
}

func TestBase62RoundTrip(t *testing.T) {
	// Test that encoding then decoding returns the original value
	testValues := []uint64{
		0, 1, 10, 61, 62, 100, 1000, 10000, 100000, 1000000, 1000000000,
	}

	for _, val := range testValues {
		encoded := EncodeBase62(val)
		decoded := DecodeBase62(encoded)
		if decoded != val {
			t.Errorf("Round trip failed for %d: encoded to %s, decoded to %d", val, encoded, decoded)
		}
	}
}

func TestDecodeBase62InvalidChars(t *testing.T) {
	invalidInputs := []string{
		"@",     // Invalid character
		"a@b",   // Invalid character in middle
		"!123",  // Invalid character at start
		"abc!",  // Invalid character at end
	}

	for _, input := range invalidInputs {
		result := DecodeBase62(input)
		if result != 0 {
			t.Errorf("DecodeBase62(%s) should return 0 for invalid input, got %d", input, result)
		}
	}
}

func TestBase62EmptyString(t *testing.T) {
	result := DecodeBase62("")
	if result != 0 {
		t.Errorf("DecodeBase62(\"\") should return 0, got %d", result)
	}
}

func BenchmarkEncodeBase62(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeBase62(uint64(i))
	}
}

func BenchmarkDecodeBase62(b *testing.B) {
	encoded := EncodeBase62(1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeBase62(encoded)
	}
} 
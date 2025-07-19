package utils

// Base62 characters: 0-9, a-z, A-Z (62 characters total)
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// EncodeBase62 converts a numeric ID to a base62 string
// Example: 1 -> "1", 62 -> "10", 63 -> "11"
func EncodeBase62(id uint64) string {
	if id == 0 {
		return "0"
	}

	result := ""
	for id > 0 {
		result = string(base62Chars[id%62]) + result
		id /= 62
	}
	return result
}

// DecodeBase62 converts a base62 string back to a numeric ID
func DecodeBase62(encoded string) uint64 {
	result := uint64(0)
	base := uint64(1)

	for i := len(encoded) - 1; i >= 0; i-- {
		char := encoded[i]
		var value uint64

		switch {
		case char >= '0' && char <= '9':
			value = uint64(char - '0')
		case char >= 'a' && char <= 'z':
			value = uint64(char-'a') + 10
		case char >= 'A' && char <= 'Z':
			value = uint64(char-'A') + 36
		default:
			// Invalid character, return 0
			return 0
		}

		result += value * base
		base *= 62
	}

	return result
} 
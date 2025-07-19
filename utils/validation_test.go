package utils

import (
	"testing"
)

func TestIsValidURL(t *testing.T) {
	validURLs := []string{
		"http://example.com",
		"https://example.com",
		"http://www.example.com",
		"https://www.example.com",
		"http://example.com/path",
		"https://example.com/path/to/resource",
		"http://example.com:8080",
		"https://example.com:443/path",
		"http://subdomain.example.com",
		"https://subdomain.example.com/path?query=value",
		"http://example.com/path?query=value&other=param",
		"https://example.com/path#fragment",
		"http://192.168.1.1",
		"https://192.168.1.1:8080/path",
		"http://localhost",
		"https://localhost:3000",
		"http://example.com/very/long/path/with/many/segments",
		"https://api.example.com/v1/users/123?include=profile&format=json",
	}

	for _, url := range validURLs {
		if !IsValidURL(url) {
			t.Errorf("IsValidURL(%s) = false; expected true", url)
		}
	}
}

func TestIsValidURLInvalid(t *testing.T) {
	invalidURLs := []string{
		"",                           // Empty string
		"   ",                        // Whitespace only
		"example.com",                // Missing scheme
		"ftp://example.com",          // Wrong scheme
		"mailto:user@example.com",    // Wrong scheme
		"file:///path/to/file",       // Wrong scheme
		"http://",                    // Missing host
		"https://",                   // Missing host
		"http:///path",               // Missing host
		"not-a-url",                  // Not a URL
		"http:/example.com",          // Malformed (single slash)
		"ttp://example.com",          // Missing h
		"http//example.com",          // Missing colon
		"javascript:alert('xss')",    // JavaScript scheme
		"data:text/plain;base64,SGVsbG8=", // Data scheme
	}

	for _, url := range invalidURLs {
		if IsValidURL(url) {
			t.Errorf("IsValidURL(%s) = true; expected false", url)
		}
	}
}

func TestIsValidURLEdgeCases(t *testing.T) {
	edgeCases := []struct {
		url      string
		expected bool
		desc     string
	}{
		{"HTTP://EXAMPLE.COM", true, "Uppercase scheme should work"},
		{"HTTPS://EXAMPLE.COM", true, "Uppercase scheme should work"},
		{"http://EXAMPLE.COM", true, "Uppercase host should work"},
		{"http://example.com/", true, "Trailing slash should work"},
		{"http://example.com//", true, "Double slash in path should work"},
		{"http://example.com?", true, "Empty query should work"},
		{"http://example.com#", true, "Empty fragment should work"},
		{"http://example.com:80", true, "Default HTTP port should work"},
		{"https://example.com:443", true, "Default HTTPS port should work"},
		{"http://example.com:0", true, "Port 0 should work (parsed as valid)"},
		{"http://user:pass@example.com", true, "Basic auth should work"},
	}

	for _, tc := range edgeCases {
		result := IsValidURL(tc.url)
		if result != tc.expected {
			t.Errorf("IsValidURL(%s) = %v; expected %v (%s)", tc.url, result, tc.expected, tc.desc)
		}
	}
}

func BenchmarkIsValidURL(b *testing.B) {
	url := "https://www.example.com/path/to/resource?query=value"
	for i := 0; i < b.N; i++ {
		IsValidURL(url)
	}
} 
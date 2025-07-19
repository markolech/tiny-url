package utils

import (
	"net/url"
	"strings"
)

// IsValidURL validates that a string is a proper HTTP or HTTPS URL
func IsValidURL(urlStr string) bool {
	// Basic empty check
	if strings.TrimSpace(urlStr) == "" {
		return false
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Check scheme is HTTP or HTTPS
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	// Check host is present
	if parsedURL.Host == "" {
		return false
	}

	return true
} 
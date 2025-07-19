package storage

import (
	"tiny-url-service/models"
)

// Storage defines the interface for URL storage operations
type Storage interface {
	// Store saves a URL mapping and returns the generated short code
	Store(mapping *models.URLMapping) (string, error)
	
	// Get retrieves the URL mapping for a given short code
	Get(shortCode string) (*models.URLMapping, error)
	
	// IsExpired checks if a URL mapping has expired
	IsExpired(mapping *models.URLMapping) bool
	
	// GetStats returns storage statistics
	GetStats() map[string]interface{}
} 
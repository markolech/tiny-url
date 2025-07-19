package storage

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"tiny-url-service/models"
	"tiny-url-service/utils"
)

// MemoryStorage implements the Storage interface using in-memory maps
type MemoryStorage struct {
	mu       sync.RWMutex                 // Protects the maps
	urls     map[string]*models.URLMapping // shortCode -> URLMapping
	counter  uint64                       // Atomic counter for unique IDs
	baseURL  string                       // Base URL for generating short URLs
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage(baseURL string) *MemoryStorage {
	return &MemoryStorage{
		urls:    make(map[string]*models.URLMapping),
		counter: 0,
		baseURL: baseURL,
	}
}

// Store saves a URL mapping and returns the generated short code
func (m *MemoryStorage) Store(mapping *models.URLMapping) (string, error) {
	// Generate unique ID
	id := atomic.AddUint64(&m.counter, 1)
	
	// Generate short code using base62 encoding
	shortCode := utils.EncodeBase62(id)
	
	// Complete the mapping
	mapping.ID = id
	mapping.ShortCode = shortCode
	mapping.CreatedAt = time.Now()
	
	// Store with write lock
	m.mu.Lock()
	m.urls[shortCode] = mapping
	m.mu.Unlock()
	
	return shortCode, nil
}

// Get retrieves the URL mapping for a given short code
func (m *MemoryStorage) Get(shortCode string) (*models.URLMapping, error) {
	m.mu.RLock()
	mapping, exists := m.urls[shortCode]
	m.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("short code not found: %s", shortCode)
	}
	
	// Check if expired
	if m.IsExpired(mapping) {
		return nil, fmt.Errorf("URL has expired: %s", shortCode)
	}
	
	return mapping, nil
}

// IsExpired checks if a URL mapping has expired
func (m *MemoryStorage) IsExpired(mapping *models.URLMapping) bool {
	if mapping.ExpirationDate == nil {
		return false // No expiration set
	}
	return time.Now().After(*mapping.ExpirationDate)
}

// GetStats returns storage statistics
func (m *MemoryStorage) GetStats() map[string]interface{} {
	m.mu.RLock()
	totalUrls := len(m.urls)
	m.mu.RUnlock()
	
	currentCounter := atomic.LoadUint64(&m.counter)
	
	return map[string]interface{}{
		"total_urls":      totalUrls,
		"current_counter": currentCounter,
		"storage_type":    "memory",
	}
} 
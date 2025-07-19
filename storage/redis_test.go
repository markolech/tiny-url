package storage

import (
	"testing"
	"time"
	"tiny-url-service/models"

	"github.com/alicebob/miniredis/v2"
)

func setupMockRedis(t *testing.T, baseURL string) (*RedisStorage, *miniredis.Miniredis) {
	// Create an in-memory Redis mock
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	// Create Redis storage with mock
	storage, err := NewRedisStorage(baseURL, "redis://"+s.Addr())
	if err != nil {
		s.Close()
		t.Fatalf("Failed to create Redis storage: %v", err)
	}

	return storage, s
}

func TestRedisStorage_Store(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	mapping := &models.URLMapping{
		LongURL: "https://www.example.com",
	}

	shortCode, err := storage.Store(mapping)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	if shortCode == "" {
		t.Error("Store() returned empty short code")
	}

	if mapping.ID == 0 {
		t.Error("Store() did not set ID")
	}

	if mapping.ShortCode != shortCode {
		t.Errorf("Store() set ShortCode to %s, expected %s", mapping.ShortCode, shortCode)
	}

	if mapping.CreatedAt.IsZero() {
		t.Error("Store() did not set CreatedAt")
	}
}

func TestRedisStorage_Get(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Store a URL first
	original := &models.URLMapping{
		LongURL: "https://www.example.com/test",
	}

	shortCode, err := storage.Store(original)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Retrieve the URL
	retrieved, err := storage.Get(shortCode)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.LongURL != original.LongURL {
		t.Errorf("Get() returned LongURL %s, expected %s", retrieved.LongURL, original.LongURL)
	}

	if retrieved.ShortCode != shortCode {
		t.Errorf("Get() returned ShortCode %s, expected %s", retrieved.ShortCode, shortCode)
	}

	if retrieved.ID != original.ID {
		t.Errorf("Get() returned ID %d, expected %d", retrieved.ID, original.ID)
	}
}

func TestRedisStorage_GetNotFound(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	_, err := storage.Get("nonexistent")
	if err == nil {
		t.Error("Get() should return error for non-existent short code")
	}
}

func TestRedisStorage_UniqueIDs(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	urls := []string{
		"https://www.example1.com",
		"https://www.example2.com",
		"https://www.example3.com",
	}

	var mappings []*models.URLMapping
	for _, url := range urls {
		mapping := &models.URLMapping{LongURL: url}
		_, err := storage.Store(mapping)
		if err != nil {
			t.Fatalf("Store() failed: %v", err)
		}
		mappings = append(mappings, mapping)
	}

	// Check that all IDs are unique
	for i := 0; i < len(mappings); i++ {
		for j := i + 1; j < len(mappings); j++ {
			if mappings[i].ID == mappings[j].ID {
				t.Errorf("Duplicate ID %d found", mappings[i].ID)
			}
			if mappings[i].ShortCode == mappings[j].ShortCode {
				t.Errorf("Duplicate ShortCode %s found", mappings[i].ShortCode)
			}
		}
	}
}

func TestRedisStorage_Expiration(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Create URL with expiration
	expirationTime := time.Now().Add(time.Hour)
	mapping := &models.URLMapping{
		LongURL:        "https://www.example.com/expire",
		ExpirationDate: &expirationTime,
	}

	shortCode, err := storage.Store(mapping)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Should be able to retrieve it
	retrieved, err := storage.Get(shortCode)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.ExpirationDate == nil {
		t.Error("ExpirationDate should not be nil")
	} else if !retrieved.ExpirationDate.Equal(expirationTime) {
		t.Errorf("ExpirationDate mismatch: got %v, expected %v", retrieved.ExpirationDate, expirationTime)
	}
}

func TestRedisStorage_GetStats(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Store some URLs
	for i := 0; i < 3; i++ {
		mapping := &models.URLMapping{
			LongURL: "https://www.example.com/" + string(rune('a'+i)),
		}
		_, err := storage.Store(mapping)
		if err != nil {
			t.Fatalf("Store() failed: %v", err)
		}
	}

	stats := storage.GetStats()

	if stats["total_urls"] != int64(3) {
		t.Errorf("total_urls should be 3, got %v", stats["total_urls"])
	}

	if stats["current_counter"] != uint64(3) {
		t.Errorf("current_counter should be 3, got %v", stats["current_counter"])
	}

	if stats["storage_type"] != "redis" {
		t.Errorf("storage_type should be 'redis', got %v", stats["storage_type"])
	}
}

func TestRedisStorage_ConcurrentAccess(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	const numGoroutines = 10
	const urlsPerGoroutine = 5

	results := make(chan *models.URLMapping, numGoroutines*urlsPerGoroutine)
	errors := make(chan error, numGoroutines*urlsPerGoroutine)

	// Start multiple goroutines storing URLs
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			for j := 0; j < urlsPerGoroutine; j++ {
				mapping := &models.URLMapping{
					LongURL: "https://www.example.com/" + string(rune('a'+workerID)) + "/" + string(rune('0'+j)),
				}
				_, err := storage.Store(mapping)
				if err != nil {
					errors <- err
					return
				}
				results <- mapping
			}
		}(i)
	}

	// Collect results
	var mappings []*models.URLMapping
	for i := 0; i < numGoroutines*urlsPerGoroutine; i++ {
		select {
		case mapping := <-results:
			mappings = append(mappings, mapping)
		case err := <-errors:
			t.Fatalf("Concurrent store failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Verify all IDs are unique
	idMap := make(map[uint64]bool)
	shortCodeMap := make(map[string]bool)

	for _, mapping := range mappings {
		if idMap[mapping.ID] {
			t.Errorf("Duplicate ID %d found in concurrent test", mapping.ID)
		}
		idMap[mapping.ID] = true

		if shortCodeMap[mapping.ShortCode] {
			t.Errorf("Duplicate ShortCode %s found in concurrent test", mapping.ShortCode)
		}
		shortCodeMap[mapping.ShortCode] = true
	}

	if len(mappings) != numGoroutines*urlsPerGoroutine {
		t.Errorf("Expected %d mappings, got %d", numGoroutines*urlsPerGoroutine, len(mappings))
	}
}

func TestRedisStorage_ConnectionFailure(t *testing.T) {
	// Test with invalid Redis URL
	_, err := NewRedisStorage("http://localhost:8080", "redis://invalid:6379")
	if err == nil {
		t.Error("NewRedisStorage should fail with invalid Redis URL")
	}
}

func TestRedisStorage_Persistence(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Store multiple URLs
	urls := []string{
		"https://www.github.com",
		"https://www.stackoverflow.com",
		"https://www.reddit.com",
	}

	var shortCodes []string
	for _, url := range urls {
		mapping := &models.URLMapping{LongURL: url}
		shortCode, err := storage.Store(mapping)
		if err != nil {
			t.Fatalf("Store() failed: %v", err)
		}
		shortCodes = append(shortCodes, shortCode)
	}

	// Verify all URLs can be retrieved
	for i, shortCode := range shortCodes {
		retrieved, err := storage.Get(shortCode)
		if err != nil {
			t.Fatalf("Get() failed for shortCode %s: %v", shortCode, err)
		}
		if retrieved.LongURL != urls[i] {
			t.Errorf("Retrieved URL %s, expected %s", retrieved.LongURL, urls[i])
		}
	}

	// Verify stats are correct
	stats := storage.GetStats()
	if stats["total_urls"] != int64(3) {
		t.Errorf("total_urls should be 3, got %v", stats["total_urls"])
	}
}

func TestRedisStorage_Close(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Test Close method
	err := storage.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// After closing, operations should fail
	mapping := &models.URLMapping{LongURL: "https://www.example.com"}
	_, err = storage.Store(mapping)
	if err == nil {
		t.Error("Store() should fail after Close()")
	}
}

func TestRedisStorage_InitCounterWithExistingValue(t *testing.T) {
	// Create mock Redis with existing counter
	mock, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mock.Close()

	// Set an existing counter value
	mock.Set("counter", "42")

	// Create Redis storage
	storage, err := NewRedisStorage("http://localhost:8080", "redis://"+mock.Addr())
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}
	defer storage.Close()

	// Counter should be initialized to the existing value
	stats := storage.GetStats()
	if stats["current_counter"] != uint64(42) {
		t.Errorf("current_counter should be 42, got %v", stats["current_counter"])
	}
}

func TestRedisStorage_StoreWithRedisFailure(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Close the mock to simulate Redis failure
	mock.Close()

	mapping := &models.URLMapping{LongURL: "https://www.example.com"}
	_, err := storage.Store(mapping)
	if err == nil {
		t.Error("Store() should fail when Redis is down")
	}
}

func TestRedisStorage_GetWithRedisFailure(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")

	// Store a URL first
	mapping := &models.URLMapping{LongURL: "https://www.example.com"}
	shortCode, err := storage.Store(mapping)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Close the mock to simulate Redis failure
	mock.Close()

	_, err = storage.Get(shortCode)
	if err == nil {
		t.Error("Get() should fail when Redis is down")
	}
}

func TestRedisStorage_GetStatsWithRedisFailure(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")

	// Close the mock to simulate Redis failure
	mock.Close()

	stats := storage.GetStats()
	
	// Should handle Redis failure gracefully - GetStats handles errors by returning 0
	if stats["total_urls"] != 0 {
		t.Errorf("total_urls should be 0 when Redis fails, got %v", stats["total_urls"])
	}
	
	// current_counter should still work (it's atomic in memory)
	if stats["storage_type"] != "redis" {
		t.Errorf("storage_type should still be 'redis', got %v", stats["storage_type"])
	}
}

func TestRedisStorage_IsExpiredMethod(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Test with nil expiration
	mapping := &models.URLMapping{
		LongURL:        "https://www.example.com",
		ExpirationDate: nil,
	}
	
	if storage.IsExpired(mapping) {
		t.Error("IsExpired() should return false for nil expiration")
	}

	// Test with future expiration
	futureTime := time.Now().Add(time.Hour)
	mapping.ExpirationDate = &futureTime
	
	if storage.IsExpired(mapping) {
		t.Error("IsExpired() should return false for future expiration")
	}

	// Test with past expiration
	pastTime := time.Now().Add(-time.Hour)
	mapping.ExpirationDate = &pastTime
	
	if !storage.IsExpired(mapping) {
		t.Error("IsExpired() should return true for past expiration")
	}
}

func TestRedisStorage_StoreExpiredURL(t *testing.T) {
	storage, mock := setupMockRedis(t, "http://localhost:8080")
	defer mock.Close()

	// Create URL with past expiration
	pastTime := time.Now().Add(-time.Hour)
	mapping := &models.URLMapping{
		LongURL:        "https://www.example.com/expired",
		ExpirationDate: &pastTime,
	}

	shortCode, err := storage.Store(mapping)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Should be able to store, but retrieving should fail because it's expired
	_, err = storage.Get(shortCode)
	if err == nil {
		t.Error("Get() should fail for expired URL")
	}
	
	// Error message should indicate expiration
	expectedError := "URL has expired: " + shortCode
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
} 
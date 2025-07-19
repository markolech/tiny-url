package storage

import (
	"sync"
	"testing"
	"time"
	"tiny-url-service/models"
)

func TestMemoryStorage_Store(t *testing.T) {
	store := NewMemoryStorage("http://localhost:8080")

	mapping := &models.URLMapping{
		LongURL: "https://www.example.com",
	}

	shortCode, err := store.Store(mapping)
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

func TestMemoryStorage_Get(t *testing.T) {
	store := NewMemoryStorage("http://localhost:8080")

	// Store a URL first
	original := &models.URLMapping{
		LongURL: "https://www.example.com/test",
	}

	shortCode, err := store.Store(original)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	// Retrieve it
	retrieved, err := store.Get(shortCode)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.LongURL != original.LongURL {
		t.Errorf("Get() returned LongURL %s, expected %s", retrieved.LongURL, original.LongURL)
	}

	if retrieved.ID != original.ID {
		t.Errorf("Get() returned ID %d, expected %d", retrieved.ID, original.ID)
	}

	if retrieved.ShortCode != original.ShortCode {
		t.Errorf("Get() returned ShortCode %s, expected %s", retrieved.ShortCode, original.ShortCode)
	}
}

func TestMemoryStorage_GetNotFound(t *testing.T) {
	store := NewMemoryStorage("http://localhost:8080")

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("Get() should return error for non-existent short code")
	}
}

func TestMemoryStorage_UniqueIDs(t *testing.T) {
	store := NewMemoryStorage("http://localhost:8080")

	var mappings []*models.URLMapping
	const numURLs = 100

	// Store multiple URLs
	for i := 0; i < numURLs; i++ {
		mapping := &models.URLMapping{
			LongURL: "https://www.example.com/test/" + string(rune(i)),
		}
		_, err := store.Store(mapping)
		if err != nil {
			t.Fatalf("Store() failed on iteration %d: %v", i, err)
		}
		mappings = append(mappings, mapping)
	}

	// Check all IDs are unique and sequential
	for i, mapping := range mappings {
		expectedID := uint64(i + 1)
		if mapping.ID != expectedID {
			t.Errorf("Mapping %d has ID %d, expected %d", i, mapping.ID, expectedID)
		}
	}

	// Check all short codes are unique
	seenCodes := make(map[string]bool)
	for i, mapping := range mappings {
		if seenCodes[mapping.ShortCode] {
			t.Errorf("Duplicate short code %s found at index %d", mapping.ShortCode, i)
		}
		seenCodes[mapping.ShortCode] = true
	}
}

func TestMemoryStorage_Expiration(t *testing.T) {
	store := NewMemoryStorage("http://localhost:8080")

	// Test URL without expiration
	mapping1 := &models.URLMapping{
		LongURL: "https://www.example.com/noexpiry",
	}
	shortCode1, err := store.Store(mapping1)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	if store.IsExpired(mapping1) {
		t.Error("URL without expiration should not be expired")
	}

	// Test URL with future expiration
	futureTime := time.Now().Add(1 * time.Hour)
	mapping2 := &models.URLMapping{
		LongURL:        "https://www.example.com/future",
		ExpirationDate: &futureTime,
	}
	shortCode2, err := store.Store(mapping2)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	if store.IsExpired(mapping2) {
		t.Error("URL with future expiration should not be expired")
	}

	// Test URL with past expiration
	pastTime := time.Now().Add(-1 * time.Hour)
	mapping3 := &models.URLMapping{
		LongURL:        "https://www.example.com/past",
		ExpirationDate: &pastTime,
	}
	shortCode3, err := store.Store(mapping3)
	if err != nil {
		t.Fatalf("Store() failed: %v", err)
	}

	if !store.IsExpired(mapping3) {
		t.Error("URL with past expiration should be expired")
	}

	// Test Get with expired URL
	_, err = store.Get(shortCode3)
	if err == nil {
		t.Error("Get() should return error for expired URL")
	}

	// Test Get with non-expired URLs
	_, err = store.Get(shortCode1)
	if err != nil {
		t.Errorf("Get() failed for non-expired URL: %v", err)
	}

	_, err = store.Get(shortCode2)
	if err != nil {
		t.Errorf("Get() failed for non-expired URL with future expiration: %v", err)
	}
}

func TestMemoryStorage_GetStats(t *testing.T) {
	store := NewMemoryStorage("http://localhost:8080")

	// Initial stats
	stats := store.GetStats()
	if stats["total_urls"] != 0 {
		t.Errorf("Initial total_urls should be 0, got %v", stats["total_urls"])
	}
	if stats["current_counter"] != uint64(0) {
		t.Errorf("Initial current_counter should be 0, got %v", stats["current_counter"])
	}
	if stats["storage_type"] != "memory" {
		t.Errorf("storage_type should be 'memory', got %v", stats["storage_type"])
	}

	// Add some URLs
	for i := 0; i < 5; i++ {
		mapping := &models.URLMapping{
			LongURL: "https://www.example.com/test/" + string(rune(i)),
		}
		_, err := store.Store(mapping)
		if err != nil {
			t.Fatalf("Store() failed: %v", err)
		}
	}

	// Check updated stats
	stats = store.GetStats()
	if stats["total_urls"] != 5 {
		t.Errorf("total_urls should be 5, got %v", stats["total_urls"])
	}
	if stats["current_counter"] != uint64(5) {
		t.Errorf("current_counter should be 5, got %v", stats["current_counter"])
	}
}

func TestMemoryStorage_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStorage("http://localhost:8080")
	
	const numGoroutines = 10
	const urlsPerGoroutine = 10
	
	var wg sync.WaitGroup
	results := make(chan *models.URLMapping, numGoroutines*urlsPerGoroutine)
	
	// Spawn multiple goroutines to store URLs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < urlsPerGoroutine; j++ {
				mapping := &models.URLMapping{
					LongURL: "https://www.example.com/concurrent/" + string(rune(goroutineID)) + "/" + string(rune(j)),
				}
				_, err := store.Store(mapping)
				if err != nil {
					t.Errorf("Store() failed in goroutine %d: %v", goroutineID, err)
					return
				}
				results <- mapping
			}
		}(i)
	}
	
	wg.Wait()
	close(results)
	
	// Collect all results
	var allMappings []*models.URLMapping
	for mapping := range results {
		allMappings = append(allMappings, mapping)
	}
	
	// Verify we got the expected number of URLs
	expectedCount := numGoroutines * urlsPerGoroutine
	if len(allMappings) != expectedCount {
		t.Errorf("Expected %d URLs, got %d", expectedCount, len(allMappings))
	}
	
	// Verify all IDs are unique
	seenIDs := make(map[uint64]bool)
	for _, mapping := range allMappings {
		if seenIDs[mapping.ID] {
			t.Errorf("Duplicate ID %d found", mapping.ID)
		}
		seenIDs[mapping.ID] = true
	}
	
	// Verify all short codes are unique
	seenCodes := make(map[string]bool)
	for _, mapping := range allMappings {
		if seenCodes[mapping.ShortCode] {
			t.Errorf("Duplicate short code %s found", mapping.ShortCode)
		}
		seenCodes[mapping.ShortCode] = true
	}
} 
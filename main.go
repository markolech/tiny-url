package main

import (
	"fmt"
	"log"
	"time"
	"tiny-url-service/models"
	"tiny-url-service/storage"
	"tiny-url-service/utils"
)

func main() {
	fmt.Println("Tiny URL Service starting...")
	
	// Test URL validation
	fmt.Println("\n=== Testing URL Validation ===")
	testURLs := []string{
		"https://www.example.com",
		"http://google.com/search?q=test",
		"invalid-url",
		"ftp://example.com",
		"",
	}
	
	for _, url := range testURLs {
		valid := utils.IsValidURL(url)
		fmt.Printf("URL: %-30s Valid: %v\n", url, valid)
	}
	
	// Test base62 encoding
	fmt.Println("\n=== Testing Base62 Encoding ===")
	testIDs := []uint64{0, 1, 61, 62, 63, 1000, 123456}
	for _, id := range testIDs {
		encoded := utils.EncodeBase62(id)
		decoded := utils.DecodeBase62(encoded)
		fmt.Printf("ID: %6d -> Encoded: %8s -> Decoded: %6d (Match: %v)\n", 
			id, encoded, decoded, id == decoded)
	}
	
	// Test storage
	fmt.Println("\n=== Testing In-Memory Storage ===")
	store := storage.NewMemoryStorage("http://localhost:8080")
	
	// Store some URLs
	mapping1 := &models.URLMapping{
		LongURL: "https://www.example.com/very/long/url/path",
	}
	
	shortCode1, err := store.Store(mapping1)
	if err != nil {
		log.Fatal("Failed to store URL:", err)
	}
	fmt.Printf("Stored URL: %s -> Short Code: %s\n", mapping1.LongURL, shortCode1)
	
	// Store URL with expiration
	expiration := time.Now().Add(24 * time.Hour)
	mapping2 := &models.URLMapping{
		LongURL:        "https://www.google.com/search?q=golang",
		ExpirationDate: &expiration,
	}
	
	shortCode2, err := store.Store(mapping2)
	if err != nil {
		log.Fatal("Failed to store URL with expiration:", err)
	}
	fmt.Printf("Stored URL with expiration: %s -> Short Code: %s\n", mapping2.LongURL, shortCode2)
	
	// Retrieve URLs
	retrieved1, err := store.Get(shortCode1)
	if err != nil {
		log.Fatal("Failed to retrieve URL:", err)
	}
	fmt.Printf("Retrieved: %s -> %s\n", shortCode1, retrieved1.LongURL)
	
	retrieved2, err := store.Get(shortCode2)
	if err != nil {
		log.Fatal("Failed to retrieve URL with expiration:", err)
	}
	fmt.Printf("Retrieved: %s -> %s (Expires: %v)\n", 
		shortCode2, retrieved2.LongURL, retrieved2.ExpirationDate.Format(time.RFC3339))
	
	// Get stats
	stats := store.GetStats()
	fmt.Printf("\n=== Storage Stats ===\n")
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}
	
	fmt.Println("\n✅ Phase 2 core logic tests completed successfully!")
} 
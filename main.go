package main

import (
	"log"
	"tiny-url-service/handlers"
	"tiny-url-service/storage"
)

func main() {
	// Configuration
	baseURL := "http://localhost:8080"
	port := 8080
	
	// Initialize storage
	store := storage.NewMemoryStorage(baseURL)
	
	// Start HTTP server
	log.Println("Initializing Tiny URL Service...")
	if err := handlers.StartServer(store, baseURL, port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 
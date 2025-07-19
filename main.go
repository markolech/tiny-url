package main

import (
	"log"
	"tiny-url-service/config"
	"tiny-url-service/handlers"
	"tiny-url-service/storage"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()
	
	// Initialize storage
	store := storage.NewMemoryStorage(cfg.BaseURL)
	
	// Start HTTP server with graceful shutdown
	log.Println("Initializing Tiny URL Service...")
	if err := handlers.StartServer(store, cfg); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 
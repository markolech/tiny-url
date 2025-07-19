package main

import (
	"log"
	"strings"
	"tiny-url-service/config"
	"tiny-url-service/handlers"
	"tiny-url-service/storage"
)

func main() {
	// Load configuration from environment variables
	cfg := config.Load()
	
	// Initialize storage based on configuration
	var store storage.Storage
	var err error
	
	switch strings.ToLower(cfg.StorageType) {
	case "redis":
		log.Println("Initializing Redis storage...")
		store, err = storage.NewRedisStorage(cfg.BaseURL, cfg.RedisURL)
		if err != nil {
			log.Fatal("Failed to initialize Redis storage:", err)
		}
		log.Println("Redis storage initialized successfully")
	case "memory":
		log.Println("Initializing in-memory storage...")
		store = storage.NewMemoryStorage(cfg.BaseURL)
		log.Println("In-memory storage initialized successfully")
	default:
		log.Fatalf("Unknown storage type: %s. Supported types: memory, redis", cfg.StorageType)
	}
	
	// Start HTTP server with graceful shutdown
	log.Println("Starting Tiny URL Service...")
	if err := handlers.StartServer(store, cfg); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 
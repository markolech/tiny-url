package handlers

import (
	"fmt"
	"log"
	"tiny-url-service/storage"

	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the Gin router with all routes and middleware
func SetupRouter(store storage.Storage, baseURL string) *gin.Engine {
	// Set Gin to release mode for production (can be overridden with GIN_MODE env var)
	gin.SetMode(gin.ReleaseMode)
	
	// Create Gin router
	r := gin.New()
	
	// Add middleware
	r.Use(gin.Logger())    // Request logging
	r.Use(gin.Recovery())  // Panic recovery
	r.Use(CORSMiddleware()) // CORS headers
	
	// Create handlers instance
	handlers := NewURLHandlers(store, baseURL)
	
	// Setup routes
	r.POST("/urls", handlers.CreateShortURL)
	r.GET("/:shortCode", handlers.RedirectToLongURL)
	r.GET("/urls/:shortCode/stats", handlers.GetURLStats)
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		stats := store.GetStats()
		c.JSON(200, gin.H{
			"status": "healthy",
			"stats":  stats,
		})
	})
	
	return r
}

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// StartServer starts the HTTP server on the specified port
func StartServer(store storage.Storage, baseURL string, port int) error {
	router := SetupRouter(store, baseURL)
	
	address := fmt.Sprintf(":%d", port)
	log.Printf("🚀 Tiny URL service starting on %s", address)
	log.Printf("📊 Health check available at: %s/health", baseURL)
	log.Printf("📝 API documentation:")
	log.Printf("   POST %s/urls - Create short URL", baseURL)
	log.Printf("   GET  %s/{shortCode} - Redirect to long URL", baseURL)
	log.Printf("   GET  %s/urls/{shortCode}/stats - Get URL stats", baseURL)
	
	return router.Run(address)
} 
package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tiny-url-service/config"
	"tiny-url-service/middleware"
	"tiny-url-service/storage"

	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the Gin router with all routes and middleware
func SetupRouter(store storage.Storage, cfg *config.Config) *gin.Engine {
	// Set Gin mode from configuration
	gin.SetMode(cfg.GinMode)
	
	// Create Gin router
	r := gin.New()
	
	// Add middleware
	r.Use(gin.Logger())           // Request logging
	r.Use(gin.Recovery())         // Panic recovery
	r.Use(CORSMiddleware())       // CORS headers
	r.Use(ContentTypeMiddleware()) // Content-Type validation
	r.Use(middleware.NewInMemoryRateLimiter()) // Rate limiting
	
	// Create handlers instance
	handlers := NewURLHandlers(store, cfg.BaseURL)
	
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

// ContentTypeMiddleware validates Content-Type for POST requests
func ContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only validate Content-Type for POST requests
		if c.Request.Method == "POST" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
				c.JSON(400, gin.H{
					"error": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// StartServer starts the HTTP server with proper configuration, timeouts, and graceful shutdown
func StartServer(store storage.Storage, cfg *config.Config) error {
	router := SetupRouter(store, cfg)
	
	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}
	
	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Tiny URL service starting on :%d", cfg.Port)
		log.Printf("📊 Health check available at: %s/health", cfg.BaseURL)
		log.Printf("📝 API documentation:")
		log.Printf("   POST %s/urls - Create short URL", cfg.BaseURL)
		log.Printf("   GET  %s/{shortCode} - Redirect to long URL", cfg.BaseURL)
		log.Printf("   GET  %s/urls/{shortCode}/stats - Get URL stats", cfg.BaseURL)
		log.Printf("⚙️  Configuration:")
		log.Printf("   Mode: %s", cfg.GinMode)
		log.Printf("   Read timeout: %v", cfg.ReadTimeout)
		log.Printf("   Write timeout: %v", cfg.WriteTimeout)
		log.Printf("   Idle timeout: %v", cfg.IdleTimeout)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	<-quit
	log.Println("🛑 Shutting down server...")
	
	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	
	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
		return err
	}
	
	log.Println("✅ Server exited gracefully")
	return nil
} 
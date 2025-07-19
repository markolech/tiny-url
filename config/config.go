package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Port           int
	BaseURL        string
	GinMode        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	ShutdownTimeout time.Duration
	
	// Storage configuration
	StorageType string // "memory" or "redis"
	RedisURL    string // Redis connection URL
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Port:            getEnvAsInt("PORT", 8080),
		BaseURL:         getEnv("BASE_URL", "http://localhost:8080"),
		GinMode:         getEnv("GIN_MODE", "release"),
		ReadTimeout:     getEnvAsDuration("READ_TIMEOUT", "10s"),
		WriteTimeout:    getEnvAsDuration("WRITE_TIMEOUT", "10s"),
		IdleTimeout:     getEnvAsDuration("IDLE_TIMEOUT", "60s"),
		ShutdownTimeout: getEnvAsDuration("SHUTDOWN_TIMEOUT", "30s"),
		
		// Storage configuration
		StorageType:     getEnv("STORAGE_TYPE", "memory"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
	}
}

// getEnv gets an environment variable with a fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a fallback default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsDuration gets an environment variable as duration with a fallback default
func getEnvAsDuration(key, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return 10 * time.Second // fallback if parsing fails
} 
package models

import "time"

// URLMapping represents a mapping between a short code and a long URL
type URLMapping struct {
	ID             uint64     `json:"id"`
	ShortCode      string     `json:"short_code"`
	LongURL        string     `json:"long_url"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"` // Optional expiration
	CreatedAt      time.Time  `json:"created_at"`
}

// ShortenRequest represents the request payload for creating a short URL
type ShortenRequest struct {
	LongURL        string     `json:"long_url" binding:"required"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// ShortenResponse represents the response for a successful URL shortening
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
} 
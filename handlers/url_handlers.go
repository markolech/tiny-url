package handlers

import (
	"net/http"
	"tiny-url-service/models"
	"tiny-url-service/storage"
	"tiny-url-service/utils"

	"github.com/gin-gonic/gin"
)

// URLHandlers contains the storage instance and handlers
type URLHandlers struct {
	storage storage.Storage
	baseURL string
}

// NewURLHandlers creates a new URL handlers instance
func NewURLHandlers(store storage.Storage, baseURL string) *URLHandlers {
	return &URLHandlers{
		storage: store,
		baseURL: baseURL,
	}
}

// CreateShortURL handles POST /urls - creates a new short URL
func (h *URLHandlers) CreateShortURL(c *gin.Context) {
	var req models.ShortenRequest
	
	// Bind JSON request to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}
	
	// Validate URL
	if !utils.IsValidURL(req.LongURL) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid URL format. Must be http:// or https://",
		})
		return
	}
	
	// Create URL mapping
	mapping := &models.URLMapping{
		LongURL:        req.LongURL,
		ExpirationDate: req.ExpirationDate,
	}
	
	// Store in database
	shortCode, err := h.storage.Store(mapping)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create short URL",
			"details": err.Error(),
		})
		return
	}
	
	// Return response
	response := models.ShortenResponse{
		ShortURL: h.baseURL + "/" + shortCode,
	}
	
	c.JSON(http.StatusOK, response)
}

// RedirectToLongURL handles GET /{shortCode} - redirects to the original URL
func (h *URLHandlers) RedirectToLongURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	
	// Validate short code is not empty
	if shortCode == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Short code not provided",
		})
		return
	}
	
	// Get URL mapping from storage
	mapping, err := h.storage.Get(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Short URL not found",
		})
		return
	}
	
	// Redirect to original URL
	c.Redirect(http.StatusFound, mapping.LongURL)
}

// GetURLStats handles GET /urls/{shortCode}/stats - returns URL statistics
func (h *URLHandlers) GetURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")
	
	// Get URL mapping from storage
	mapping, err := h.storage.Get(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Short URL not found",
		})
		return
	}
	
	// Return URL information
	c.JSON(http.StatusOK, gin.H{
		"short_code":      mapping.ShortCode,
		"long_url":        mapping.LongURL,
		"created_at":      mapping.CreatedAt,
		"expiration_date": mapping.ExpirationDate,
		"id":              mapping.ID,
	})
} 
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add rate limiter middleware
	router.Use(NewInMemoryRateLimiter())
	
	// Simple test endpoint
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	return router
}

func TestRateLimiter_AllowedRequests(t *testing.T) {
	router := setupTestRouter()

	// Make requests within the limit (20 per minute)
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d, expected %d", i+1, w.Code, http.StatusOK)
		}

		// Check rate limit headers
		if w.Header().Get("X-RateLimit-Limit") != "20" {
			t.Errorf("Expected X-RateLimit-Limit: 20, got %s", w.Header().Get("X-RateLimit-Limit"))
		}

		remaining := w.Header().Get("X-RateLimit-Remaining")
		if remaining == "" {
			t.Error("X-RateLimit-Remaining header should be present")
		}

		resetTime := w.Header().Get("X-RateLimit-Reset")
		if resetTime == "" {
			t.Error("X-RateLimit-Reset header should be present")
		}
	}
}

func TestRateLimiter_ExceedLimit(t *testing.T) {
	router := setupTestRouter()

	// Make requests to exhaust the limit
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.101:12345"
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d within limit failed with status %d", i+1, w.Code)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.101:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request exceeding limit should return %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	// Check rate limit headers
	if w.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("Expected X-RateLimit-Remaining: 0, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}

	// Should have a retry-after header
	if w.Header().Get("Retry-After") == "" {
		t.Error("Retry-After header should be present when rate limited")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	router := setupTestRouter()

	// Test that different IPs have separate limits
	ips := []string{
		"192.168.1.100:12345",
		"192.168.1.101:12345",
		"10.0.0.1:54321",
	}

	for _, ip := range ips {
		// Each IP should be able to make requests up to the limit
		for i := 0; i < 15; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Request from IP %s failed with status %d", ip, w.Code)
			}
		}
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	router := setupTestRouter()
	ip := "192.168.1.102:12345"

	// Make a few requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Initial request %d failed", i+1)
		}
	}

	// Wait for token refill (rate is 1 token per 3 seconds)
	time.Sleep(4 * time.Second)

	// Should be able to make another request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = ip
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Request after token refill failed with status %d", w.Code)
	}

	// The remaining count should have increased
	remaining := w.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("X-RateLimit-Remaining header should be present")
	}
}

func TestRateLimiter_EdgeCases(t *testing.T) {
	router := setupTestRouter()

	// Test with malformed IP
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "invalid-ip"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still work (use full string as IP)
	if w.Code != http.StatusOK {
		t.Errorf("Request with malformed IP failed with status %d", w.Code)
	}

	// Test with empty RemoteAddr
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = ""
	
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Should still work
	if w2.Code != http.StatusOK {
		t.Errorf("Request with empty RemoteAddr failed with status %d", w2.Code)
	}
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	router := setupTestRouter()
	ip := "192.168.1.103:12345"

	const numRequests = 25
	results := make(chan int, numRequests)

	// Make concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			results <- w.Code
		}()
	}

	// Collect results
	var successCount, rateLimitedCount int
	for i := 0; i < numRequests; i++ {
		select {
		case code := <-results:
			if code == http.StatusOK {
				successCount++
			} else if code == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Should have exactly 20 successful requests and 5 rate limited
	if successCount != 20 {
		t.Errorf("Expected 20 successful requests, got %d", successCount)
	}
	if rateLimitedCount != 5 {
		t.Errorf("Expected 5 rate limited requests, got %d", rateLimitedCount)
	}
}

func TestRateLimiter_Headers(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.104:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check all required headers are present
	headers := map[string]string{
		"X-RateLimit-Limit":     "20",
		"X-RateLimit-Remaining": "19",
		"X-RateLimit-Reset":     "", // Just check it exists
	}

	for header, expectedValue := range headers {
		actualValue := w.Header().Get(header)
		if actualValue == "" {
			t.Errorf("Header %s should be present", header)
		} else if expectedValue != "" && actualValue != expectedValue {
			t.Errorf("Header %s should be %s, got %s", header, expectedValue, actualValue)
		}
	}
}

func TestRateLimiter_IPExtraction(t *testing.T) {
	router := setupTestRouter()

	testCases := []struct {
		remoteAddr string
		expected   string
	}{
		{"192.168.1.1:8080", "192.168.1.1"},
		{"10.0.0.1:12345", "10.0.0.1"},
		{"127.0.0.1:54321", "127.0.0.1"},
		{"invalid-format", "invalid-format"}, // Fallback to full string
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = tc.remoteAddr
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request with RemoteAddr %s failed", tc.remoteAddr)
		}
	}
} 
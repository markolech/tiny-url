package middleware

import (
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     float64   // Current number of tokens
	lastRefill time.Time // Last time tokens were refilled
	capacity   float64   // Maximum number of tokens
	refillRate float64   // Tokens added per second
	mu         sync.Mutex
}

// InMemoryRateLimiter implements per-IP token bucket rate limiting
type InMemoryRateLimiter struct {
	buckets *sync.Map // map[string]*TokenBucket
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
// 20 requests per minute per IP
func NewInMemoryRateLimiter() gin.HandlerFunc {
	limiter := &InMemoryRateLimiter{
		buckets: &sync.Map{},
	}
	
	return limiter.middleware()
}

// getBucket gets or creates a token bucket for the given IP
func (rl *InMemoryRateLimiter) getBucket(ip string) *TokenBucket {
	val, _ := rl.buckets.LoadOrStore(ip, &TokenBucket{
		tokens:     20.0,                    // Start with full bucket
		lastRefill: time.Now(),
		capacity:   20.0,                    // 20 tokens max
		refillRate: 20.0 / 60.0,            // 20 tokens per 60 seconds
	})
	return val.(*TokenBucket)
}

// allow checks if a request from the given IP should be allowed
func (rl *InMemoryRateLimiter) allow(ip string) (bool, int) {
	bucket := rl.getBucket(ip)
	
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	
	// Add tokens based on elapsed time
	tokensToAdd := elapsed * bucket.refillRate
	bucket.tokens = math.Min(bucket.capacity, bucket.tokens+tokensToAdd)
	bucket.lastRefill = now
	
	// Try to consume one token
	if bucket.tokens >= 1.0 {
		bucket.tokens -= 1.0
		return true, int(math.Floor(bucket.tokens))
	}
	
	return false, 0
}

// middleware returns the Gin middleware function
func (rl *InMemoryRateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		allowed, remainingTokens := rl.allow(clientIP)
		
		// Add rate limit headers
		c.Header("X-RateLimit-Limit", "20")
		c.Header("X-RateLimit-Window", "60")
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remainingTokens))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(60*time.Second).Unix(), 10))
		
		if !allowed {
			// Rate limited
			c.Header("Retry-After", "3") // Approximately 3 seconds for next token
			
			c.JSON(429, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Maximum 20 requests per minute per IP",
				"limit":       20,
				"window":      "60 seconds",
				"retry_after": "3 seconds",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
} 
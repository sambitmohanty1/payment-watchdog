package mediators

import (
	"sync"
	"time"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	provider         string
	requestsPerMinute int
	burstSize         int
	retryAfter        time.Duration
	
	// Token bucket implementation
	tokens           int
	lastRefill       time.Time
	refillRate       float64
	mutex            sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	if config == nil {
		config = &RateLimitConfig{
			RequestsPerMinute: 100,
			BurstSize:         10,
			RetryAfter:        1 * time.Minute,
		}
	}
	
	limiter := &RateLimiter{
		provider:         "unknown",
		requestsPerMinute: config.RequestsPerMinute,
		burstSize:         config.BurstSize,
		retryAfter:        config.RetryAfter,
		tokens:           config.BurstSize,
		lastRefill:       time.Now(),
		refillRate:       float64(config.RequestsPerMinute) / 60.0, // tokens per second
	}
	
	return limiter
}

// Wait blocks until a token is available
func (r *RateLimiter) Wait() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Refill tokens based on time elapsed
	r.refillTokens()
	
	// If no tokens available, wait
	if r.tokens <= 0 {
		// Calculate wait time
		waitTime := time.Duration(float64(time.Second) / r.refillRate)
		r.mutex.Unlock()
		time.Sleep(waitTime)
		r.mutex.Lock()
		r.refillTokens()
	}
	
	// Consume token
	r.tokens--
}

// TryAcquire attempts to acquire a token without blocking
func (r *RateLimiter) TryAcquire() bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.refillTokens()
	
	if r.tokens > 0 {
		r.tokens--
		return true
	}
	
	return false
}

// refillTokens refills tokens based on time elapsed
func (r *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	
	// Calculate tokens to add
	tokensToAdd := int(elapsed.Seconds() * r.refillRate)
	
	if tokensToAdd > 0 {
		r.tokens = min(r.tokens+tokensToAdd, r.burstSize)
		r.lastRefill = now
	}
}

// GetRateLimitInfo returns current rate limit information
func (r *RateLimiter) GetRateLimitInfo() map[string]interface{} {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.refillTokens()
	
	// Calculate reset time
	timeToNextToken := time.Duration(float64(time.Second) / r.refillRate)
	resetTime := time.Now().Add(timeToNextToken)
	
	return map[string]interface{}{
		"provider_id":         r.provider,
		"requests_remaining":  r.tokens,
		"reset_time":         resetTime,
		"limit":              r.burstSize,
	}
}

// SetProvider sets the provider identifier
func (r *RateLimiter) SetProvider(provider string) {
	r.provider = provider
}

// UpdateConfig updates the rate limiter configuration
func (r *RateLimiter) UpdateConfig(config *RateLimitConfig) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.requestsPerMinute = config.RequestsPerMinute
	r.burstSize = config.BurstSize
	r.retryAfter = config.RetryAfter
	r.refillRate = float64(config.RequestsPerMinute) / 60.0
	
	// Reset tokens to burst size
	r.tokens = r.burstSize
	r.lastRefill = time.Now()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

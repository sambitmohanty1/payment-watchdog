package mediators

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		RetryAfter:        1 * time.Second,
	}

	limiter := NewRateLimiter(config)

	assert.NotNil(t, limiter)
	
	// Test that the rate limiter is working by checking its behavior
	info := limiter.GetRateLimitInfo()
	assert.NotNil(t, info)
	assert.Contains(t, info, "requests_remaining")
	assert.Contains(t, info, "limit")
}

func TestRateLimiter_Wait(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		RetryAfter:        1 * time.Second,
	}

	limiter := NewRateLimiter(config)

	// Test that we can consume tokens
	start := time.Now()
	limiter.Wait()
	duration := time.Since(start)

	// Should be very fast (no waiting)
	assert.Less(t, duration, 100*time.Millisecond)
}

func TestRateLimiter_Wait_RateLimited(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         1,
		RetryAfter:        100 * time.Millisecond,
	}

	limiter := NewRateLimiter(config)

	// First request should be fast
	start := time.Now()
	limiter.Wait()
	duration := time.Since(start)
	assert.Less(t, duration, 10*time.Millisecond)

	// Second request should be rate limited
	start = time.Now()
	limiter.Wait()
	duration = time.Since(start)
	assert.GreaterOrEqual(t, duration, 50*time.Millisecond) // Allow some tolerance
}

func TestRateLimiter_GetRateLimitInfo(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		RetryAfter:        1 * time.Second,
	}

	limiter := NewRateLimiter(config)

	info := limiter.GetRateLimitInfo()
	assert.NotNil(t, info)
	assert.Contains(t, info, "requests_remaining")
	assert.Contains(t, info, "limit")
	assert.Contains(t, info, "reset_time")
}

func TestRateLimiter_TryAcquire(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         2,
		RetryAfter:        1 * time.Second,
	}

	limiter := NewRateLimiter(config)

	// First two attempts should succeed
	assert.True(t, limiter.TryAcquire())
	assert.True(t, limiter.TryAcquire())
	
	// Third attempt should fail
	assert.False(t, limiter.TryAcquire())
}

func TestRateLimiter_SetProvider(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		RetryAfter:        1 * time.Second,
	}

	limiter := NewRateLimiter(config)
	
	// Set provider
	limiter.SetProvider("stripe")
	
	// Check that provider was set
	info := limiter.GetRateLimitInfo()
	assert.Equal(t, "stripe", info["provider_id"])
}

func TestRateLimiter_UpdateConfig(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		RetryAfter:        1 * time.Second,
	}

	limiter := NewRateLimiter(config)
	
	// Update config
	newConfig := &RateLimitConfig{
		RequestsPerMinute: 120,
		BurstSize:         20,
		RetryAfter:        2 * time.Second,
	}
	
	limiter.UpdateConfig(newConfig)
	
	// Check that config was updated
	info := limiter.GetRateLimitInfo()
	assert.Equal(t, 20, info["limit"])
}

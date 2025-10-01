package ratelimit

import (
	"context"
	"sync"
	"time"
)

// SimpleRateLimiter provides a simple rate limiting implementation
type SimpleRateLimiter struct {
	limiters map[string]*TokenBucket
	mu       sync.RWMutex
	rate     int
	burst    int
}

// TokenBucket implements a simple token bucket algorithm
type TokenBucket struct {
	tokens     int
	capacity   int
	lastRefill time.Time
	rate       int
	mu         sync.Mutex
}

// NewSimpleRateLimiter creates a new simple rate limiter
func NewSimpleRateLimiter(rate, burst int) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		limiters: make(map[string]*TokenBucket),
		rate:     rate,
		burst:    burst,
	}
}

// AllowIP checks if an IP is allowed to make a request
func (r *SimpleRateLimiter) AllowIP(ctx context.Context, ip string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, exists := r.limiters[ip]
	if !exists {
		limiter = &TokenBucket{
			tokens:     r.burst,
			capacity:   r.burst,
			lastRefill: time.Now(),
			rate:       r.rate,
		}
		r.limiters[ip] = limiter
	}

	return limiter.Allow(), nil
}

// GetLimiter gets the rate limiter for an IP
func (r *SimpleRateLimiter) GetLimiter(ip string) *TokenBucket {
	r.mu.Lock()
	defer r.mu.Unlock()

	limiter, exists := r.limiters[ip]
	if !exists {
		limiter = &TokenBucket{
			tokens:     r.burst,
			capacity:   r.burst,
			lastRefill: time.Now(),
			rate:       r.rate,
		}
		r.limiters[ip] = limiter
	}

	return limiter
}

// Cleanup removes old limiters to prevent memory leaks
func (r *SimpleRateLimiter) Cleanup(maxAge time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for ip := range r.limiters {
		// Simple cleanup based on time
		if time.Since(now) > maxAge {
			delete(r.limiters, ip)
		}
	}
}

// GetStats returns rate limiter statistics
func (r *SimpleRateLimiter) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"total_limiters": len(r.limiters),
		"rate":           r.rate,
		"burst":          r.burst,
	}
}

// Allow checks if a request is allowed
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Add tokens based on elapsed time
	tokensToAdd := int(elapsed.Seconds()) * tb.rate
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

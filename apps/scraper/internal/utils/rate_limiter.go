package utils

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu           sync.Mutex
	tokens       int
	maxTokens    int
	refillRate   int
	lastRefill   time.Time
	refillPeriod time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, refillPeriod time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:       maxTokens,
		maxTokens:    maxTokens,
		refillRate:   1,
		lastRefill:   time.Now(),
		refillPeriod: refillPeriod,
	}
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if rl.tryAcquire() {
			return nil
		}

		// Wait a bit before trying again
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// tryAcquire attempts to acquire a token
func (rl *RateLimiter) tryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens if needed
	rl.refill()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// refill refills tokens based on time elapsed
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.refillPeriod {
		refillAmount := int(elapsed/rl.refillPeriod) * rl.refillRate
		rl.tokens = min(rl.maxTokens, rl.tokens+refillAmount)
		rl.lastRefill = now
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetStats returns rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return map[string]interface{}{
		"tokens_available": rl.tokens,
		"max_tokens":       rl.maxTokens,
		"refill_rate":      rl.refillRate,
		"refill_period":    rl.refillPeriod,
		"last_refill":      rl.lastRefill,
	}
}

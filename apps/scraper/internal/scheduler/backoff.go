package scheduler

import (
	"math"
	"time"
)

// BackoffStrategy defines the interface for backoff strategies
type BackoffStrategy interface {
	GetDelay(attempt int) time.Duration
	Reset()
}

// NoBackoff implements a strategy with no delay
type NoBackoff struct{}

func (n *NoBackoff) GetDelay(attempt int) time.Duration {
	return 0
}

func (n *NoBackoff) Reset() {
	// No state to reset
}

// LinearBackoff implements a linear backoff strategy
type LinearBackoff struct {
	baseDelay time.Duration
}

func NewLinearBackoff(baseDelay time.Duration) *LinearBackoff {
	return &LinearBackoff{baseDelay: baseDelay}
}

func (l *LinearBackoff) GetDelay(attempt int) time.Duration {
	return time.Duration(attempt) * l.baseDelay
}

func (l *LinearBackoff) Reset() {
	// No state to reset
}

// ExponentialBackoff implements an exponential backoff strategy
type ExponentialBackoff struct {
	baseDelay time.Duration
	maxDelay  time.Duration
}

func NewExponentialBackoff(baseDelay, maxDelay time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
	}
}

func (e *ExponentialBackoff) GetDelay(attempt int) time.Duration {
	delay := e.baseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
	if delay > e.maxDelay {
		delay = e.maxDelay
	}
	return delay
}

func (e *ExponentialBackoff) Reset() {
	// No state to reset
}

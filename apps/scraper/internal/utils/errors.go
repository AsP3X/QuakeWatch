package utils

import (
	"context"
	"fmt"
	"time"
)

// ErrorType represents the type of error
type ErrorType int

const (
	ErrorTypeNetwork ErrorType = iota
	ErrorTypeAPI
	ErrorTypeValidation
	ErrorTypeStorage
	ErrorTypeConfiguration
	ErrorTypeTimeout
	ErrorTypeRateLimit
	ErrorTypeUnknown
)

// CollectionError represents a structured error with context
type CollectionError struct {
	Type      ErrorType
	Source    string
	Message   string
	Retryable bool
	Context   map[string]interface{}
	Timestamp time.Time
	Original  error
}

// Error implements the error interface
func (e *CollectionError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Source, e.Type.String(), e.Message)
}

// Unwrap returns the original error
func (e *CollectionError) Unwrap() error {
	return e.Original
}

// String returns the string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeNetwork:
		return "NETWORK"
	case ErrorTypeAPI:
		return "API"
	case ErrorTypeValidation:
		return "VALIDATION"
	case ErrorTypeStorage:
		return "STORAGE"
	case ErrorTypeConfiguration:
		return "CONFIGURATION"
	case ErrorTypeTimeout:
		return "TIMEOUT"
	case ErrorTypeRateLimit:
		return "RATE_LIMIT"
	case ErrorTypeUnknown:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

// NewCollectionError creates a new collection error
func NewCollectionError(errType ErrorType, source, message string, retryable bool, original error) *CollectionError {
	return &CollectionError{
		Type:      errType,
		Source:    source,
		Message:   message,
		Retryable: retryable,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Original:  original,
	}
}

// RetryStrategy defines the retry behavior
type RetryStrategy struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	Jitter            bool
}

// DefaultRetryStrategy returns a default retry strategy
func DefaultRetryStrategy() *RetryStrategy {
	return &RetryStrategy{
		MaxAttempts:       3,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            true,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry
func RetryWithBackoff(ctx context.Context, strategy *RetryStrategy, fn func() error) error {
	var lastErr error
	delay := strategy.InitialDelay

	for attempt := 0; attempt < strategy.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err

			// Check if error is retryable
			if collectionErr, ok := err.(*CollectionError); ok && !collectionErr.Retryable {
				return err
			}
		}

		// Don't sleep on the last attempt
		if attempt == strategy.MaxAttempts-1 {
			break
		}

		// Calculate next delay
		nextDelay := time.Duration(float64(delay) * strategy.BackoffMultiplier)
		if nextDelay > strategy.MaxDelay {
			nextDelay = strategy.MaxDelay
		}

		// Add jitter if enabled
		if strategy.Jitter {
			nextDelay = addJitter(nextDelay)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay = nextDelay
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", strategy.MaxAttempts, lastErr)
}

// addJitter adds random jitter to the delay to prevent thundering herd
func addJitter(delay time.Duration) time.Duration {
	jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
	return delay + time.Duration(time.Now().UnixNano()%int64(jitter))
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if collectionErr, ok := err.(*CollectionError); ok {
		return collectionErr.Retryable
	}

	// Default retryable errors
	switch err.Error() {
	case "context deadline exceeded", "context canceled":
		return false
	default:
		return true
	}
}

// ErrorContext adds context to an error
func ErrorContext(err error, key string, value interface{}) error {
	if collectionErr, ok := err.(*CollectionError); ok {
		collectionErr.Context[key] = value
		return collectionErr
	}
	return err
}

package utils

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the current state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for improved resilience
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitBreakerState
	failures         int
	threshold        int
	timeout          time.Duration
	lastFailure      time.Time
	successes        int
	successThreshold int
}

// NewCircuitBreaker creates a new circuit breaker with the specified configuration
func NewCircuitBreaker(threshold int, timeout time.Duration, successThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		threshold:        threshold,
		timeout:          timeout,
		successThreshold: successThreshold,
	}
}

// Execute runs the provided function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.Ready() {
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	cb.RecordResult(err)
	return err
}

// Ready checks if the circuit breaker is ready to execute operations
func (cb *CircuitBreaker) Ready() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordResult records the result of an operation
func (cb *CircuitBreaker) RecordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailure = time.Now()
	cb.successes = 0

	if cb.state == StateHalfOpen || (cb.state == StateClosed && cb.failures >= cb.threshold) {
		cb.state = StateOpen
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.failures = 0
	cb.successes++

	if cb.state == StateHalfOpen && cb.successes >= cb.successThreshold {
		cb.state = StateClosed
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":       cb.state,
		"failures":    cb.failures,
		"successes":   cb.successes,
		"lastFailure": cb.lastFailure,
		"threshold":   cb.threshold,
		"timeout":     cb.timeout,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.lastFailure = time.Time{}
}

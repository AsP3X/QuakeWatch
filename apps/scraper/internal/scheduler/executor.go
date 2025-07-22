package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"
)

// CommandExecutor handles the execution of CLI commands with retry logic
type CommandExecutor struct {
	backoff    BackoffStrategy
	logger     *log.Logger
	retryCount int
	executor   func(ctx context.Context, args []string) error
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(logger *log.Logger) *CommandExecutor {
	return &CommandExecutor{
		backoff:    NewExponentialBackoff(5*time.Second, 30*time.Second),
		logger:     logger,
		retryCount: 3,
	}
}

// NewCommandExecutorWithFunction creates a new command executor with a custom execution function
func NewCommandExecutorWithFunction(logger *log.Logger, executor func(ctx context.Context, args []string) error) *CommandExecutor {
	return &CommandExecutor{
		backoff:    NewExponentialBackoff(5*time.Second, 30*time.Second),
		logger:     logger,
		retryCount: 3,
		executor:   executor,
	}
}

// Execute runs a command once
func (e *CommandExecutor) Execute(ctx context.Context, command string, args []string) error {
	e.logger.Printf("Executing command: %s %v", command, args)

	if e.executor != nil {
		// Use internal executor function
		return e.executor(ctx, args)
	}

	// Fallback to external command execution (for backward compatibility)
	return fmt.Errorf("no executor function provided")
}

// ExecuteWithRetry runs a command with retry logic and backoff
func (e *CommandExecutor) ExecuteWithRetry(ctx context.Context, command string, args []string) error {
	var lastErr error

	for attempt := 0; attempt <= e.retryCount; attempt++ {
		if attempt > 0 {
			delay := e.backoff.GetDelay(attempt)
			e.logger.Printf("Retry attempt %d after %v delay", attempt, delay)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		if err := e.Execute(ctx, command, args); err != nil {
			lastErr = err
			e.logger.Printf("Command execution failed (attempt %d): %v", attempt+1, err)

			// Don't retry on context cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Continue to next attempt if we haven't exhausted retries
			if attempt < e.retryCount {
				continue
			}
		} else {
			// Success, reset backoff
			e.backoff.Reset()
			return nil
		}
	}

	return fmt.Errorf("command failed after %d attempts: %w", e.retryCount+1, lastErr)
}

// SetRetryCount sets the number of retry attempts
func (e *CommandExecutor) SetRetryCount(count int) {
	e.retryCount = count
}

// SetBackoffStrategy sets the backoff strategy
func (e *CommandExecutor) SetBackoffStrategy(strategy BackoffStrategy) {
	e.backoff = strategy
}

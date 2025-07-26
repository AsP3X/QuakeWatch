package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"quakewatch-scraper/internal/config"
)

// IntervalScheduler manages the execution of commands at specified intervals
type IntervalScheduler struct {
	config    *config.IntervalConfig
	executor  *CommandExecutor
	logger    *log.Logger
	stopChan  chan struct{}
	doneChan  chan struct{}
	daemon    DaemonManager
	metrics   *Metrics
	mu        sync.RWMutex
	isRunning bool
}

// NewIntervalScheduler creates a new interval scheduler
func NewIntervalScheduler(cfg *config.IntervalConfig, logger *log.Logger) *IntervalScheduler {
	return &IntervalScheduler{
		config:   cfg,
		executor: NewCommandExecutor(logger),
		logger:   logger,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		daemon: NewDaemonManager(DaemonConfig{
			PIDFile: cfg.PIDFile,
			LogFile: cfg.LogFile,
			Logger:  logger,
		}),
		metrics: NewMetrics(),
	}
}

// Start begins the interval execution of the specified command
func (s *IntervalScheduler) Start(ctx context.Context, command string, args []string) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	s.logger.Printf("Starting interval scheduler with command: %s", command)
	s.logger.Printf("Interval: %v, Max Runtime: %v, Max Executions: %d",
		s.config.DefaultInterval, s.config.MaxRuntime, s.config.MaxExecutions)

	// Create context with timeout if max runtime is specified
	var cancel context.CancelFunc
	if s.config.MaxRuntime > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.config.MaxRuntime)
		defer cancel()
	}

	// Start health monitoring if enabled
	if s.config.HealthCheckInterval > 0 {
		healthMonitor := NewHealthMonitor(s.config.HealthCheckInterval, s.logger, s.metrics)
		go healthMonitor.Start(ctx)
	}

	executionCount := 0
	ticker := time.NewTicker(s.config.DefaultInterval)
	defer ticker.Stop()

	// Execute immediately on start
	if err := s.executeCommand(ctx, command, args, executionCount); err != nil {
		s.logger.Printf("Initial execution failed: %v", err)
		if !s.config.ContinueOnError {
			s.mu.Lock()
			s.isRunning = false
			s.mu.Unlock()
			return err
		}
	}
	executionCount++

	// Main execution loop
	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("Context cancelled, stopping scheduler")
			return ctx.Err()

		case <-s.stopChan:
			s.logger.Printf("Stop signal received, stopping scheduler")
			return nil

		case <-ticker.C:
			// Check if we've reached the maximum number of executions
			if s.config.MaxExecutions > 0 && executionCount >= s.config.MaxExecutions {
				s.logger.Printf("Reached maximum executions (%d), stopping scheduler", s.config.MaxExecutions)
				return nil
			}

			if err := s.executeCommand(ctx, command, args, executionCount); err != nil {
				s.logger.Printf("Execution %d failed: %v", executionCount, err)
				if !s.config.ContinueOnError {
					s.mu.Lock()
					s.isRunning = false
					s.mu.Unlock()
					return err
				}
			}
			executionCount++
		}
	}
}

// executeCommand executes a single command with proper error handling and backoff
func (s *IntervalScheduler) executeCommand(ctx context.Context, command string, args []string, attempt int) error {
	s.logger.Printf("Executing command (attempt %d): %s", attempt, command)

	startTime := time.Now()
	err := s.executor.ExecuteWithRetry(ctx, command, args)
	executionTime := time.Since(startTime)

	// Update metrics
	s.metrics.RecordExecution(executionTime, err)

	if err != nil {
		s.logger.Printf("Command execution failed after %v: %v", executionTime, err)
		return err
	}

	s.logger.Printf("Command executed successfully in %v", executionTime)
	return nil
}

// Stop gracefully stops the scheduler
func (s *IntervalScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return nil
	}

	s.logger.Printf("Stopping interval scheduler")
	close(s.stopChan)
	s.isRunning = false

	// Wait for the scheduler to finish
	select {
	case <-s.doneChan:
		s.logger.Printf("Scheduler stopped successfully")
	case <-time.After(30 * time.Second):
		s.logger.Printf("Scheduler stop timeout")
	}

	return nil
}

// IsRunning returns whether the scheduler is currently running
func (s *IntervalScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// StartDaemon starts the scheduler in daemon mode
func (s *IntervalScheduler) StartDaemon(ctx context.Context, command string, args []string) error {
	if err := s.daemon.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Start the scheduler in a goroutine
	go func() {
		defer func() {
			s.daemon.Stop()
			close(s.doneChan)
		}()

		if err := s.Start(ctx, command, args); err != nil {
			s.logger.Printf("Scheduler error: %v", err)
		}
	}()

	return nil
}

// GetMetrics returns the current metrics
func (s *IntervalScheduler) GetMetrics() *Metrics {
	return s.metrics
}

// GetExecutor returns the command executor
func (s *IntervalScheduler) GetExecutor() *CommandExecutor {
	return s.executor
}

// SetExecutor sets the command executor
func (s *IntervalScheduler) SetExecutor(executor *CommandExecutor) {
	s.executor = executor
}

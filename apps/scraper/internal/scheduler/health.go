package scheduler

import (
	"context"
	"log"
	"runtime"
	"time"
)

// HealthMonitor monitors system health during interval execution
type HealthMonitor struct {
	checkInterval time.Duration
	logger        *log.Logger
	metrics       *Metrics
	stopChan      chan struct{}
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(checkInterval time.Duration, logger *log.Logger, metrics *Metrics) *HealthMonitor {
	return &HealthMonitor{
		checkInterval: checkInterval,
		logger:        logger,
		metrics:       metrics,
		stopChan:      make(chan struct{}),
	}
}

// Start begins health monitoring
func (h *HealthMonitor) Start(ctx context.Context) {
	h.logger.Printf("Starting health monitor with interval: %v", h.checkInterval)

	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Printf("Health monitor context cancelled")
			return

		case <-h.stopChan:
			h.logger.Printf("Health monitor stopped")
			return

		case <-ticker.C:
			if err := h.CheckHealth(); err != nil {
				h.logger.Printf("Health check failed: %v", err)
			}
		}
	}
}

// Stop stops the health monitor
func (h *HealthMonitor) Stop() {
	close(h.stopChan)
}

// CheckHealth performs a health check and returns any issues
func (h *HealthMonitor) CheckHealth() error {
	var issues []string

	// Check memory usage
	if memIssue := h.checkMemoryUsage(); memIssue != "" {
		issues = append(issues, memIssue)
	}

	// Check goroutine count
	if goroutineIssue := h.checkGoroutineCount(); goroutineIssue != "" {
		issues = append(issues, goroutineIssue)
	}

	// Check metrics
	if metricsIssue := h.checkMetrics(); metricsIssue != "" {
		issues = append(issues, metricsIssue)
	}

	// Log health status
	if len(issues) == 0 {
		h.logger.Printf("Health check passed")
		return nil
	}

	// Log issues
	for _, issue := range issues {
		h.logger.Printf("Health issue: %s", issue)
	}

	return nil
}

// checkMemoryUsage checks memory usage and returns an issue description if problematic
func (h *HealthMonitor) checkMemoryUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert to MB for readability
	allocMB := m.Alloc / 1024 / 1024
	sysMB := m.Sys / 1024 / 1024

	// Log memory stats
	h.logger.Printf("Memory usage - Alloc: %d MB, Sys: %d MB, NumGC: %d",
		allocMB, sysMB, m.NumGC)

	// Check for potential memory issues
	if allocMB > 1000 { // 1GB threshold
		return "High memory allocation detected"
	}

	if sysMB > 2000 { // 2GB threshold
		return "High system memory usage detected"
	}

	return ""
}

// checkGoroutineCount checks goroutine count and returns an issue description if problematic
func (h *HealthMonitor) checkGoroutineCount() string {
	count := runtime.NumGoroutine()

	h.logger.Printf("Goroutine count: %d", count)

	// Check for potential goroutine leak
	if count > 1000 {
		return "High goroutine count detected - potential leak"
	}

	return ""
}

// checkMetrics checks metrics for potential issues
func (h *HealthMonitor) checkMetrics() string {
	if h.metrics == nil {
		return ""
	}

	executions := h.metrics.GetExecutions()
	failures := h.metrics.GetFailures()
	successRate := h.metrics.GetSuccessRate()

	h.logger.Printf("Metrics - Executions: %d, Failures: %d, Success Rate: %.2f%%",
		executions, failures, successRate)

	// Check for high failure rate
	if executions > 10 && successRate < 80.0 {
		return "Low success rate detected"
	}

	// Check for no recent executions
	lastExecution := h.metrics.GetLastExecution()
	if !lastExecution.IsZero() && time.Since(lastExecution) > 30*time.Minute {
		return "No recent executions detected"
	}

	return ""
}

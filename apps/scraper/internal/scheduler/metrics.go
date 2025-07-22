package scheduler

import (
	"sync"
	"time"
)

// Metrics tracks execution statistics and performance
type Metrics struct {
	executions    int64
	failures      int64
	lastExecution time.Time
	totalRuntime  time.Duration
	mu            sync.RWMutex
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordExecution records an execution with its duration and success status
func (m *Metrics) RecordExecution(duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.executions++
	m.lastExecution = time.Now()
	m.totalRuntime += duration

	if err != nil {
		m.failures++
	}
}

// GetExecutions returns the total number of executions
func (m *Metrics) GetExecutions() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.executions
}

// GetFailures returns the total number of failures
func (m *Metrics) GetFailures() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.failures
}

// GetSuccessRate returns the success rate as a percentage
func (m *Metrics) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.executions == 0 {
		return 0.0
	}

	successes := m.executions - m.failures
	return float64(successes) / float64(m.executions) * 100.0
}

// GetLastExecution returns the time of the last execution
func (m *Metrics) GetLastExecution() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastExecution
}

// GetTotalRuntime returns the total runtime
func (m *Metrics) GetTotalRuntime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalRuntime
}

// GetAverageRuntime returns the average runtime per execution
func (m *Metrics) GetAverageRuntime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.executions == 0 {
		return 0
	}

	return m.totalRuntime / time.Duration(m.executions)
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.executions = 0
	m.failures = 0
	m.lastExecution = time.Time{}
	m.totalRuntime = 0
}

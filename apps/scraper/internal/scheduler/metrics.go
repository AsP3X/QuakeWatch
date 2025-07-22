package scheduler

import (
	"sync"
	"sync/atomic"
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

	atomic.AddInt64(&m.executions, 1)
	m.lastExecution = time.Now()
	m.totalRuntime += duration

	if err != nil {
		atomic.AddInt64(&m.failures, 1)
	}
}

// GetExecutions returns the total number of executions
func (m *Metrics) GetExecutions() int64 {
	return atomic.LoadInt64(&m.executions)
}

// GetFailures returns the total number of failures
func (m *Metrics) GetFailures() int64 {
	return atomic.LoadInt64(&m.failures)
}

// GetSuccessRate returns the success rate as a percentage
func (m *Metrics) GetSuccessRate() float64 {
	executions := m.GetExecutions()
	if executions == 0 {
		return 0.0
	}

	failures := m.GetFailures()
	successes := executions - failures
	return float64(successes) / float64(executions) * 100.0
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
	executions := m.GetExecutions()
	if executions == 0 {
		return 0
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.totalRuntime / time.Duration(executions)
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreInt64(&m.executions, 0)
	atomic.StoreInt64(&m.failures, 0)
	m.lastExecution = time.Time{}
	m.totalRuntime = 0
}

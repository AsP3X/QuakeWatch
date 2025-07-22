package models

import (
	"time"
)

// IntervalExecution represents a single interval execution
type IntervalExecution struct {
	ID            string        `json:"id"`
	Command       string        `json:"command"`
	Args          []string      `json:"args"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time,omitempty"`
	Duration      time.Duration `json:"duration,omitempty"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
	Attempt       int           `json:"attempt"`
	DataCollected int           `json:"data_collected,omitempty"`
}

// IntervalStatus represents the current status of an interval scheduler
type IntervalStatus struct {
	IsRunning      bool          `json:"is_running"`
	StartTime      time.Time     `json:"start_time,omitempty"`
	LastExecution  time.Time     `json:"last_execution,omitempty"`
	NextExecution  time.Time     `json:"next_execution,omitempty"`
	Executions     int64         `json:"executions"`
	Failures       int64         `json:"failures"`
	SuccessRate    float64       `json:"success_rate"`
	TotalRuntime   time.Duration `json:"total_runtime"`
	AverageRuntime time.Duration `json:"average_runtime"`
	Command        string        `json:"command"`
	Interval       time.Duration `json:"interval"`
	MaxExecutions  int           `json:"max_executions"`
	MaxRuntime     time.Duration `json:"max_runtime"`
}

// IntervalConfig represents the configuration for interval execution
type IntervalConfig struct {
	Interval            time.Duration `json:"interval"`
	MaxRuntime          time.Duration `json:"max_runtime"`
	MaxExecutions       int           `json:"max_executions"`
	BackoffStrategy     string        `json:"backoff_strategy"`
	MaxBackoff          time.Duration `json:"max_backoff"`
	ContinueOnError     bool          `json:"continue_on_error"`
	SkipEmpty           bool          `json:"skip_empty"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	DaemonMode          bool          `json:"daemon_mode"`
	PIDFile             string        `json:"pid_file"`
	LogFile             string        `json:"log_file"`
}

// CustomIntervalCommand represents a custom command for interval execution
type CustomIntervalCommand struct {
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Args        []string `json:"args"`
	Description string   `json:"description,omitempty"`
	Enabled     bool     `json:"enabled"`
}

// IntervalExecutionResult represents the result of an interval execution
type IntervalExecutionResult struct {
	Execution *IntervalExecution     `json:"execution"`
	Status    *IntervalStatus        `json:"status"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

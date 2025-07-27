package utils

import (
	"fmt"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType int

const (
	MetricTypeCounter MetricType = iota
	MetricTypeGauge
	MetricTypeHistogram
	MetricTypeSummary
)

// Metric represents a single metric
type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
	events  chan MetricEvent
}

// MetricEvent represents a metric event
type MetricEvent struct {
	Type      string
	Source    string
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*Metric),
		events:  make(chan MetricEvent, 1000),
	}
}

// IncrementCounter increments a counter metric
func (mc *MetricsCollector) IncrementCounter(name string, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.createKey(name, labels)
	if metric, exists := mc.metrics[key]; exists {
		metric.Value++
		metric.Timestamp = time.Now()
	} else {
		mc.metrics[key] = &Metric{
			Name:      name,
			Type:      MetricTypeCounter,
			Value:     1,
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}

	// Send event
	select {
	case mc.events <- MetricEvent{
		Type:      "counter",
		Source:    name,
		Value:     1,
		Labels:    labels,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, skip event
	}
}

// SetGauge sets a gauge metric
func (mc *MetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.createKey(name, labels)
	mc.metrics[key] = &Metric{
		Name:      name,
		Type:      MetricTypeGauge,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	// Send event
	select {
	case mc.events <- MetricEvent{
		Type:      "gauge",
		Source:    name,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, skip event
	}
}

// RecordHistogram records a histogram metric
func (mc *MetricsCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	key := mc.createKey(name, labels)
	if metric, exists := mc.metrics[key]; exists {
		// For simplicity, we'll just track the current value
		// In a real implementation, you'd track buckets, sum, count, etc.
		metric.Value = value
		metric.Timestamp = time.Now()
	} else {
		mc.metrics[key] = &Metric{
			Name:      name,
			Type:      MetricTypeHistogram,
			Value:     value,
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}

	// Send event
	select {
	case mc.events <- MetricEvent{
		Type:      "histogram",
		Source:    name,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, skip event
	}
}

// GetMetric retrieves a metric by name and labels
func (mc *MetricsCollector) GetMetric(name string, labels map[string]string) (*Metric, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	key := mc.createKey(name, labels)
	metric, exists := mc.metrics[key]
	return metric, exists
}

// GetAllMetrics returns all collected metrics
func (mc *MetricsCollector) GetAllMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*Metric)
	for key, metric := range mc.metrics {
		result[key] = metric
	}
	return result
}

// GetEvents returns the events channel
func (mc *MetricsCollector) GetEvents() <-chan MetricEvent {
	return mc.events
}

// createKey creates a unique key for a metric
func (mc *MetricsCollector) createKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}

	// Create a deterministic key from labels
	key := name
	for k, v := range labels {
		key += fmt.Sprintf("_%s_%s", k, v)
	}
	return key
}

// Reset clears all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = make(map[string]*Metric)
}

// CollectionMetrics tracks collection-specific metrics
type CollectionMetrics struct {
	collector *MetricsCollector
}

// NewCollectionMetrics creates new collection metrics
func NewCollectionMetrics() *CollectionMetrics {
	return &CollectionMetrics{
		collector: NewMetricsCollector(),
	}
}

// RecordCollectionStart records the start of a collection operation
func (cm *CollectionMetrics) RecordCollectionStart(source string) {
	cm.collector.SetGauge("collection_active", 1, map[string]string{"source": source})
}

// RecordCollectionEnd records the end of a collection operation
func (cm *CollectionMetrics) RecordCollectionEnd(source string, duration time.Duration, records int, errors int) {
	cm.collector.SetGauge("collection_active", 0, map[string]string{"source": source})
	cm.collector.RecordHistogram("collection_duration_seconds", duration.Seconds(), map[string]string{"source": source})
	cm.collector.IncrementCounter("collection_records_total", map[string]string{"source": source})

	if errors > 0 {
		cm.collector.IncrementCounter("collection_errors_total", map[string]string{"source": source})
	}
}

// RecordAPIError records an API error
func (cm *CollectionMetrics) RecordAPIError(source string, errorType string) {
	cm.collector.IncrementCounter("api_errors_total", map[string]string{
		"source": source,
		"type":   errorType,
	})
}

// RecordDataQuality records data quality metrics
func (cm *CollectionMetrics) RecordDataQuality(source string, qualityScore float64) {
	cm.collector.SetGauge("data_quality_score", qualityScore, map[string]string{"source": source})
}

// GetCollector returns the underlying metrics collector
func (cm *CollectionMetrics) GetCollector() *MetricsCollector {
	return cm.collector
}

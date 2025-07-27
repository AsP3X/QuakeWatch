package utils

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StructuredLogger provides structured logging capabilities
type StructuredLogger struct {
	logger *zap.Logger
	fields map[string]interface{}
}

// CollectionEvent represents a collection event for logging
type CollectionEvent struct {
	Source           string
	RecordsCollected int
	Duration         time.Duration
	Status           string
	QualityScore     float64
	Errors           []error
	Metadata         map[string]interface{}
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(level string, format string) (*StructuredLogger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	var config zap.Config
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &StructuredLogger{
		logger: logger,
		fields: make(map[string]interface{}),
	}, nil
}

// WithField adds a field to the logger
func (l *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
	newLogger := &StructuredLogger{
		logger: l.logger,
		fields: make(map[string]interface{}),
	}

	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	newLogger.fields[key] = value

	return newLogger
}

// WithFields adds multiple fields to the logger
func (l *StructuredLogger) WithFields(fields map[string]interface{}) *StructuredLogger {
	newLogger := &StructuredLogger{
		logger: l.logger,
		fields: make(map[string]interface{}),
	}

	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, l.convertFields()...)
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, l.convertFields()...)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, l.convertFields()...)
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, l.convertFields()...)
}

// LogCollection logs a collection event
func (l *StructuredLogger) LogCollection(ctx context.Context, event CollectionEvent) {
	fields := []zap.Field{
		zap.String("source", event.Source),
		zap.Int("records_collected", event.RecordsCollected),
		zap.Duration("duration", event.Duration),
		zap.String("status", event.Status),
		zap.Float64("quality_score", event.QualityScore),
	}

	if len(event.Errors) > 0 {
		fields = append(fields, zap.Errors("errors", event.Errors))
	}

	for key, value := range event.Metadata {
		fields = append(fields, zap.Any(key, value))
	}

	switch event.Status {
	case "success":
		l.Info("data_collection_completed", fields...)
	case "error":
		l.Error("data_collection_failed", fields...)
	case "partial":
		l.Warn("data_collection_partial", fields...)
	default:
		l.Info("data_collection_event", fields...)
	}
}

// LogAPIRequest logs an API request
func (l *StructuredLogger) LogAPIRequest(ctx context.Context, method, url string, duration time.Duration, statusCode int, err error) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("url", url),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("api_request_failed", fields...)
	} else {
		l.Info("api_request_completed", fields...)
	}
}

// LogCircuitBreaker logs circuit breaker state changes
func (l *StructuredLogger) LogCircuitBreaker(ctx context.Context, source string, state string, stats map[string]interface{}) {
	fields := []zap.Field{
		zap.String("source", source),
		zap.String("state", state),
	}

	for key, value := range stats {
		fields = append(fields, zap.Any(key, value))
	}

	l.Info("circuit_breaker_state_change", fields...)
}

// LogValidation logs validation results
func (l *StructuredLogger) LogValidation(ctx context.Context, source string, valid bool, score float64, errors []error) {
	fields := []zap.Field{
		zap.String("source", source),
		zap.Bool("valid", valid),
		zap.Float64("quality_score", score),
	}

	if len(errors) > 0 {
		fields = append(fields, zap.Errors("validation_errors", errors))
	}

	if valid {
		l.Info("data_validation_passed", fields...)
	} else {
		l.Warn("data_validation_failed", fields...)
	}
}

// convertFields converts map fields to zap fields
func (l *StructuredLogger) convertFields() []zap.Field {
	var fields []zap.Field
	for key, value := range l.fields {
		fields = append(fields, zap.Any(key, value))
	}
	return fields
}

// Sync flushes any buffered log entries
func (l *StructuredLogger) Sync() error {
	return l.logger.Sync()
}

// Global logger instance
var globalLogger *StructuredLogger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(level string, format string) error {
	logger, err := NewStructuredLogger(level, format)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetLogger returns the global logger
func GetLogger() *StructuredLogger {
	if globalLogger == nil {
		// Fallback to basic logger
		logger, _ := NewStructuredLogger("info", "console")
		globalLogger = logger
	}
	return globalLogger
}

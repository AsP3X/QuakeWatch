package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger provides structured logging functionality
type Logger struct {
	logger *logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level string, format string) *Logger {
	logger := logrus.New()

	// Set log level
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set output
	logger.SetOutput(os.Stdout)

	return &Logger{
		logger: logger,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Debug(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Info(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Warn(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Error(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.logger.WithFields(fields).Fatal(msg)
}

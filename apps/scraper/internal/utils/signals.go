package utils

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalManager handles platform-specific signal management
type SignalManager struct {
	platform *Platform
}

// NewSignalManager creates a new signal manager
func NewSignalManager() *SignalManager {
	return &SignalManager{
		platform: GetPlatform(),
	}
}

// SetupGracefulShutdown sets up signal handlers for graceful shutdown
func (s *SignalManager) SetupGracefulShutdown(shutdownFunc func()) {
	sigChan := make(chan os.Signal, 1)

	if s.platform.IsWindows {
		// Windows supports SIGINT and SIGTERM
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	} else {
		// Unix systems support additional signals
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	}

	go func() {
		<-sigChan
		// Log the signal received (could be enhanced with logging)
		shutdownFunc()
	}()
}

// SetupShutdownWithCallback sets up signal handlers with a callback that receives the signal
func (s *SignalManager) SetupShutdownWithCallback(callback func(os.Signal)) {
	sigChan := make(chan os.Signal, 1)

	if s.platform.IsWindows {
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	} else {
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	}

	go func() {
		sig := <-sigChan
		callback(sig)
	}()
}

// GetSupportedSignals returns the list of signals supported on the current platform
func (s *SignalManager) GetSupportedSignals() []os.Signal {
	if s.platform.IsWindows {
		return []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	return []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP}
}

// IsShutdownSignal checks if a signal is a shutdown signal for the current platform
func (s *SignalManager) IsShutdownSignal(sig os.Signal) bool {
	supportedSignals := s.GetSupportedSignals()
	for _, supported := range supportedSignals {
		if sig == supported {
			return true
		}
	}
	return false
}

// GetSignalName returns a human-readable name for a signal
func GetSignalName(sig os.Signal) string {
	switch sig {
	case syscall.SIGINT:
		return "SIGINT"
	case syscall.SIGTERM:
		return "SIGTERM"
	case syscall.SIGHUP:
		return "SIGHUP"
	default:
		return sig.String()
	}
}

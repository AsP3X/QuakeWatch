package scheduler

import (
	"log"
)

// DaemonManager handles daemon/service process management
type DaemonManager interface {
	Start() error
	Stop() error
	IsRunning() bool
	WritePID() error
	RemovePID() error
	SetupLogging() error
}

// DaemonConfig holds configuration for daemon management
type DaemonConfig struct {
	PIDFile     string
	LogFile     string
	Logger      *log.Logger
	ServiceName string
	Description string
}

// NewDaemonManager creates a platform-appropriate daemon manager
func NewDaemonManager(config DaemonConfig) DaemonManager {
	return newPlatformDaemonManager(config)
}

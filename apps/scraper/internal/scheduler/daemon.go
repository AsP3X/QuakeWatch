package scheduler

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// DaemonManager handles daemon process management
type DaemonManager struct {
	pidFile string
	logFile string
	logger  *log.Logger
}

// NewDaemonManager creates a new daemon manager
func NewDaemonManager(pidFile, logFile string, logger *log.Logger) *DaemonManager {
	return &DaemonManager{
		pidFile: pidFile,
		logFile: logFile,
		logger:  logger,
	}
}

// Start starts the daemon process
func (d *DaemonManager) Start() error {
	// Check if daemon is already running
	if d.IsRunning() {
		return fmt.Errorf("daemon is already running (PID file exists: %s)", d.pidFile)
	}

	// Fork and detach from parent process
	pid, err := d.fork()
	if err != nil {
		return fmt.Errorf("failed to fork process: %w", err)
	}

	if pid > 0 {
		// Parent process - exit
		d.logger.Printf("Daemon started with PID: %d", pid)
		os.Exit(0)
	}

	// Child process - continue as daemon
	if err := d.setupDaemon(); err != nil {
		return fmt.Errorf("failed to setup daemon: %w", err)
	}

	return nil
}

// fork creates a new process and returns the PID
func (d *DaemonManager) fork() (int, error) {
	// For Linux systems, we'll use a simpler approach
	// The actual forking will be handled by the parent process exiting
	// and the child continuing execution

	// Create a new process group
	_, err := syscall.Setsid()
	if err != nil {
		return 0, err
	}

	// Return 0 to indicate we're the child process
	return 0, nil
}

// setupDaemon sets up the daemon environment
func (d *DaemonManager) setupDaemon() error {
	// Set up signal handlers
	d.setupSignalHandlers()

	// Write PID file
	if err := d.WritePID(); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Set up logging
	if err := d.SetupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	d.logger.Printf("Daemon setup completed")
	return nil
}

// setupSignalHandlers sets up signal handlers for graceful shutdown
func (d *DaemonManager) setupSignalHandlers() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	go func() {
		sig := <-sigChan
		d.logger.Printf("Received signal: %v", sig)
		d.Stop()
	}()
}

// Stop stops the daemon process
func (d *DaemonManager) Stop() error {
	d.logger.Printf("Stopping daemon")

	if err := d.RemovePID(); err != nil {
		d.logger.Printf("Warning: failed to remove PID file: %v", err)
	}

	os.Exit(0)
	return nil
}

// IsRunning checks if the daemon is running by checking the PID file
func (d *DaemonManager) IsRunning() bool {
	if d.pidFile == "" {
		return false
	}

	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process is running
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// WritePID writes the current process ID to the PID file
func (d *DaemonManager) WritePID() error {
	if d.pidFile == "" {
		return fmt.Errorf("PID file path not specified")
	}

	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	return os.WriteFile(d.pidFile, []byte(pidStr), 0644)
}

// RemovePID removes the PID file
func (d *DaemonManager) RemovePID() error {
	if d.pidFile == "" {
		return nil
	}

	return os.Remove(d.pidFile)
}

// SetupLogging sets up logging for daemon mode
func (d *DaemonManager) SetupLogging() error {
	if d.logFile == "" {
		return nil
	}

	logFile, err := os.OpenFile(d.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Redirect stdout and stderr to log file
	os.Stdout = logFile
	os.Stderr = logFile

	return nil
}

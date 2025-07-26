//go:build windows

package scheduler

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows/svc"
)

// WindowsDaemonManager handles daemon/service process management on Windows
type WindowsDaemonManager struct {
	pidFile string
	logFile string
	logger  *log.Logger
	service *quakewatchService
}

type quakewatchService struct {
	daemon *WindowsDaemonManager
}

// newPlatformDaemonManager creates a Windows daemon manager
func newPlatformDaemonManager(config DaemonConfig) DaemonManager {
	return &WindowsDaemonManager{
		pidFile: config.PIDFile,
		logFile: config.LogFile,
		logger:  config.Logger,
	}
}

// Start starts the daemon process
func (d *WindowsDaemonManager) Start() error {
	// Check if running as Windows service
	isService, err := svc.IsWindowsService()
	if err != nil {
		return fmt.Errorf("failed to check service status: %w", err)
	}

	if isService {
		// Run as Windows service
		d.service = &quakewatchService{daemon: d}
		return svc.Run("QuakeWatchScraper", d.service)
	} else {
		// Run as regular process
		return d.startAsProcess()
	}
}

// startAsProcess starts the application as a regular Windows process
func (d *WindowsDaemonManager) startAsProcess() error {
	// Set up signal handlers for Windows
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Write PID file
	if err := d.WritePID(); err != nil {
		return err
	}

	// Set up logging
	if err := d.SetupLogging(); err != nil {
		return err
	}

	d.logger.Printf("Windows process started")

	// Wait for shutdown signal
	<-sigChan
	return d.Stop()
}

// setupDaemon sets up the daemon environment
func (d *WindowsDaemonManager) setupDaemon() error {
	// Write PID file
	if err := d.WritePID(); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Set up logging
	if err := d.SetupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	d.logger.Printf("Windows service setup completed")
	return nil
}

// Windows service implementation
func (s *quakewatchService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	changes <- svc.Status{State: svc.StartPending}

	// Set up the daemon environment
	if err := s.daemon.setupDaemon(); err != nil {
		return true, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	// Service loop
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				s.daemon.logger.Printf("Windows service stopping")
				return
			default:
				s.daemon.logger.Printf("Unexpected service control request: %d", c)
			}
		}
	}
}

// Stop stops the daemon process
func (d *WindowsDaemonManager) Stop() error {
	d.logger.Printf("Stopping Windows daemon")

	if err := d.RemovePID(); err != nil {
		d.logger.Printf("Warning: failed to remove PID file: %v", err)
	}

	os.Exit(0)
	return nil
}

// IsRunning checks if the daemon is running by checking the PID file
func (d *WindowsDaemonManager) IsRunning() bool {
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
func (d *WindowsDaemonManager) WritePID() error {
	if d.pidFile == "" {
		return fmt.Errorf("PID file path not specified")
	}

	// Ensure directory exists
	dir := filepath.Dir(d.pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create PID file directory: %w", err)
	}

	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)

	return os.WriteFile(d.pidFile, []byte(pidStr), 0644)
}

// RemovePID removes the PID file
func (d *WindowsDaemonManager) RemovePID() error {
	if d.pidFile == "" {
		return nil
	}

	return os.Remove(d.pidFile)
}

// SetupLogging sets up logging for daemon mode
func (d *WindowsDaemonManager) SetupLogging() error {
	if d.logFile == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(d.logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log file directory: %w", err)
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

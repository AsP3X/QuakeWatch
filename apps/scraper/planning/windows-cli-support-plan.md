# Windows CLI App Support Implementation Plan

## Executive Summary

This document outlines the comprehensive plan for implementing Windows CLI app support for the QuakeWatch Data Scraper. The current application is designed for Unix-like systems and requires significant modifications to provide full Windows compatibility while maintaining cross-platform functionality.

## Current State Analysis

### Existing Architecture
- **Language**: Go 1.24
- **CLI Framework**: Cobra
- **Current Platforms**: Linux, macOS
- **Key Dependencies**: PostgreSQL, JSON storage, interval scheduling
- **Daemon Support**: Unix-specific process management

### Windows Compatibility Issues Identified

1. **System Calls**: Uses Unix-specific syscalls (`syscall.Setsid`, `syscall.SIGTERM`)
2. **Process Management**: Daemon functionality relies on Unix process management
3. **File Paths**: Unix-style paths and permissions
4. **Build System**: Makefile missing Windows targets
5. **Service Management**: No Windows service support
6. **Signal Handling**: Unix-specific signal handling

## Implementation Strategy

### Phase 1: Platform Detection and Conditional Compilation

#### 1.1 Create Platform Detection Utilities
**File**: `internal/utils/platform.go`

```go
package utils

import (
    "runtime"
    "strings"
)

// Platform represents the current operating system platform
type Platform struct {
    OS      string
    Arch    string
    IsUnix  bool
    IsWindows bool
}

// GetPlatform returns the current platform information
func GetPlatform() *Platform {
    os := runtime.GOOS
    arch := runtime.GOARCH
    
    return &Platform{
        OS:         os,
        Arch:       arch,
        IsUnix:     os == "linux" || os == "darwin" || os == "freebsd" || os == "openbsd",
        IsWindows:  os == "windows",
    }
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
    return runtime.GOOS == "windows"
}

// IsUnix returns true if running on Unix-like system
func IsUnix() bool {
    return runtime.GOOS == "linux" || runtime.GOOS == "darwin" || 
           runtime.GOOS == "freebsd" || runtime.GOOS == "openbsd"
}
```

#### 1.2 Create Platform-Specific Build Tags
**Files**: 
- `internal/scheduler/daemon_unix.go` (Unix daemon implementation)
- `internal/scheduler/daemon_windows.go` (Windows service implementation)
- `internal/scheduler/daemon_common.go` (Common interface)

### Phase 2: Windows Service Implementation

#### 2.1 Create Windows Service Interface
**File**: `internal/scheduler/daemon_common.go`

```go
package scheduler

import (
    "context"
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

// NewDaemonManager creates a platform-appropriate daemon manager
func NewDaemonManager(pidFile, logFile string, logger *log.Logger) DaemonManager {
    return newPlatformDaemonManager(pidFile, logFile, logger)
}
```

#### 2.2 Implement Windows Service
**File**: `internal/scheduler/daemon_windows.go`

```go
//go:build windows

package scheduler

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "path/filepath"
    "strconv"
    "syscall"
    "time"
    
    "golang.org/x/sys/windows/svc"
    "golang.org/x/sys/windows/svc/debug"
    "golang.org/x/sys/windows/svc/eventlog"
)

type windowsDaemonManager struct {
    pidFile string
    logFile string
    logger  *log.Logger
    service *quakewatchService
}

type quakewatchService struct {
    daemon *windowsDaemonManager
}

func newPlatformDaemonManager(pidFile, logFile string, logger *log.Logger) DaemonManager {
    return &windowsDaemonManager{
        pidFile: pidFile,
        logFile: logFile,
        logger:  logger,
    }
}

func (d *windowsDaemonManager) Start() error {
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
```

### Phase 3: File Path and Permission Handling

#### 3.1 Create Platform-Agnostic Path Utilities
**File**: `internal/utils/paths.go`

```go
package utils

import (
    "os"
    "path/filepath"
    "runtime"
)

// PathManager handles platform-specific path operations
type PathManager struct {
    platform *Platform
}

// NewPathManager creates a new path manager
func NewPathManager() *PathManager {
    return &PathManager{
        platform: GetPlatform(),
    }
}

// GetDefaultPIDFile returns the default PID file path for the platform
func (p *PathManager) GetDefaultPIDFile() string {
    if p.platform.IsWindows {
        return filepath.Join(os.Getenv("TEMP"), "quakewatch-scraper.pid")
    }
    return "/var/run/quakewatch-scraper.pid"
}

// GetDefaultLogFile returns the default log file path for the platform
func (p *PathManager) GetDefaultLogFile() string {
    if p.platform.IsWindows {
        return filepath.Join(os.Getenv("TEMP"), "quakewatch-scraper.log")
    }
    return "/var/log/quakewatch-scraper.log"
}

// EnsureDirectoryExists creates directory with appropriate permissions
func (p *PathManager) EnsureDirectoryExists(path string) error {
    if err := os.MkdirAll(path, p.getDefaultDirPerms()); err != nil {
        return err
    }
    return nil
}

// getDefaultDirPerms returns platform-appropriate directory permissions
func (p *PathManager) getDefaultDirPerms() os.FileMode {
    if p.platform.IsWindows {
        return 0755 // Windows doesn't use Unix permissions, but Go handles this
    }
    return 0755
}
```

### Phase 4: Signal Handling and Process Management

#### 4.1 Create Platform-Agnostic Signal Handler
**File**: `internal/utils/signals.go`

```go
package utils

import (
    "os"
    "os/signal"
    "runtime"
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
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    } else {
        signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
    }
    
    go func() {
        sig := <-sigChan
        // Log the signal received
        shutdownFunc()
    }()
}
```

### Phase 5: Build System Updates

#### 5.1 Update Makefile for Windows Support
**File**: `Makefile`

```makefile
# Add Windows build targets
build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/quakewatch-scraper-windows-amd64.exe cmd/scraper/main.go

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build -o bin/quakewatch-scraper-windows-arm64.exe cmd/scraper/main.go

# Update build-all to include Windows
build-all: build-linux build-darwin build-windows

# Windows-specific setup
setup-windows:
	mkdir -p bin data\earthquakes data\faults

# Windows service management
install-service: build-windows
	./bin/quakewatch-scraper-windows-amd64.exe install-service

uninstall-service: build-windows
	./bin/quakewatch-scraper-windows-amd64.exe uninstall-service

start-service: build-windows
	./bin/quakewatch-scraper-windows-amd64.exe start-service

stop-service: build-windows
	./bin/quakewatch-scraper-windows-amd64.exe stop-service
```

#### 5.2 Add Windows Dependencies
**File**: `go.mod`

```go
require (
    // ... existing dependencies ...
    golang.org/x/sys v0.31.0 // for Windows service support
)
```

### Phase 6: Configuration Updates

#### 6.1 Update Configuration for Windows Paths
**File**: `internal/config/config.go`

```go
// Update DefaultConfig to use platform-appropriate paths
func DefaultConfig() *Config {
    pathManager := utils.NewPathManager()
    
    return &Config{
        // ... existing config ...
        Interval: IntervalConfig{
            DefaultInterval:     1 * time.Hour,
            MaxRuntime:          24 * time.Hour,
            MaxExecutions:       1000,
            BackoffStrategy:     "exponential",
            MaxBackoff:          30 * time.Minute,
            ContinueOnError:     true,
            SkipEmpty:           false,
            HealthCheckInterval: 5 * time.Minute,
            DaemonMode:          false,
            PIDFile:             pathManager.GetDefaultPIDFile(),
            LogFile:             pathManager.GetDefaultLogFile(),
        },
    }
}
```

### Phase 7: CLI Command Updates

#### 7.1 Add Windows Service Commands
**File**: `pkg/cli/commands.go`

```go
// Add to setupCommands method
func (a *App) setupCommands() {
    // ... existing commands ...
    
    // Add Windows service commands
    if utils.IsWindows() {
        a.rootCmd.AddCommand(a.newWindowsServiceCmd())
    }
}

// Add Windows service command
func (a *App) newWindowsServiceCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "service",
        Short: "Windows service management",
        Long:  `Manage QuakeWatch Scraper as a Windows service`,
    }
    
    cmd.AddCommand(a.newInstallServiceCmd())
    cmd.AddCommand(a.newUninstallServiceCmd())
    cmd.AddCommand(a.newStartServiceCmd())
    cmd.AddCommand(a.newStopServiceCmd())
    
    return cmd
}

func (a *App) newInstallServiceCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "install",
        Short: "Install as Windows service",
        RunE:  a.runInstallService,
    }
}

func (a *App) runInstallService(cmd *cobra.Command, args []string) error {
    // Implementation for installing Windows service
    return nil
}

// Similar implementations for uninstall, start, stop
```

### Phase 8: Testing and Documentation

#### 8.1 Create Windows-Specific Tests
**File**: `internal/utils/platform_test.go`

```go
package utils

import (
    "testing"
)

func TestPlatformDetection(t *testing.T) {
    platform := GetPlatform()
    
    if platform.OS == "" {
        t.Error("Platform OS should not be empty")
    }
    
    if platform.Arch == "" {
        t.Error("Platform Arch should not be empty")
    }
}

func TestWindowsPaths(t *testing.T) {
    if !IsWindows() {
        t.Skip("Skipping Windows-specific test on non-Windows platform")
    }
    
    pathManager := NewPathManager()
    pidFile := pathManager.GetDefaultPIDFile()
    
    if pidFile == "" {
        t.Error("PID file path should not be empty")
    }
}
```

#### 8.2 Update Documentation
**File**: `README.md`

```markdown
## Windows Support

### Windows Service Installation

```cmd
# Install as Windows service
quakewatch-scraper.exe service install

# Start the service
quakewatch-scraper.exe service start

# Stop the service
quakewatch-scraper.exe service stop

# Uninstall the service
quakewatch-scraper.exe service uninstall
```

### Windows Build

```cmd
# Build for Windows
make build-windows

# Build for Windows ARM64
make build-windows-arm64
```

### Windows Configuration

The application automatically detects Windows and uses appropriate:
- PID file location: `%TEMP%\quakewatch-scraper.pid`
- Log file location: `%TEMP%\quakewatch-scraper.log`
- Service management through Windows Service Control Manager
```

### Phase 9: CI/CD Updates

#### 9.1 Create GitHub Actions Workflow for Windows
**File**: `.github/workflows/build-windows.yml`

```yaml
name: Build Windows

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build-windows:
    runs-on: windows-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    
    - name: Build
      run: |
        go build -o bin/quakewatch-scraper.exe cmd/scraper/main.go
    
    - name: Test
      run: go test ./...
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: quakewatch-scraper-windows
        path: bin/quakewatch-scraper.exe
```

## Implementation Timeline

### Week 1: Foundation
- [ ] Create platform detection utilities
- [ ] Implement platform-agnostic interfaces
- [ ] Create Windows service framework

### Week 2: Core Windows Support
- [ ] Implement Windows service functionality
- [ ] Create platform-specific path handling
- [ ] Update signal handling for Windows

### Week 3: Build and Configuration
- [ ] Update Makefile for Windows builds
- [ ] Add Windows dependencies
- [ ] Update configuration system

### Week 4: CLI and Testing
- [ ] Add Windows service CLI commands
- [ ] Create Windows-specific tests
- [ ] Update documentation

### Week 5: CI/CD and Polish
- [ ] Set up Windows CI/CD pipeline
- [ ] Final testing and bug fixes
- [ ] Performance optimization

## Success Criteria

1. **Cross-Platform Compatibility**: Application runs seamlessly on Windows, Linux, and macOS
2. **Windows Service Support**: Can be installed and managed as a Windows service
3. **Consistent CLI Experience**: Same commands work across all platforms
4. **Proper Error Handling**: Platform-specific errors are handled gracefully
5. **Comprehensive Testing**: All functionality tested on Windows
6. **Documentation**: Complete Windows-specific documentation
7. **CI/CD Integration**: Automated Windows builds and tests

## Risk Mitigation

1. **Windows Service Complexity**: Use well-tested libraries like `golang.org/x/sys`
2. **Path Differences**: Implement comprehensive path abstraction layer
3. **Permission Issues**: Test with different Windows user permission levels
4. **Build Dependencies**: Ensure all dependencies support Windows
5. **Testing Coverage**: Maintain high test coverage for Windows-specific code

## Technical Considerations

### Windows Service Management
- Use Windows Service Control Manager (SCM)
- Implement proper service lifecycle management
- Handle service dependencies and startup order
- Provide service recovery options

### File System Differences
- Handle Windows path separators (`\` vs `/`)
- Manage Windows file permissions and ACLs
- Handle Windows-specific file locking mechanisms
- Consider Windows file system case sensitivity

### Process Management
- Replace Unix daemon concepts with Windows services
- Implement proper Windows process termination
- Handle Windows-specific process signals
- Manage Windows process priority and affinity

### Network and I/O
- Ensure HTTP client compatibility with Windows
- Handle Windows-specific network configurations
- Manage Windows firewall and security policies
- Consider Windows proxy and VPN configurations

## Testing Strategy

### Unit Testing
- Platform detection logic
- Path handling utilities
- Signal management
- Configuration loading

### Integration Testing
- Windows service installation/uninstallation
- Service start/stop operations
- Cross-platform command execution
- Database connectivity on Windows

### End-to-End Testing
- Full application workflow on Windows
- Windows service lifecycle
- Error handling and recovery
- Performance benchmarking

### Manual Testing
- Different Windows versions (Windows 10, Windows 11, Windows Server)
- Various user permission levels
- Different Windows configurations
- Third-party software interactions

## Deployment Considerations

### Windows Installation
- Create Windows installer (MSI/EXE)
- Handle Windows registry modifications
- Manage Windows service dependencies
- Provide uninstallation procedures

### Configuration Management
- Windows-specific configuration files
- Environment variable handling
- Registry-based configuration
- Group Policy integration

### Monitoring and Logging
- Windows Event Log integration
- Windows Performance Counters
- Windows Task Scheduler integration
- Windows Management Instrumentation (WMI)

## Conclusion

This comprehensive plan provides a roadmap for implementing full Windows CLI app support while maintaining cross-platform compatibility. The phased approach ensures systematic implementation with proper testing and validation at each stage.

The implementation will result in a robust, cross-platform application that provides the same functionality and user experience across Windows, Linux, and macOS platforms. 
# QuakeWatch Scraper - Interval Scraping Feature Plan

## Overview

This document outlines the plan for adding interval scraping functionality to the QuakeWatch scraper. The feature will allow users to run the scraper continuously at specified intervals while supporting all existing options like storage location, query parameters, and data filtering.

## Current Architecture Analysis

### Existing Command Structure
The current scraper supports the following earthquake collection commands:
- `recent` - Collect recent earthquakes (last hour)
- `time-range` - Collect earthquakes by time range
- `magnitude` - Collect earthquakes by magnitude range
- `region` - Collect earthquakes by geographic region
- `significant` - Collect significant earthquakes (M4.5+)
- `country` - Collect earthquakes by country

### Global Flags
- `--config` - Configuration file path
- `--output-dir` - Output directory for JSON files
- `--verbose` - Enable verbose logging
- `--quiet` - Suppress output
- `--log-level` - Set log level
- `--dry-run` - Show what would be done without executing
- `--stdout` - Output data to stdout instead of saving to file

## Feature Requirements

### Core Requirements
1. **Interval-based execution** - Run scraping commands at specified intervals
2. **Support all existing commands** - Work with all current earthquake and fault collection commands
3. **Preserve all options** - Maintain support for all existing flags and parameters
4. **Configurable intervals** - Allow users to specify time intervals (minutes, hours, days)
5. **Graceful shutdown** - Handle interruption signals properly
6. **Logging and monitoring** - Provide clear feedback about interval execution
7. **Error handling** - Continue running on individual command failures
8. **Resource management** - Prevent memory leaks and excessive resource usage

### Additional Requirements
1. **Maximum runtime** - Option to limit total execution time
2. **Execution count limit** - Option to limit number of executions
3. **Conditional execution** - Skip execution if no new data is available
4. **Backoff strategy** - Implement exponential backoff for API failures
5. **Health checks** - Monitor system health during long-running intervals
6. **Daemon mode** - Run in background as a daemon process with PID file management

## Technical Design

### New Command Structure

#### Primary Interval Command
```bash
quakewatch-scraper interval [subcommand] [options]
```

#### Subcommands
```bash
# Run earthquake collection at intervals
quakewatch-scraper interval earthquakes [command] [options]

# Run fault collection at intervals  
quakewatch-scraper interval faults [command] [options]

# Run custom command combinations at intervals
quakewatch-scraper interval custom [options]
```

### New Flags for Interval Commands

#### Interval-specific Flags
```bash
--interval, -i string        # Time interval (e.g., "5m", "1h", "24h")
--max-runtime string         # Maximum total runtime (e.g., "24h", "7d")
--max-executions int         # Maximum number of executions
--backoff string             # Backoff strategy ("none", "linear", "exponential")
--max-backoff string         # Maximum backoff duration
--continue-on-error          # Continue running on individual command failures
--skip-empty                 # Skip execution if no new data is found
--health-check-interval string # Health check interval
--daemon, -d                 # Run in daemon mode (background)
--pid-file string            # PID file location (default: /var/run/quakewatch-scraper.pid)
--log-file string            # Log file location for daemon mode
```

#### Examples
```bash
# Collect recent earthquakes every 5 minutes for 24 hours
quakewatch-scraper interval earthquakes recent --interval 5m --max-runtime 24h

# Collect significant earthquakes every hour with exponential backoff
quakewatch-scraper interval earthquakes significant \
  --start "2024-01-01" --end "2024-01-31" \
  --interval 1h --backoff exponential --max-backoff 30m

# Collect country-specific earthquakes every 6 hours
quakewatch-scraper interval earthquakes country \
  --country "Japan" --interval 6h --max-executions 100

# Custom interval with multiple commands
quakewatch-scraper interval custom \
  --interval 1h \
  --commands "earthquakes recent,earthquakes significant --start 2024-01-01 --end 2024-01-31"

# Run in daemon mode (background)
quakewatch-scraper interval earthquakes recent --interval 5m --daemon --log-file /var/log/quakewatch-scraper.log

# Daemon mode with custom PID file
quakewatch-scraper interval earthquakes country --country "Japan" --interval 1h --daemon --pid-file /tmp/quakewatch-japan.pid
```
```

## Implementation Plan

### Phase 1: Core Interval Infrastructure

#### 1.1 New Package Structure
```
internal/
├── scheduler/
│   ├── interval.go          # Core interval scheduling logic
│   ├── executor.go          # Command execution engine
│   ├── backoff.go           # Backoff strategies
│   ├── health.go            # Health monitoring
│   └── daemon.go            # Daemon process management
├── models/
│   └── interval.go          # Interval configuration models
```

#### 1.2 Configuration Extensions
Add to `internal/config/config.go`:
```go
type IntervalConfig struct {
    DefaultInterval    time.Duration `mapstructure:"default_interval"`
    MaxRuntime         time.Duration `mapstructure:"max_runtime"`
    MaxExecutions      int           `mapstructure:"max_executions"`
    BackoffStrategy    string        `mapstructure:"backoff_strategy"`
    MaxBackoff         time.Duration `mapstructure:"max_backoff"`
    ContinueOnError    bool          `mapstructure:"continue_on_error"`
    SkipEmpty          bool          `mapstructure:"skip_empty"`
    HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
    DaemonMode         bool          `mapstructure:"daemon_mode"`
    PIDFile            string        `mapstructure:"pid_file"`
    LogFile            string        `mapstructure:"log_file"`
}
```

#### 1.3 Core Components

**Interval Scheduler (`internal/scheduler/interval.go`)**
```go
type IntervalScheduler struct {
    config     *config.IntervalConfig
    executor   *CommandExecutor
    logger     *log.Logger
    stopChan   chan struct{}
    doneChan   chan struct{}
    daemon     *DaemonManager
}

func (s *IntervalScheduler) Start(ctx context.Context, command string, args []string) error
func (s *IntervalScheduler) Stop() error
func (s *IntervalScheduler) IsRunning() bool
func (s *IntervalScheduler) StartDaemon(ctx context.Context, command string, args []string) error
```

**Command Executor (`internal/scheduler/executor.go`)**
```go
type CommandExecutor struct {
    app        *cli.App
    backoff    BackoffStrategy
    logger     *log.Logger
}

func (e *CommandExecutor) Execute(ctx context.Context, command string, args []string) error
func (e *CommandExecutor) ExecuteWithRetry(ctx context.Context, command string, args []string) error
```

**Daemon Manager (`internal/scheduler/daemon.go`)**
```go
type DaemonManager struct {
    pidFile    string
    logFile    string
    logger     *log.Logger
}

func (d *DaemonManager) Start() error
func (d *DaemonManager) Stop() error
func (d *DaemonManager) IsRunning() bool
func (d *DaemonManager) WritePID() error
func (d *DaemonManager) RemovePID() error
func (d *DaemonManager) SetupLogging() error
```

### Phase 2: CLI Integration

#### 2.1 New CLI Commands
Add to `pkg/cli/commands.go`:

```go
func (a *App) newIntervalCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "interval",
        Short: "Run commands at specified intervals",
        Long:  `Execute scraping commands at regular intervals with configurable options.`,
    }
    
    cmd.AddCommand(a.newIntervalEarthquakesCmd())
    cmd.AddCommand(a.newIntervalFaultsCmd())
    cmd.AddCommand(a.newIntervalCustomCmd())
    
    return cmd
}
```

#### 2.2 Interval Command Handlers
```go
func (a *App) runIntervalEarthquakes(cmd *cobra.Command, args []string) error
func (a *App) runIntervalFaults(cmd *cobra.Command, args []string) error
func (a *App) runIntervalCustom(cmd *cobra.Command, args []string) error
func (a *App) handleDaemonMode(cmd *cobra.Command, args []string) error
```

### Phase 3: Advanced Features

#### 3.1 Backoff Strategies
```go
type BackoffStrategy interface {
    GetDelay(attempt int) time.Duration
    Reset()
}

type NoBackoff struct{}
type LinearBackoff struct{ baseDelay time.Duration }
type ExponentialBackoff struct{ baseDelay, maxDelay time.Duration }
```

#### 3.2 Health Monitoring
```go
type HealthMonitor struct {
    checkInterval time.Duration
    logger        *log.Logger
    metrics       *Metrics
}

func (h *HealthMonitor) Start(ctx context.Context)
func (h *HealthMonitor) CheckHealth() error
```

#### 3.3 Metrics and Monitoring
```go
type Metrics struct {
    Executions    int64
    Failures      int64
    LastExecution time.Time
    TotalRuntime  time.Duration
}
```

## Configuration Updates

### Updated `configs/config.yaml`
```yaml
interval:
    default_interval: 1h
    max_runtime: 24h
    max_executions: 1000
    backoff_strategy: exponential
    max_backoff: 30m
    continue_on_error: true
    skip_empty: false
    health_check_interval: 5m
    daemon_mode: false
    pid_file: /var/run/quakewatch-scraper.pid
    log_file: /var/log/quakewatch-scraper.log
```

## Daemon Mode Implementation

### Daemon Process Management
1. **Process Detachment** - Fork and detach from parent process
2. **PID File Management** - Create and manage PID file for process tracking
3. **Signal Handling** - Handle SIGTERM, SIGINT, SIGHUP for graceful shutdown
4. **Logging Redirection** - Redirect stdout/stderr to log files
5. **Working Directory** - Set appropriate working directory for daemon

### Daemon Lifecycle
```go
// Daemon startup sequence
1. Parse command line arguments
2. Check if daemon is already running (PID file)
3. Fork process and detach from parent
4. Set up signal handlers
5. Create PID file
6. Redirect logging to file
7. Start interval scheduler
8. Wait for shutdown signal
9. Clean up PID file and exit
```

### Daemon Control Commands
```bash
# Start daemon
quakewatch-scraper interval earthquakes recent --interval 5m --daemon

# Check daemon status
ps aux | grep quakewatch-scraper
cat /var/run/quakewatch-scraper.pid

# Stop daemon
kill $(cat /var/run/quakewatch-scraper.pid)

# Restart daemon
kill $(cat /var/run/quakewatch-scraper.pid) && sleep 2 && quakewatch-scraper interval earthquakes recent --interval 5m --daemon
```

## Error Handling Strategy

### Error Categories
1. **Transient Errors** - API timeouts, network issues (retry with backoff)
2. **Permanent Errors** - Invalid parameters, authentication failures (stop execution)
3. **Resource Errors** - Memory/disk space issues (pause and retry)
4. **Daemon Errors** - PID file conflicts, permission issues, signal handling failures

### Error Recovery
- Implement exponential backoff for transient errors
- Log all errors with appropriate severity levels
- Provide option to continue or stop on errors
- Send notifications for critical failures

## Resource Management

### Memory Management
- Clear collected data after each execution
- Implement garbage collection hints
- Monitor memory usage and log warnings

### Disk Space Management
- Check available disk space before each execution
- Implement file rotation for long-running intervals
- Clean up old files based on retention policy

### API Rate Limiting
- Respect API rate limits across intervals
- Implement intelligent delays between requests
- Track and log API usage patterns

### Daemon Resource Management
- Monitor daemon process resources
- Implement log rotation for long-running daemons
- Handle daemon restart scenarios
- Manage multiple daemon instances

## Testing Strategy

### Unit Tests
- Test interval scheduling logic
- Test backoff strategies
- Test command execution engine
- Test error handling scenarios
- Test daemon process management
- Test PID file operations
- Test signal handling

### Integration Tests
- Test full interval execution cycles
- Test with real API endpoints
- Test resource management
- Test graceful shutdown
- Test daemon mode with real intervals
- Test daemon restart scenarios
- Test multiple daemon instances

### Performance Tests
- Test memory usage over long periods
- Test CPU usage patterns
- Test disk I/O patterns
- Test API rate limiting compliance

## Security Considerations

### API Key Management
- Secure storage of API keys
- Rotation of API keys
- Audit logging of API usage

### File System Security
- Secure file permissions
- Validation of file paths
- Protection against path traversal attacks
- Secure PID file permissions
- Log file access controls

### Network Security
- TLS certificate validation
- Secure HTTP headers
- Protection against injection attacks

## Monitoring and Observability

### Logging
- Structured logging with consistent format
- Log levels for different environments
- Log rotation and retention policies

### Metrics
- Execution count and success rate
- API response times
- Resource usage metrics
- Error rates and types

### Health Checks
- Application health endpoint
- Dependency health checks
- Resource availability checks

## Deployment Considerations

### Container Support
- Docker image optimization
- Health check endpoints
- Graceful shutdown handling

### Systemd Service
- Service file configuration
- Restart policies
- Log management
- Daemon integration with systemd
- PID file management

### Kubernetes Support
- Deployment manifests
- ConfigMap integration
- Secret management

## Migration Strategy

### Backward Compatibility
- All existing commands remain unchanged
- New interval commands are additive
- Configuration files remain compatible

### Documentation Updates
- Update README with interval examples
- Add interval-specific documentation
- Update configuration examples

### User Training
- Provide migration guide
- Create example configurations
- Document best practices

## Success Metrics

### Functional Metrics
- All existing commands work with intervals
- Error handling works correctly
- Resource usage remains stable

### Performance Metrics
- Memory usage stays within limits
- CPU usage is reasonable
- Disk I/O is optimized

### User Experience Metrics
- Clear error messages
- Intuitive command syntax
- Comprehensive documentation

## Risk Assessment

### Technical Risks
- Memory leaks in long-running processes
- API rate limiting issues
- File system corruption

### Mitigation Strategies
- Comprehensive testing
- Resource monitoring
- Graceful error handling
- Regular health checks

## Timeline

### Phase 1 (Week 1-2): Core Infrastructure
- Implement interval scheduler
- Add basic CLI commands
- Create configuration structure
- Implement daemon process management
- Add PID file handling

### Phase 2 (Week 3-4): Advanced Features
- Implement backoff strategies
- Add health monitoring
- Create metrics collection
- Implement signal handling for daemon
- Add log rotation and management

### Phase 3 (Week 5-6): Testing and Documentation
- Comprehensive testing
- Documentation updates
- Performance optimization
- Daemon mode testing
- Systemd service integration

### Phase 4 (Week 7): Deployment and Monitoring
- Deployment preparation
- Monitoring setup
- User training materials

## Conclusion

The interval scraping feature will significantly enhance the QuakeWatch scraper's capabilities by enabling continuous data collection while maintaining all existing functionality. The modular design ensures easy maintenance and future enhancements while providing users with flexible configuration options for their specific use cases. 
# QuakeWatch - Go Data Scraper Plan

## Project Overview
The QuakeWatch Data Scraper is a standalone Go application that compiles to a binary, designed to fetch, validate, clean, and save earthquake and fault data from multiple seismological sources to JSON files. This application operates independently and serves as the data ingestion layer for the entire QuakeWatch system.

## Data Sources

### Earthquake Data Sources
- **USGS FDSNWS Earthquake API**: https://earthquake.usgs.gov/fdsnws/event/1/
- **Query Endpoint**: https://earthquake.usgs.gov/fdsnws/event/1/query
- **Format**: GeoJSON, CSV, XML, QuakeML
- **Content**: Real-time and historical earthquake data with magnitude, location, timing, and detailed metadata
- **Update Frequency**: Real-time (every 5-15 minutes)
- **API Standard**: FDSN Web Services (International Federation of Digital Seismograph Networks)

### Fault Data Sources
- **EMSC-CSEM API**: https://www.emsc-csem.org/javascript/gem_active_faults.geojson
- **Format**: GeoJSON
- **Content**: Active fault data with geographical coordinates and fault properties
- **Update Frequency**: Variable (to be determined based on API availability)

## Go Application Architecture

### Project Structure
```
quakewatch-scraper/
├── cmd/
│   └── scraper/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── usgs.go
│   │   └── emsc.go
│   ├── models/
│   │   ├── earthquake.go
│   │   ├── fault.go
│   │   └── interval.go
│   ├── collector/
│   │   ├── earthquake.go
│   │   └── fault.go
│   ├── scheduler/
│   │   ├── interval.go
│   │   ├── executor.go
│   │   ├── backoff.go
│   │   ├── health.go
│   │   └── daemon.go
│   ├── storage/
│   │   └── json.go
│   └── utils/
│       ├── logger.go
│       └── validator.go
├── pkg/
│   └── cli/
│       └── commands.go
├── configs/
│   └── config.yaml
├── data/
│   ├── earthquakes/
│   └── faults/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Command-Line Interface (CLI)

The scraper application provides a simplified CLI focused on JSON file operations:

#### CLI Structure
```bash
quakewatch-scraper [command] [options]
```

#### Available Commands

**1. Collect Earthquake Data**
```bash
# Collect recent earthquakes (last hour)
quakewatch-scraper earthquakes recent [options]

# Collect earthquakes by time range
quakewatch-scraper earthquakes time-range --start "2024-01-01" --end "2024-01-02" [options]

# Collect earthquakes by magnitude range
quakewatch-scraper earthquakes magnitude --min 4.5 --max 10 [options]

# Collect earthquakes by geographic region
quakewatch-scraper earthquakes region --min-lat 32 --max-lat 42 --min-lon -125 --max-lon -114 [options]

# Collect significant earthquakes (M4.5+) for a period
quakewatch-scraper earthquakes significant --start "2024-01-01" --end "2024-01-02" [options]
```

**2. Collect Fault Data**
```bash
# Collect fault data from EMSC
quakewatch-scraper faults collect [options]

# Update fault data
quakewatch-scraper faults update [options]
```

**3. Data Management**
```bash
# Validate data integrity
quakewatch-scraper validate [options]

# Show data statistics
quakewatch-scraper stats [options]

# List available data files
quakewatch-scraper list [options]
```

**4. Interval Scraping**
```bash
# Run earthquake collection at intervals
quakewatch-scraper interval earthquakes [command] [options]

# Run fault collection at intervals  
quakewatch-scraper interval faults [command] [options]

# Run custom command combinations at intervals
quakewatch-scraper interval custom [options]
```

**5. System Operations**
```bash
# Check system health
quakewatch-scraper health [options]

# Show version information
quakewatch-scraper version [options]

# Show help
quakewatch-scraper help [options]
```

#### Global Options
```bash
# All commands support these global options
--config, -c <file>          # Configuration file path (default: ./configs/config.yaml)
--verbose, -v                # Enable verbose logging
--quiet, -q                  # Suppress output
--log-level <level>          # Set log level (error, warn, info, debug)
--output-dir, -o <dir>       # Output directory for JSON files (default: ./data)
--dry-run                    # Show what would be done without executing
--help, -h                   # Show help for command
```

#### Interval-Specific Options
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

#### Command-Specific Options

**Earthquake Commands**
```bash
--limit <number>             # Limit number of records (default: 1000)
--format <format>            # Data format (geojson, csv, xml) (default: geojson)
--validate                   # Validate data before saving
--clean                      # Clean data before saving
--filename <name>            # Custom filename (without extension)
```

**Fault Commands**
```bash
--force                      # Force update even if data is recent
--validate                   # Validate data before saving
```

**Data Management Commands**
```bash
--file <path>                # Specific file to validate/stats
--type <type>                # Data type (earthquakes, faults, all)
```

## Implementation Details

### Main Application Entry Point
```go
// cmd/scraper/main.go
package main

import (
    "log"
    "os"
    
    "quakewatch-scraper/pkg/cli"
)

func main() {
    app := cli.NewApp()
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}
```

### CLI Package Structure
```go
// pkg/cli/commands.go
package cli

import (
    "github.com/spf13/cobra"
)

type App struct {
    rootCmd *cobra.Command
}

func NewApp() *App {
    app := &App{
        rootCmd: &cobra.Command{
            Use:   "quakewatch-scraper",
            Short: "QuakeWatch Data Scraper - Collect earthquake and fault data",
            Long:  `A Go application for collecting earthquake and fault data from various sources and saving to JSON files.`,
        },
    }
    
    app.setupCommands()
    app.setupFlags()
    
    return app
}

func (a *App) setupCommands() {
    // Add earthquake commands
    a.rootCmd.AddCommand(newEarthquakeCmd())
    
    // Add fault commands
    a.rootCmd.AddCommand(newFaultCmd())
    
    // Add interval commands
    a.rootCmd.AddCommand(newIntervalCmd())
    
    // Add utility commands
    a.rootCmd.AddCommand(newValidateCmd())
    a.rootCmd.AddCommand(newStatsCmd())
    a.rootCmd.AddCommand(newListCmd())
    a.rootCmd.AddCommand(newHealthCmd())
    a.rootCmd.AddCommand(newVersionCmd())
}

func (a *App) setupFlags() {
    a.rootCmd.PersistentFlags().StringP("config", "c", "./configs/config.yaml", "Configuration file path")
    a.rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
    a.rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress output")
    a.rootCmd.PersistentFlags().String("log-level", "info", "Set log level (error, warn, info, debug)")
    a.rootCmd.PersistentFlags().StringP("output-dir", "o", "./data", "Output directory for JSON files")
    a.rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without executing")
}

func (a *App) Run(args []string) error {
    a.rootCmd.SetArgs(args)
    return a.rootCmd.Execute()
}

// Interval command implementation
func newIntervalCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "interval",
        Short: "Run commands at specified intervals",
        Long:  `Execute scraping commands at regular intervals with configurable options.`,
    }
    
    cmd.AddCommand(newIntervalEarthquakesCmd())
    cmd.AddCommand(newIntervalFaultsCmd())
    cmd.AddCommand(newIntervalCustomCmd())
    
    return cmd
}

func newIntervalEarthquakesCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "earthquakes [command]",
        Short: "Run earthquake collection at intervals",
        Long:  `Execute earthquake collection commands at regular intervals.`,
        RunE:  runIntervalEarthquakes,
    }
    
    // Add interval-specific flags
    cmd.Flags().StringP("interval", "i", "1h", "Time interval (e.g., '5m', '1h', '24h')")
    cmd.Flags().String("max-runtime", "", "Maximum total runtime (e.g., '24h', '7d')")
    cmd.Flags().Int("max-executions", 0, "Maximum number of executions")
    cmd.Flags().String("backoff", "exponential", "Backoff strategy (none, linear, exponential)")
    cmd.Flags().String("max-backoff", "30m", "Maximum backoff duration")
    cmd.Flags().Bool("continue-on-error", true, "Continue running on individual command failures")
    cmd.Flags().Bool("skip-empty", false, "Skip execution if no new data is found")
    cmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
    cmd.Flags().String("pid-file", "/var/run/quakewatch-scraper.pid", "PID file location")
    cmd.Flags().String("log-file", "", "Log file location for daemon mode")
    
    return cmd
}

func newIntervalFaultsCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "faults [command]",
        Short: "Run fault collection at intervals",
        Long:  `Execute fault collection commands at regular intervals.`,
        RunE:  runIntervalFaults,
    }
    
    // Add same interval-specific flags as earthquakes
    cmd.Flags().StringP("interval", "i", "1h", "Time interval (e.g., '5m', '1h', '24h')")
    cmd.Flags().String("max-runtime", "", "Maximum total runtime (e.g., '24h', '7d')")
    cmd.Flags().Int("max-executions", 0, "Maximum number of executions")
    cmd.Flags().String("backoff", "exponential", "Backoff strategy (none, linear, exponential)")
    cmd.Flags().String("max-backoff", "30m", "Maximum backoff duration")
    cmd.Flags().Bool("continue-on-error", true, "Continue running on individual command failures")
    cmd.Flags().Bool("skip-empty", false, "Skip execution if no new data is found")
    cmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
    cmd.Flags().String("pid-file", "/var/run/quakewatch-scraper.pid", "PID file location")
    cmd.Flags().String("log-file", "", "Log file location for daemon mode")
    
    return cmd
}

func newIntervalCustomCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "custom",
        Short: "Run custom command combinations at intervals",
        Long:  `Execute custom command combinations at regular intervals.`,
        RunE:  runIntervalCustom,
    }
    
    cmd.Flags().StringP("interval", "i", "1h", "Time interval (e.g., '5m', '1h', '24h')")
    cmd.Flags().String("commands", "", "Comma-separated list of commands to execute")
    cmd.Flags().String("max-runtime", "", "Maximum total runtime (e.g., '24h', '7d')")
    cmd.Flags().Int("max-executions", 0, "Maximum number of executions")
    cmd.Flags().String("backoff", "exponential", "Backoff strategy (none, linear, exponential)")
    cmd.Flags().String("max-backoff", "30m", "Maximum backoff duration")
    cmd.Flags().Bool("continue-on-error", true, "Continue running on individual command failures")
    cmd.Flags().Bool("skip-empty", false, "Skip execution if no new data is found")
    cmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
    cmd.Flags().String("pid-file", "/var/run/quakewatch-scraper.pid", "PID file location")
    cmd.Flags().String("log-file", "", "Log file location for daemon mode")
    
    return cmd
}

func runIntervalEarthquakes(cmd *cobra.Command, args []string) error {
    // Implementation for interval earthquake collection
    return nil
}

func runIntervalFaults(cmd *cobra.Command, args []string) error {
    // Implementation for interval fault collection
    return nil
}

func runIntervalCustom(cmd *cobra.Command, args []string) error {
    // Implementation for interval custom commands
    return nil
}

### Data Models
```go
// internal/models/earthquake.go
package models

import (
    "time"
    "encoding/json"
)

type Earthquake struct {
    ID          string    `json:"id"`
    Magnitude   float64   `json:"magnitude"`
    Place       string    `json:"place"`
    Time        time.Time `json:"time"`
    Updated     time.Time `json:"updated"`
    URL         string    `json:"url"`
    Detail      string    `json:"detail"`
    Felt        int       `json:"felt,omitempty"`
    CDI         float64   `json:"cdi,omitempty"`
    MMI         float64   `json:"mmi,omitempty"`
    Alert       string    `json:"alert,omitempty"`
    Status      string    `json:"status"`
    Tsunami     int       `json:"tsunami"`
    Sig         int       `json:"sig"`
    Net         string    `json:"net"`
    Code        string    `json:"code"`
    IDs         string    `json:"ids"`
    Sources     string    `json:"sources"`
    Types       string    `json:"types"`
    Nst         int       `json:"nst,omitempty"`
    Dmin        float64   `json:"dmin,omitempty"`
    RMS         float64   `json:"rms,omitempty"`
    Gap         float64   `json:"gap,omitempty"`
    MagType     string    `json:"magType,omitempty"`
    Type        string    `json:"type"`
    Title       string    `json:"title"`
    Geometry    Geometry  `json:"geometry"`
    Properties  Properties `json:"properties"`
}

type Geometry struct {
    Type        string    `json:"type"`
    Coordinates []float64 `json:"coordinates"`
}

type Properties struct {
    Mag         float64 `json:"mag"`
    Place       string  `json:"place"`
    Time        int64   `json:"time"`
    Updated     int64   `json:"updated"`
    URL         string  `json:"url"`
    Detail      string  `json:"detail"`
    Felt        int     `json:"felt,omitempty"`
    CDI         float64 `json:"cdi,omitempty"`
    MMI         float64 `json:"mmi,omitempty"`
    Alert       string  `json:"alert,omitempty"`
    Status      string  `json:"status"`
    Tsunami     int     `json:"tsunami"`
    Sig         int     `json:"sig"`
    Net         string  `json:"net"`
    Code        string  `json:"code"`
    IDs         string  `json:"ids"`
    Sources     string  `json:"sources"`
    Types       string  `json:"types"`
    Nst         int     `json:"nst,omitempty"`
    Dmin        float64 `json:"dmin,omitempty"`
    RMS         float64 `json:"rms,omitempty"`
    Gap         float64 `json:"gap,omitempty"`
    MagType     string  `json:"magType,omitempty"`
    Type        string  `json:"type"`
    Title       string  `json:"title"`
}

// internal/models/fault.go
package models

type Fault struct {
    Type     string     `json:"type"`
    Features []FaultFeature `json:"features"`
}

type FaultFeature struct {
    Type       string         `json:"type"`
    Properties FaultProperties `json:"properties"`
    Geometry   FaultGeometry   `json:"geometry"`
}

type FaultProperties struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Type        string  `json:"type"`
    SlipRate    float64 `json:"slip_rate,omitempty"`
    SlipType    string  `json:"slip_type,omitempty"`
    Dip         float64 `json:"dip,omitempty"`
    Rake        float64 `json:"rake,omitempty"`
    Length      float64 `json:"length,omitempty"`
    Width       float64 `json:"width,omitempty"`
    MaxMagnitude float64 `json:"max_magnitude,omitempty"`
}

type FaultGeometry struct {
    Type        string      `json:"type"`
    Coordinates [][]float64 `json:"coordinates"`
}

// internal/models/interval.go
package models

import (
    "time"
)

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

type IntervalMetrics struct {
    Executions    int64
    Failures      int64
    LastExecution time.Time
    TotalRuntime  time.Duration
}

### API Clients
```go
// internal/api/usgs.go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"
    
    "quakewatch-scraper/internal/models"
)

type USGSClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewUSGSClient() *USGSClient {
    return &USGSClient{
        baseURL: "https://earthquake.usgs.gov/fdsnws/event/1",
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *USGSClient) GetEarthquakes(params map[string]string) (*models.Earthquake, error) {
    u, err := url.Parse(c.baseURL + "/query")
    if err != nil {
        return nil, err
    }
    
    q := u.Query()
    for key, value := range params {
        q.Set(key, value)
    }
    u.RawQuery = q.Encode()
    
    resp, err := c.httpClient.Get(u.String())
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
    }
    
    var earthquakes models.Earthquake
    if err := json.NewDecoder(resp.Body).Decode(&earthquakes); err != nil {
        return nil, err
    }
    
    return &earthquakes, nil
}

// internal/api/emsc.go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "quakewatch-scraper/internal/models"
)

type EMSCClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewEMSCClient() *EMSCClient {
    return &EMSCClient{
        baseURL: "https://www.emsc-csem.org/javascript",
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *EMSCClient) GetFaults() (*models.Fault, error) {
    resp, err := c.httpClient.Get(c.baseURL + "/gem_active_faults.geojson")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
    }
    
    var faults models.Fault
    if err := json.NewDecoder(resp.Body).Decode(&faults); err != nil {
        return nil, err
    }
    
    return &faults, nil
}
```

### Scheduler Components

```go
// internal/scheduler/interval.go
package scheduler

import (
    "context"
    "time"
    "log"
    
    "quakewatch-scraper/internal/config"
)

type IntervalScheduler struct {
    config     *config.IntervalConfig
    executor   *CommandExecutor
    logger     *log.Logger
    stopChan   chan struct{}
    doneChan   chan struct{}
    daemon     *DaemonManager
}

func NewIntervalScheduler(config *config.IntervalConfig) *IntervalScheduler {
    return &IntervalScheduler{
        config:   config,
        executor: NewCommandExecutor(),
        logger:   log.New(os.Stdout, "[SCHEDULER] ", log.LstdFlags),
        stopChan: make(chan struct{}),
        doneChan: make(chan struct{}),
        daemon:   NewDaemonManager(config.PIDFile, config.LogFile),
    }
}

func (s *IntervalScheduler) Start(ctx context.Context, command string, args []string) error {
    ticker := time.NewTicker(s.config.DefaultInterval)
    defer ticker.Stop()
    
    s.logger.Printf("Starting interval scheduler with interval: %v", s.config.DefaultInterval)
    
    for {
        select {
        case <-ctx.Done():
            s.logger.Println("Context cancelled, stopping scheduler")
            return ctx.Err()
        case <-s.stopChan:
            s.logger.Println("Stop signal received, stopping scheduler")
            return nil
        case <-ticker.C:
            if err := s.executor.Execute(ctx, command, args); err != nil {
                s.logger.Printf("Execution failed: %v", err)
                if !s.config.ContinueOnError {
                    return err
                }
            }
        }
    }
}

func (s *IntervalScheduler) Stop() error {
    close(s.stopChan)
    <-s.doneChan
    return nil
}

func (s *IntervalScheduler) IsRunning() bool {
    select {
    case <-s.stopChan:
        return false
    default:
        return true
    }
}

// internal/scheduler/executor.go
package scheduler

import (
    "context"
    "time"
    "log"
)

type CommandExecutor struct {
    app        *cli.App
    backoff    BackoffStrategy
    logger     *log.Logger
}

func NewCommandExecutor() *CommandExecutor {
    return &CommandExecutor{
        app:     cli.NewApp(),
        backoff: NewExponentialBackoff(5*time.Second, 30*time.Minute),
        logger:  log.New(os.Stdout, "[EXECUTOR] ", log.LstdFlags),
    }
}

func (e *CommandExecutor) Execute(ctx context.Context, command string, args []string) error {
    e.logger.Printf("Executing command: %s %v", command, args)
    
    // Execute the command with retry logic
    return e.ExecuteWithRetry(ctx, command, args)
}

func (e *CommandExecutor) ExecuteWithRetry(ctx context.Context, command string, args []string) error {
    maxAttempts := 3
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        if err := e.app.Run(append([]string{command}, args...)); err != nil {
            if attempt == maxAttempts {
                return err
            }
            
            delay := e.backoff.GetDelay(attempt)
            e.logger.Printf("Execution failed, retrying in %v (attempt %d/%d)", delay, attempt, maxAttempts)
            
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
                continue
            }
        }
        return nil
    }
    return nil
}

// internal/scheduler/backoff.go
package scheduler

import (
    "time"
    "math"
)

type BackoffStrategy interface {
    GetDelay(attempt int) time.Duration
    Reset()
}

type NoBackoff struct{}

func (n *NoBackoff) GetDelay(attempt int) time.Duration { return 0 }
func (n *NoBackoff) Reset() {}

type LinearBackoff struct{ baseDelay time.Duration }

func (l *LinearBackoff) GetDelay(attempt int) time.Duration {
    return time.Duration(attempt) * l.baseDelay
}

func (l *LinearBackoff) Reset() {}

type ExponentialBackoff struct{ baseDelay, maxDelay time.Duration }

func (e *ExponentialBackoff) GetDelay(attempt int) time.Duration {
    delay := time.Duration(float64(e.baseDelay) * math.Pow(2, float64(attempt-1)))
    if delay > e.maxDelay {
        delay = e.maxDelay
    }
    return delay
}

func (e *ExponentialBackoff) Reset() {}

func NewExponentialBackoff(baseDelay, maxDelay time.Duration) BackoffStrategy {
    return &ExponentialBackoff{baseDelay: baseDelay, maxDelay: maxDelay}
}

// internal/scheduler/daemon.go
package scheduler

import (
    "fmt"
    "os"
    "syscall"
    "log"
)

type DaemonManager struct {
    pidFile    string
    logFile    string
    logger     *log.Logger
}

func NewDaemonManager(pidFile, logFile string) *DaemonManager {
    return &DaemonManager{
        pidFile: pidFile,
        logFile: logFile,
        logger:  log.New(os.Stdout, "[DAEMON] ", log.LstdFlags),
    }
}

func (d *DaemonManager) Start() error {
    // Fork and detach from parent process
    if err := d.fork(); err != nil {
        return err
    }
    
    // Set up signal handlers
    d.setupSignalHandlers()
    
    // Write PID file
    if err := d.WritePID(); err != nil {
        return err
    }
    
    // Set up logging
    if err := d.SetupLogging(); err != nil {
        return err
    }
    
    d.logger.Println("Daemon started successfully")
    return nil
}

func (d *DaemonManager) Stop() error {
    if err := d.RemovePID(); err != nil {
        return err
    }
    d.logger.Println("Daemon stopped")
    return nil
}

func (d *DaemonManager) IsRunning() bool {
    if _, err := os.Stat(d.pidFile); os.IsNotExist(err) {
        return false
    }
    
    pid, err := d.readPID()
    if err != nil {
        return false
    }
    
    process, err := os.FindProcess(pid)
    if err != nil {
        return false
    }
    
    return process.Signal(syscall.Signal(0)) == nil
}

func (d *DaemonManager) WritePID() error {
    pid := os.Getpid()
    return os.WriteFile(d.pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func (d *DaemonManager) RemovePID() error {
    return os.Remove(d.pidFile)
}

func (d *DaemonManager) SetupLogging() error {
    if d.logFile == "" {
        return nil
    }
    
    file, err := os.OpenFile(d.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return err
    }
    
    // Redirect stdout and stderr to log file
    os.Stdout = file
    os.Stderr = file
    
    return nil
}

func (d *DaemonManager) fork() error {
    // Fork the process
    pid, err := syscall.Fork()
    if err != nil {
        return err
    }
    
    if pid > 0 {
        // Parent process - exit
        os.Exit(0)
    }
    
    // Child process - create new session
    if _, err := syscall.Setsid(); err != nil {
        return err
    }
    
    return nil
}

func (d *DaemonManager) setupSignalHandlers() {
    // Set up signal handling for graceful shutdown
    // Implementation depends on the signal handling library used
}

func (d *DaemonManager) readPID() (int, error) {
    data, err := os.ReadFile(d.pidFile)
    if err != nil {
        return 0, err
    }
    
    var pid int
    _, err = fmt.Sscanf(string(data), "%d", &pid)
    return pid, err
}

### Storage Layer
```go
// internal/storage/json.go
package storage

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
    
    "quakewatch-scraper/internal/models"
)

type JSONStorage struct {
    outputDir string
}

func NewJSONStorage(outputDir string) *JSONStorage {
    return &JSONStorage{
        outputDir: outputDir,
    }
}

func (s *JSONStorage) SaveEarthquakes(earthquakes *models.Earthquake, filename string) error {
    if filename == "" {
        timestamp := time.Now().Format("2006-01-02_15-04-05")
        filename = fmt.Sprintf("earthquakes_%s.json", timestamp)
    } else if !filepath.HasSuffix(filename, ".json") {
        filename += ".json"
    }
    
    filepath := filepath.Join(s.outputDir, "earthquakes", filename)
    
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(filepath), 0755); err != nil {
        return err
    }
    
    file, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    
    return encoder.Encode(earthquakes)
}

func (s *JSONStorage) SaveFaults(faults *models.Fault, filename string) error {
    if filename == "" {
        timestamp := time.Now().Format("2006-01-02_15-04-05")
        filename = fmt.Sprintf("faults_%s.json", timestamp)
    } else if !filepath.HasSuffix(filename, ".json") {
        filename += ".json"
    }
    
    filepath := filepath.Join(s.outputDir, "faults", filename)
    
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(filepath), 0755); err != nil {
        return err
    }
    
    file, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    
    return encoder.Encode(faults)
}

func (s *JSONStorage) ListFiles(dataType string) ([]string, error) {
    var dir string
    switch dataType {
    case "earthquakes":
        dir = filepath.Join(s.outputDir, "earthquakes")
    case "faults":
        dir = filepath.Join(s.outputDir, "faults")
    default:
        return nil, fmt.Errorf("unknown data type: %s", dataType)
    }
    
    files, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return []string{}, nil
        }
        return nil, err
    }
    
    var filenames []string
    for _, file := range files {
        if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
            filenames = append(filenames, file.Name())
        }
    }
    
    return filenames, nil
}
```

### Configuration
```yaml
# configs/config.yaml
api:
  usgs:
    base_url: "https://earthquake.usgs.gov/fdsnws/event/1"
    timeout: 30s
    rate_limit: 60
  emsc:
    base_url: "https://www.emsc-csem.org/javascript"
    timeout: 30s

storage:
  output_dir: "./data"
  earthquakes_dir: "earthquakes"
  faults_dir: "faults"

logging:
  level: "info"
  format: "json"
  output: "stdout"

collection:
  default_limit: 1000
  max_limit: 10000
  retry_attempts: 3
  retry_delay: 5s

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

## Build and Deployment

### Makefile
```makefile
# Makefile
.PHONY: build clean test run install

# Build the application
build:
	go build -o bin/quakewatch-scraper cmd/scraper/main.go

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/quakewatch-scraper-linux-amd64 cmd/scraper/main.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/quakewatch-scraper-darwin-amd64 cmd/scraper/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/quakewatch-scraper-windows-amd64.exe cmd/scraper/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run tests
test:
	go test ./...

# Run the application
run: build
	./bin/quakewatch-scraper

# Install dependencies
install:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Generate documentation
docs:
	godoc -http=:6060
```

### Go Module
```go
// go.mod
module quakewatch-scraper

go 1.21

require (
    github.com/spf13/cobra v1.7.0
    github.com/spf13/viper v1.16.0
    github.com/sirupsen/logrus v1.9.3
)

require (
    github.com/fsnotify/fsnotify v1.6.0 // indirect
    github.com/hashicorp/hcl v1.0.0 // indirect
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/magiconair/properties v1.8.7 // indirect
    github.com/mitchellh/mapstructure v1.5.0 // indirect
    github.com/pelletier/go-toml/v2 v2.0.8 // indirect
    github.com/spf13/afero v1.9.5 // indirect
    github.com/spf13/cast v1.5.1 // indirect
    github.com/spf13/jwalterweatherman v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    github.com/subosito/gotenv v1.4.2 // indirect
    golang.org/x/sys v0.8.0 // indirect
    golang.org/x/text v0.9.0 // indirect
    gopkg.in/ini.v1 v1.67.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

## Usage Examples

### Basic Usage
```bash
# Build the application
make build

# Collect recent earthquakes
./bin/quakewatch-scraper earthquakes recent

# Collect earthquakes for a specific time range
./bin/quakewatch-scraper earthquakes time-range --start "2024-01-01" --end "2024-01-02"

# Collect significant earthquakes
./bin/quakewatch-scraper earthquakes significant --start "2024-01-01" --end "2024-01-02"

# Collect fault data
./bin/quakewatch-scraper faults collect

# Validate collected data
./bin/quakewatch-scraper validate --type earthquakes

# Show statistics
./bin/quakewatch-scraper stats --type all

# List available files
./bin/quakewatch-scraper list --type earthquakes
```

### Advanced Usage
```bash
# Use custom output directory
./bin/quakewatch-scraper earthquakes recent --output-dir /path/to/data

# Enable verbose logging
./bin/quakewatch-scraper earthquakes recent --verbose --log-level debug

# Dry run to see what would be collected
./bin/quakewatch-scraper earthquakes recent --dry-run

# Custom filename
./bin/quakewatch-scraper earthquakes recent --filename my_earthquakes

# Limit number of records
./bin/quakewatch-scraper earthquakes recent --limit 100
```

### Interval Usage Examples
```bash
# Collect recent earthquakes every 5 minutes for 24 hours
./bin/quakewatch-scraper interval earthquakes recent --interval 5m --max-runtime 24h

# Collect significant earthquakes every hour with exponential backoff
./bin/quakewatch-scraper interval earthquakes significant \
  --start "2024-01-01" --end "2024-01-31" \
  --interval 1h --backoff exponential --max-backoff 30m

# Collect country-specific earthquakes every 6 hours
./bin/quakewatch-scraper interval earthquakes country \
  --country "Japan" --interval 6h --max-executions 100

# Custom interval with multiple commands
./bin/quakewatch-scraper interval custom \
  --interval 1h \
  --commands "earthquakes recent,earthquakes significant --start 2024-01-01 --end 2024-01-31"

# Run in daemon mode (background)
./bin/quakewatch-scraper interval earthquakes recent --interval 5m --daemon --log-file /var/log/quakewatch-scraper.log

# Daemon mode with custom PID file
./bin/quakewatch-scraper interval earthquakes country --country "Japan" --interval 1h --daemon --pid-file /tmp/quakewatch-japan.pid

# Collect fault data every 12 hours
./bin/quakewatch-scraper interval faults collect --interval 12h --continue-on-error

# Health check and monitoring
./bin/quakewatch-scraper interval earthquakes recent --interval 5m --health-check-interval 1m
```

## Future Enhancements

This comprehensive Go scraper plan provides a solid foundation that can be extended with:

1. **Database Integration**: Add PostgreSQL/MySQL support for persistent storage
2. **Real-time Monitoring**: Add WebSocket support for real-time data streaming
3. **Advanced Filtering**: Add more sophisticated query capabilities
4. **Data Processing**: Add data cleaning, validation, and transformation features
5. **API Server**: Add HTTP API endpoints for data access
6. **Metrics and Monitoring**: Add Prometheus metrics and health checks
7. **Configuration Management**: Add environment-based configuration
8. **Testing**: Add comprehensive unit and integration tests
9. **Documentation**: Add API documentation and user guides
10. **Advanced Scheduling**: Add cron-based scheduling as an alternative to interval-based execution
11. **Distributed Execution**: Add support for running multiple scraper instances
12. **Data Analytics**: Add built-in data analysis and reporting capabilities

The current plan includes interval-based execution, daemon mode, and comprehensive error handling while maintaining simplicity and reliability.

## Interval Scraping Features

### Core Interval Functionality
The interval scraping feature enables continuous data collection with the following capabilities:

1. **Flexible Intervals**: Support for various time intervals (minutes, hours, days)
2. **Runtime Limits**: Configurable maximum runtime and execution count limits
3. **Error Handling**: Robust error handling with configurable retry strategies
4. **Backoff Strategies**: Exponential, linear, or no backoff for failed executions
5. **Daemon Mode**: Background execution with PID file management
6. **Health Monitoring**: Built-in health checks and monitoring capabilities
7. **Resource Management**: Memory and disk space monitoring
8. **Graceful Shutdown**: Proper signal handling for clean termination

### Daemon Mode Features
- **Process Management**: Fork and detach from parent process
- **PID File Management**: Create and manage PID files for process tracking
- **Signal Handling**: Handle SIGTERM, SIGINT, SIGHUP for graceful shutdown
- **Logging Redirection**: Redirect stdout/stderr to log files
- **Working Directory**: Set appropriate working directory for daemon

### Error Recovery Strategies
- **Transient Errors**: API timeouts, network issues (retry with backoff)
- **Permanent Errors**: Invalid parameters, authentication failures (stop execution)
- **Resource Errors**: Memory/disk space issues (pause and retry)
- **Daemon Errors**: PID file conflicts, permission issues, signal handling failures

### Resource Management
- **Memory Management**: Clear collected data after each execution
- **Disk Space Management**: Check available space before each execution
- **API Rate Limiting**: Respect API rate limits across intervals
- **Log Rotation**: Implement log rotation for long-running daemons

### Security Considerations
- **API Key Management**: Secure storage and rotation of API keys
- **File System Security**: Secure file permissions and path validation
- **Network Security**: TLS certificate validation and secure HTTP headers
- **Process Security**: Secure PID file permissions and access controls

### Monitoring and Observability
- **Structured Logging**: Consistent log format with appropriate levels
- **Metrics Collection**: Execution count, success rate, API response times
- **Health Checks**: Application health endpoints and dependency checks
- **Resource Monitoring**: Memory usage, CPU usage, disk I/O patterns

### Deployment Considerations
- **Container Support**: Docker image optimization and health check endpoints
- **Systemd Integration**: Service file configuration and restart policies
- **Kubernetes Support**: Deployment manifests and ConfigMap integration
- **Multi-Instance Support**: Support for running multiple daemon instances 
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
│   │   └── fault.go
│   ├── collector/
│   │   ├── earthquake.go
│   │   └── fault.go
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

**4. System Operations**
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
```

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
```

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

## Future Enhancements

This simplified Go scraper plan provides a solid foundation that can be extended with:

1. **Database Integration**: Add PostgreSQL/MySQL support for persistent storage
2. **Real-time Monitoring**: Add WebSocket support for real-time data streaming
3. **Advanced Filtering**: Add more sophisticated query capabilities
4. **Data Processing**: Add data cleaning, validation, and transformation features
5. **Scheduling**: Add cron-based scheduling for automated collection
6. **API Server**: Add HTTP API endpoints for data access
7. **Metrics and Monitoring**: Add Prometheus metrics and health checks
8. **Configuration Management**: Add environment-based configuration
9. **Testing**: Add comprehensive unit and integration tests
10. **Documentation**: Add API documentation and user guides

The current plan focuses on simplicity and reliability while providing a clear path for future enhancements. 
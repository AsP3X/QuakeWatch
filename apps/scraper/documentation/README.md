# QuakeWatch Scraper

A comprehensive Go-based data collection tool for earthquake and fault data from multiple seismological sources. The QuakeWatch Scraper is designed to fetch, validate, clean, and store earthquake and fault data in both JSON files and PostgreSQL databases.

## 🚀 Features

- **Multi-Source Data Collection**: Collects earthquake data from USGS FDSNWS API and fault data from EMSC-CSEM
- **Flexible Data Storage**: Supports both JSON file storage and PostgreSQL database storage
- **Scheduled Collection**: Built-in scheduler for automated data collection at configurable intervals
- **Comprehensive CLI**: Rich command-line interface with multiple collection strategies
- **Data Validation**: Built-in validation and cleaning of collected data
- **Cross-Platform**: Supports Linux, macOS, and Windows
- **Database Management**: Full database migration and management capabilities
- **Health Monitoring**: Built-in health checks and monitoring capabilities

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Architecture](#architecture)
- [API Reference](#api-reference)
- [Database Schema](#database-schema)
- [Development](#development)
- [Contributing](#contributing)

## 🛠️ Installation

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 12+ (optional, for database storage)
- Git

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd quakewatch-scraper

# Install dependencies
make install

# Build the application
make build

# Or build for all platforms
make build-all
```

### Docker Setup (Optional)

```bash
# Set up PostgreSQL with Docker
make db-setup-docker

# Set up environment
make setup-env
```

## 🚀 Quick Start

### Basic Usage

```bash
# Check version
./bin/quakewatch-scraper version

# Check health
./bin/quakewatch-scraper health

# Collect recent earthquakes (last hour)
./bin/quakewatch-scraper earthquakes recent --limit 100

# Collect fault data
./bin/quakewatch-scraper faults collect

# View statistics
./bin/quakewatch-scraper stats
```

### Database Setup (Optional)

```bash
# Initialize database
./bin/quakewatch-scraper db init

# Run migrations
./bin/quakewatch-scraper db migrate up

# Check database status
./bin/quakewatch-scraper db status
```

## ⚙️ Configuration

The application uses a YAML configuration file. Create `config.yaml` in your working directory or specify with `--config` flag.

### Default Configuration

```yaml
api:
    emsc:
        base_url: https://www.emsc-csem.org/javascript
        timeout: 30s
    usgs:
        base_url: https://earthquake.usgs.gov/fdsnws/event/1
        rate_limit: 60
        timeout: 30s
collection:
    default_limit: 1000
    max_limit: 10000
    retry_attempts: 3
    retry_delay: 5s
database:
    connection_timeout: 30s
    database: quakewatch
    enabled: true
    host: localhost
    max_connections: 10
    password: postgres
    port: 5432
    ssl_mode: disable
    type: postgres
    username: postgres
logging:
    format: json
    level: info
    output: stdout
storage:
    earthquakes_dir: earthquakes
    faults_dir: faults
    output_dir: ./data
```

### Environment Variables

You can also use environment variables for configuration:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=quakewatch
export DB_SSL_MODE=disable
```

## 📖 Usage

### Earthquake Data Collection

#### Recent Earthquakes
```bash
# Collect earthquakes from the last hour
./bin/quakewatch-scraper earthquakes recent

# Limit the number of results
./bin/quakewatch-scraper earthquakes recent --limit 50

# Output to specific file
./bin/quakewatch-scraper earthquakes recent --output recent_quakes.json
```

#### Time Range Collection
```bash
# Collect earthquakes for a specific time range
./bin/quakewatch-scraper earthquakes time-range \
    --start "2024-01-01T00:00:00Z" \
    --end "2024-01-02T00:00:00Z"

# Use relative time
./bin/quakewatch-scraper earthquakes time-range \
    --start "2024-01-01" \
    --end "2024-01-02"
```

#### Magnitude-Based Collection
```bash
# Collect earthquakes by magnitude range
./bin/quakewatch-scraper earthquakes magnitude \
    --min 4.5 \
    --max 10.0 \
    --start "2024-01-01" \
    --end "2024-01-31"
```

#### Significant Earthquakes
```bash
# Collect significant earthquakes (M4.5+)
./bin/quakewatch-scraper earthquakes significant \
    --start "2024-01-01" \
    --end "2024-01-31"
```

#### Geographic Region Collection
```bash
# Collect earthquakes in a specific region
./bin/quakewatch-scraper earthquakes region \
    --min-lat 32.0 \
    --max-lat 42.0 \
    --min-lon -125.0 \
    --max-lon -114.0 \
    --start "2024-01-01" \
    --end "2024-01-31"
```

#### Country-Specific Collection
```bash
# Collect earthquakes for a specific country
./bin/quakewatch-scraper earthquakes country \
    --country "United States" \
    --start "2024-01-01" \
    --end "2024-01-31"
```

### Fault Data Collection

```bash
# Collect fault data from EMSC
./bin/quakewatch-scraper faults collect

# Update existing fault data
./bin/quakewatch-scraper faults update
```

### Scheduled Collection

#### Interval-Based Collection
```bash
# Collect recent earthquakes every 5 minutes
./bin/quakewatch-scraper interval earthquakes recent \
    --interval 5m \
    --duration 1h

# Collect significant earthquakes every hour
./bin/quakewatch-scraper interval earthquakes significant \
    --interval 1h \
    --duration 24h \
    --start "2024-01-01" \
    --end "2024-01-31"
```

#### Fault Data Scheduling
```bash
# Collect fault data daily
./bin/quakewatch-scraper interval faults collect \
    --interval 24h \
    --duration 7d
```

### Database Operations

```bash
# Initialize database
./bin/quakewatch-scraper db init

# Run migrations
./bin/quakewatch-scraper db migrate up

# Check migration status
./bin/quakewatch-scraper db status

# Rollback migrations
./bin/quakewatch-scraper db migrate down

# Force migration version
./bin/quakewatch-scraper db force-version 1
```

### Utility Commands

#### Statistics
```bash
# View collection statistics
./bin/quakewatch-scraper stats

# View statistics for specific time range
./bin/quakewatch-scraper stats --start "2024-01-01" --end "2024-01-31"
```

#### Data Validation
```bash
# Validate collected data
./bin/quakewatch-scraper validate

# Validate specific file
./bin/quakewatch-scraper validate --file earthquakes.json
```

#### Data Management
```bash
# List collected data
./bin/quakewatch-scraper list

# Purge old data
./bin/quakewatch-scraper purge --older-than 30d

# Dry run purge
./bin/quakewatch-scraper purge --older-than 30d --dry-run
```

#### Health Checks
```bash
# Check application health
./bin/quakewatch-scraper health

# Check database health
./bin/quakewatch-scraper health --check-db
```

## 🏗️ Architecture

### Project Structure

```
quakewatch-scraper/
├── cmd/
│   └── scraper/
│       └── main.go                 # Application entry point
├── internal/
│   ├── api/                        # API clients
│   │   ├── usgs.go                # USGS FDSNWS API client
│   │   └── emsc.go                # EMSC-CSEM API client
│   ├── collector/                  # Data collection logic
│   │   ├── earthquake.go          # Earthquake data collector
│   │   └── fault.go               # Fault data collector
│   ├── config/                     # Configuration management
│   │   ├── config.go              # Main configuration
│   │   └── database.go            # Database configuration
│   ├── models/                     # Data models
│   │   ├── earthquake.go          # Earthquake data structures
│   │   ├── fault.go               # Fault data structures
│   │   └── interval.go            # Interval configuration
│   ├── scheduler/                  # Scheduling and daemon logic
│   │   ├── interval.go            # Interval management
│   │   ├── executor.go            # Command execution
│   │   ├── daemon_common.go       # Common daemon functionality
│   │   ├── daemon_unix.go         # Unix daemon implementation
│   │   ├── daemon_windows.go      # Windows daemon implementation
│   │   ├── health.go              # Health monitoring
│   │   ├── metrics.go             # Metrics collection
│   │   └── backoff.go             # Retry logic
│   ├── storage/                    # Data storage
│   │   ├── interface.go           # Storage interface
│   │   ├── json.go                # JSON file storage
│   │   ├── postgresql.go          # PostgreSQL storage
│   │   ├── migrate.go             # Database migrations
│   │   └── postgresql_test.go     # PostgreSQL tests
│   └── utils/                      # Utility functions
│       ├── logger.go              # Logging utilities
│       ├── paths.go               # Path utilities
│       ├── platform.go            # Platform detection
│       ├── platform_test.go       # Platform tests
│       └── signals.go             # Signal handling
├── pkg/
│   └── cli/                        # Command-line interface
│       └── commands.go            # CLI commands implementation
├── configs/                        # Configuration files
│   └── config.yaml                # Default configuration
├── migrations/                     # Database migrations
│   ├── 000001_combined_schema.up.sql
│   └── 000001_combined_schema.down.sql
├── bin/                           # Compiled binaries
├── data/                          # Data storage directory
│   ├── earthquakes/               # Earthquake data files
│   └── faults/                    # Fault data files
├── go.mod                         # Go module definition
├── go.sum                         # Go module checksums
└── Makefile                       # Build and development tasks
```

### Core Components

#### 1. API Clients (`internal/api/`)
- **USGS Client**: Interfaces with USGS FDSNWS API for earthquake data
- **EMSC Client**: Interfaces with EMSC-CSEM API for fault data

#### 2. Data Collectors (`internal/collector/`)
- **Earthquake Collector**: Orchestrates earthquake data collection
- **Fault Collector**: Orchestrates fault data collection

#### 3. Storage Layer (`internal/storage/`)
- **JSON Storage**: File-based storage in JSON format
- **PostgreSQL Storage**: Database storage with full CRUD operations
- **Migration System**: Database schema management

#### 4. Scheduler (`internal/scheduler/`)
- **Interval Management**: Configurable collection intervals
- **Daemon Support**: Background process management
- **Health Monitoring**: System health checks
- **Retry Logic**: Exponential backoff for failed operations

#### 5. CLI Interface (`pkg/cli/`)
- **Command Structure**: Organized command hierarchy
- **Configuration Management**: Dynamic configuration loading
- **Output Formatting**: Structured output in JSON format

## 🔧 Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Testing

```bash
# Run all tests
make test

# Test specific package
go test ./internal/storage/...

# Run with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code
make fmt

# Generate documentation
make docs
```

### Database Development

```bash
# Set up development database
make db-setup-docker

# Run migrations
make db-migrate-up

# Check status
make db-status
```

## 📊 Data Sources

### USGS FDSNWS API

- **Base URL**: https://earthquake.usgs.gov/fdsnws/event/1
- **Format**: GeoJSON, CSV, XML, QuakeML
- **Update Frequency**: Real-time (every 5-15 minutes)
- **Rate Limit**: 60 requests per minute
- **Data Types**: Earthquake events with magnitude, location, timing, and metadata

### EMSC-CSEM API

- **Base URL**: https://www.emsc-csem.org/javascript
- **Format**: GeoJSON
- **Update Frequency**: Variable
- **Data Types**: Active fault data with geographical coordinates and properties

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards
- Add tests for new functionality
- Update documentation for API changes
- Use conventional commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:

- Create an issue in the repository
- Check the [documentation](documentation/) folder
- Review the [API Reference](documentation/API_REFERENCE.md)

## 🔄 Changelog

See [CHANGELOG.md](documentation/CHANGELOG.md) for a detailed history of changes. 
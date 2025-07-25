# QuakeWatch Data Scraper

A powerful Go application for collecting earthquake and fault data from various seismological sources with advanced interval scraping capabilities and PostgreSQL database storage.

## 🚀 Key Features

- **🌍 Earthquake Data Collection**: Fetch earthquake data from USGS FDSNWS API
- **⏰ Interval Scraping**: Run data collection at specified intervals with monitoring
- **🗄️ PostgreSQL Storage**: Store data in PostgreSQL database with advanced querying
- **🔄 Smart Collection**: Avoid duplicates with time-based filtering and metadata tracking
- **👻 Daemon Mode**: Run in background as a daemon process
- **📊 Health Monitoring**: Comprehensive system health checks and metrics
- **🛡️ Error Handling**: Robust error handling with exponential backoff
- **📁 JSON Storage**: Save all data to timestamped JSON files
- **🌐 Fault Data Collection**: Fetch fault data from EMSC-CSEM API
- **🔍 Advanced Filtering**: Filter by country, region, magnitude, and time range
- **📈 Data Validation**: Built-in data validation and statistics
- **🔄 Database Migrations**: Automated database schema management

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Interval Scraping](#interval-scraping)
- [Database Storage](#database-storage)
- [Basic Usage](#basic-usage)
- [Advanced Features](#advanced-features)
- [Configuration](#configuration)
- [Examples](#examples)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## 🛠️ Installation

### Prerequisites

- Go 1.21 or later
- Git
- PostgreSQL (optional, for database storage)

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd quakewatch-scraper

# Install dependencies
make install

# Build the application
make build

# Create necessary directories
make setup
```

### Database Setup (Optional)

If you want to use PostgreSQL storage:

```bash
# Setup PostgreSQL with Docker
make db-setup-docker

# Run database migrations
make db-migrate-up

# Check migration status
make db-status
```

## 🚀 Quick Start

### Basic Data Collection

```bash
# Collect recent earthquakes
./bin/quakewatch-scraper earthquakes recent

# Collect earthquakes and save to database
./bin/quakewatch-scraper earthquakes recent --storage postgresql

# Collect fault data
./bin/quakewatch-scraper faults collect --storage postgresql
```

### Interval Scraping (Recommended)

```bash
# Collect earthquakes every 5 minutes and save to database
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --smart

# Run in background as daemon
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --daemon \
  --pid-file /var/run/quakewatch-scraper.pid
```

## ⏰ Interval Scraping

The interval scraping feature is the core functionality for continuous data collection. It allows you to run data collection commands at specified intervals with comprehensive monitoring, error handling, and scheduling capabilities.

### Basic Interval Scraping

```bash
# Collect recent earthquakes every 5 minutes
./bin/quakewatch-scraper interval earthquakes recent --interval 5m

# Collect significant earthquakes every hour
./bin/quakewatch-scraper interval earthquakes significant --interval 1h

# Collect fault data every 12 hours
./bin/quakewatch-scraper interval faults collect --interval 12h
```

### Production-Ready Interval Scraping

```bash
# Production setup with comprehensive monitoring
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --max-runtime 168h \
  --backoff exponential \
  --max-backoff 30m \
  --continue-on-error \
  --health-check-interval 10m \
  --daemon \
  --pid-file /var/run/quakewatch-scraper.pid \
  --log-file /var/log/quakewatch-scraper.log \
  --smart
```

### Time Format Options

The `--interval` flag accepts various time formats:

- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `6h` - 6 hours
- `1d` - 1 day
- `1w` - 1 week

### Advanced Interval Features

#### Error Handling and Retries

```bash
# Continue running even if commands fail
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --continue-on-error

# Use exponential backoff for failed operations
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --backoff exponential \
  --max-backoff 30m
```

#### Limited Runtime and Executions

```bash
# Run for maximum 24 hours
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --max-runtime 24h

# Run maximum 100 times
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --max-executions 100
```

#### Custom Multi-Command Workflows

```bash
# Run multiple commands in sequence
./bin/quakewatch-scraper interval custom \
  --interval 1h \
  --storage postgresql \
  --commands "earthquakes recent,earthquakes significant --start 2024-01-01 --end 2024-01-31,faults collect"
```

### Daemon Mode

Run the interval scraper in the background as a daemon process:

```bash
# Start daemon
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --daemon \
  --pid-file /var/run/quakewatch-scraper.pid \
  --log-file /var/log/quakewatch-scraper.log

# Check if daemon is running
ps aux | grep quakewatch-scraper

# Stop daemon
kill $(cat /var/run/quakewatch-scraper.pid)
```

## 🗄️ Database Storage

The application supports PostgreSQL as a storage backend, providing structured data storage with advanced querying capabilities.

### Database Setup

```bash
# Setup PostgreSQL with Docker
make db-setup-docker

# Run database migrations
make db-migrate-up

# Check migration status
make db-status
```

### Database Operations

```bash
# Initialize database (creates tables and indexes)
./bin/quakewatch-scraper db init

# Check database status
./bin/quakewatch-scraper db status

# Run migrations
./bin/quakewatch-scraper db migrate up

# Rollback migrations
./bin/quakewatch-scraper db migrate down
```

### Data Collection with Database Storage

```bash
# Collect earthquakes and save to database
./bin/quakewatch-scraper earthquakes recent --storage postgresql

# Collect faults and save to database
./bin/quakewatch-scraper faults collect --storage postgresql

# Query earthquakes from database
./bin/quakewatch-scraper earthquakes query --storage postgresql --limit 100
```

### Advanced Database Queries

```bash
# Query by time range
./bin/quakewatch-scraper earthquakes query \
  --storage postgresql \
  --start-time "2024-01-01T00:00:00Z" \
  --end-time "2024-01-02T00:00:00Z"

# Query by magnitude range
./bin/quakewatch-scraper earthquakes query \
  --storage postgresql \
  --min-magnitude 4.5 \
  --max-magnitude 10.0

# Query by geographic region
./bin/quakewatch-scraper earthquakes query \
  --storage postgresql \
  --min-lat 32 --max-lat 42 \
  --min-lon -125 --max-lon -114
```

## 📖 Basic Usage

### Basic Commands

```bash
# Show version information
./bin/quakewatch-scraper version

# Check system health
./bin/quakewatch-scraper health

# Show help
./bin/quakewatch-scraper help

# Quick data preview (output to terminal)
./bin/quakewatch-scraper earthquakes recent --stdout --limit 3
```

### Earthquake Data Collection

```bash
# Collect recent earthquakes (last hour)
./bin/quakewatch-scraper earthquakes recent

# Collect recent earthquakes with limit
./bin/quakewatch-scraper earthquakes recent --limit 100

# Collect earthquakes by time range
./bin/quakewatch-scraper earthquakes time-range --start "2024-01-01" --end "2024-01-02"

# Collect earthquakes by magnitude range
./bin/quakewatch-scraper earthquakes magnitude --min 4.5 --max 10.0

# Collect significant earthquakes (M4.5+)
./bin/quakewatch-scraper earthquakes significant --start "2024-01-01" --end "2024-01-02"

# Collect earthquakes by geographic region
./bin/quakewatch-scraper earthquakes region --min-lat 32 --max-lat 42 --min-lon -125 --max-lon -114

# Collect earthquakes by country
./bin/quakewatch-scraper earthquakes country --country "Japan" --min-mag 4.0
```

### Smart Collection

The smart collection feature prevents duplicate data collection by tracking the last collection time:

```bash
# Smart collection (avoids duplicates)
./bin/quakewatch-scraper earthquakes recent --smart --storage postgresql

# Time-based collection (last 3 hours)
./bin/quakewatch-scraper earthquakes recent --hours-back 3 --storage postgresql

# Smart collection with JSON storage
./bin/quakewatch-scraper earthquakes recent --smart --storage json
```

### Fault Data Collection

```bash
# Collect fault data from EMSC
./bin/quakewatch-scraper faults collect

# Update fault data with retry logic
./bin/quakewatch-scraper faults update --retries 5 --retry-delay 10s
```

### Data Management

```bash
# List available data files
./bin/quakewatch-scraper list

# List earthquake files only
./bin/quakewatch-scraper list --type earthquakes

# Show data statistics
./bin/quakewatch-scraper stats

# Show statistics for specific file
./bin/quakewatch-scraper stats --file earthquakes_2024-01-01_15-04-05.json

# Validate data integrity
./bin/quakewatch-scraper validate

# Validate specific file
./bin/quakewatch-scraper validate --file earthquakes_2024-01-01_15-04-05.json
```

### Output to Standard Output

The `--stdout` flag allows you to output data directly to the terminal:

```bash
# Output recent earthquakes to stdout
./bin/quakewatch-scraper earthquakes recent --stdout --limit 5

# Output earthquakes by country to stdout
./bin/quakewatch-scraper earthquakes country --country "Japan" --stdout

# Pipe to jq for filtering and formatting
./bin/quakewatch-scraper earthquakes recent --stdout | jq '.features[] | select(.properties.mag > 4.0)'

# Count earthquakes in a region
./bin/quakewatch-scraper earthquakes region --min-lat 32 --max-lat 42 --min-lon -125 --max-lon -114 --stdout | jq '.features | length'
```

## 🔧 Advanced Features

### Advanced Options

```bash
# Use custom output directory
./bin/quakewatch-scraper earthquakes recent --output-dir /path/to/data

# Enable verbose logging
./bin/quakewatch-scraper earthquakes recent --verbose --log-level debug

# Dry run to see what would be collected
./bin/quakewatch-scraper earthquakes recent --dry-run

# Custom filename
./bin/quakewatch-scraper earthquakes recent --filename my_earthquakes

# Use configuration file
./bin/quakewatch-scraper earthquakes recent --config ./configs/config.yaml
```

### Health Monitoring

```bash
# Check all system components
./bin/quakewatch-scraper health

# Check database health
./bin/quakewatch-scraper health --storage postgresql

# View collection logs
./bin/quakewatch-scraper db logs --limit 10
```

## ⚙️ Configuration

The application uses a YAML configuration file (`configs/config.yaml`):

```yaml
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

database:
  host: "localhost"
  port: 5432
  user: "quakewatch_user"
  password: "your_password"
  name: "quakewatch"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"
  conn_max_idle_time: "5m"

interval:
  default_interval: "1h"
  max_runtime: "24h"
  max_executions: 0
  backoff_strategy: "exponential"
  max_backoff: "30m"
  continue_on_error: false
  skip_empty: false
  health_check_interval: "5m"
  daemon_mode: false
  pid_file: "./quakewatch-scraper.pid"
  log_file: "./quakewatch-scraper.log"
```

### Environment Variables

For PostgreSQL storage, you can also use environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=quakewatch_user
DB_PASSWORD=your_password
DB_NAME=quakewatch
DB_SSL_MODE=disable

# Connection Pool Settings (optional)
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=5m
```

## 📝 Examples

### Production Setup Examples

#### Continuous Earthquake Monitoring

```bash
# Monitor recent earthquakes every 5 minutes
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --storage postgresql \
  --smart \
  --daemon \
  --pid-file /var/run/quakewatch-scraper.pid \
  --log-file /var/log/quakewatch-scraper.log
```

#### Regional Earthquake Monitoring

```bash
# Monitor California earthquakes every 15 minutes
./bin/quakewatch-scraper interval earthquakes region \
  --interval 15m \
  --storage postgresql \
  --min-lat 32.0 \
  --max-lat 42.0 \
  --min-lon -125.0 \
  --max-lon -114.0 \
  --smart
```

#### Significant Earthquake Monitoring

```bash
# Monitor significant earthquakes every hour
./bin/quakewatch-scraper interval earthquakes significant \
  --interval 1h \
  --storage postgresql \
  --start "2024-01-01" \
  --end "2024-12-31"
```

### Development Examples

#### Quick Testing

```bash
# Test recent earthquake collection
./bin/quakewatch-scraper earthquakes recent --stdout --limit 5

# Test interval scraping for 10 minutes
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 30s \
  --max-executions 20 \
  --storage postgresql
```

#### Data Analysis

```bash
# Query earthquakes from database
./bin/quakewatch-scraper earthquakes query \
  --storage postgresql \
  --min-magnitude 4.0 \
  --limit 100

# Export data for analysis
./bin/quakewatch-scraper earthquakes query \
  --storage postgresql \
  --start-time "2024-01-01T00:00:00Z" \
  --end-time "2024-01-31T23:59:59Z" \
  --stdout > january_earthquakes.json
```

## 📊 Monitoring

### Health Checks

```bash
# Check system health
./bin/quakewatch-scraper health

# Check database health
./bin/quakewatch-scraper health --storage postgresql

# Check specific components
./bin/quakewatch-scraper health --storage postgresql --api usgs
```

### Logging and Metrics

The interval scraper provides detailed logging:

```
[2024-01-15 10:00:00] Starting interval scraper
[2024-01-15 10:00:00] Configuration: interval=5m, max-runtime=24h, backoff=exponential
[2024-01-15 10:00:00] Executing command: earthquakes recent
[2024-01-15 10:00:00] Found 15 earthquakes
[2024-01-15 10:00:00] Successfully collected and saved 15 earthquakes
[2024-01-15 10:00:00] Next execution in 5 minutes
```

### Database Statistics

```bash
# Get database statistics
./bin/quakewatch-scraper db stats

# Monitor collection logs
./bin/quakewatch-scraper db logs --limit 10
```

## 🔧 Troubleshooting

### Common Issues

1. **API Connection Errors**
   - Check internet connectivity
   - Verify API endpoints are accessible
   - Check rate limiting settings

2. **Database Connection Issues**
   - Verify PostgreSQL is running
   - Check database credentials
   - Ensure migrations have been run

3. **File Permission Errors**
   - Ensure write permissions to output directory
   - Check disk space availability

4. **Interval Scraper Issues**
   - Check PID file permissions
   - Verify log file directory exists
   - Monitor system resources

### Debug Mode

```bash
# Enable verbose logging
./bin/quakewatch-scraper earthquakes recent --verbose --log-level debug

# Debug interval scraper
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --verbose \
  --log-level debug
```

### Health Check

```bash
# Run comprehensive health check
./bin/quakewatch-scraper health
```

## 📚 Additional Documentation

- **[DATABASE.md](DATABASE.md)**: Complete PostgreSQL database implementation guide
- **[INTERVAL_README.md](INTERVAL_README.md)**: Detailed interval scraping documentation
- **[SMART_COLLECTION.md](SMART_COLLECTION.md)**: Smart collection feature documentation
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)**: Implementation summary and overview

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For issues and questions:
- Create an issue in the repository
- Check the troubleshooting section
- Review the configuration options
- Consult the additional documentation files

## 🗺️ Roadmap

- [x] Database integration (PostgreSQL)
- [x] Smart collection features
- [x] Database migrations
- [x] Health monitoring
- [x] Interval scraping with monitoring
- [ ] Real-time monitoring capabilities
- [ ] Advanced filtering and processing
- [ ] API server functionality
- [ ] Comprehensive testing suite
- [ ] Docker containerization
- [ ] Kubernetes deployment
- [ ] Metrics and monitoring (Prometheus)
- [ ] Web interface
- [ ] Data visualization tools 
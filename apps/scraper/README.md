# QuakeWatch Data Scraper

A Go application for collecting earthquake and fault data from various seismological sources and saving to JSON files.

## Features

- **Earthquake Data Collection**: Fetch earthquake data from USGS FDSNWS API
- **Country Filtering**: Filter earthquakes by country or region
- **Fault Data Collection**: Fetch fault data from EMSC-CSEM API
- **JSON Storage**: Save all data to timestamped JSON files
- **Standard Output**: Output data directly to terminal with `--stdout` flag
- **Command Line Interface**: Easy-to-use CLI with various collection options
- **Data Validation**: Built-in data validation and statistics
- **Cross-platform**: Single binary for Linux, macOS, and Windows

## Installation

### Prerequisites

- Go 1.21 or later
- Git

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

### Cross-platform Builds

```bash
# Build for all platforms
make build-all

# Or build for specific platform
make build-linux
make build-darwin
make build-windows
```

## Usage

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

# Delete all data files (with confirmation)
./bin/quakewatch-scraper purge

# Delete specific data type
./bin/quakewatch-scraper purge --type earthquakes

# Force delete without confirmation
./bin/quakewatch-scraper purge --force

# Show what would be deleted (dry run)
./bin/quakewatch-scraper purge --dry-run

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

### Output to Standard Output

The `--stdout` flag allows you to output data directly to the terminal instead of saving to files. This is useful for:

- **Piping to other tools**: Process data with `jq`, `grep`, or other command-line tools
- **Real-time processing**: View data immediately without file I/O overhead
- **Integration with scripts**: Capture output for further processing
- **Debugging**: Quickly inspect data structure and content

```bash
# Output recent earthquakes to stdout
./bin/quakewatch-scraper earthquakes recent --stdout --limit 5

# Output earthquakes by country to stdout
./bin/quakewatch-scraper earthquakes country --country "Japan" --stdout

# Output earthquakes by magnitude range to stdout
./bin/quakewatch-scraper earthquakes magnitude --min 4.0 --max 5.0 --stdout

# Output fault data to stdout
./bin/quakewatch-scraper faults collect --stdout

# Pipe to jq for filtering and formatting
./bin/quakewatch-scraper earthquakes recent --stdout | jq '.features[] | select(.properties.mag > 4.0)'

# Count earthquakes in a region
./bin/quakewatch-scraper earthquakes region --min-lat 32 --max-lat 42 --min-lon -125 --max-lon -114 --stdout | jq '.features | length'

# Extract specific earthquake properties
./bin/quakewatch-scraper earthquakes recent --stdout | jq '.features[] | {magnitude: .properties.mag, place: .properties.place, time: .properties.time}'

# Save stdout output to a custom file
./bin/quakewatch-scraper earthquakes recent --stdout > my_earthquakes.json

# Combine with other tools for analysis
./bin/quakewatch-scraper earthquakes recent --stdout | jq -r '.features[] | "\(.properties.mag) \(.properties.place)"' | sort -n
```

## Configuration

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
```

## Data Sources

### USGS Earthquake API
- **Endpoint**: https://earthquake.usgs.gov/fdsnws/event/1/
- **Format**: GeoJSON
- **Update Frequency**: Real-time (every 5-15 minutes)
- **Documentation**: https://earthquake.usgs.gov/fdsnws/event/1/

### EMSC-CSEM Fault API
- **Endpoint**: https://www.emsc-csem.org/javascript/gem_active_faults.geojson
- **Format**: GeoJSON
- **Content**: Active fault data with geographical coordinates and properties

## Data Structure

### Earthquake Data
Earthquake data is stored in GeoJSON format with the following structure:

```json
{
  "type": "FeatureCollection",
  "metadata": {
    "generated": 1640995200000,
    "url": "https://earthquake.usgs.gov/fdsnws/event/1/query",
    "title": "USGS Earthquakes",
    "status": 200,
    "api": "1.10.3",
    "count": 1
  },
  "features": [
    {
      "type": "Feature",
      "properties": {
        "mag": 4.5,
        "place": "10km ENE of Somewhere, CA",
        "time": 1640995200000,
        "updated": 1640995300000,
        "url": "https://earthquake.usgs.gov/earthquakes/eventpage/...",
        "detail": "https://earthquake.usgs.gov/fdsnws/event/1/query...",
        "felt": 25,
        "cdi": 3.4,
        "mmi": 4.2,
        "alert": "green",
        "status": "reviewed",
        "tsunami": 0,
        "sig": 312,
        "net": "us",
        "code": "7000abcd",
        "ids": ",us7000abcd,",
        "sources": ",us,",
        "types": ",dyfi,origin,phase-data,",
        "nst": 45,
        "dmin": 0.123,
        "rms": 0.67,
        "gap": 45,
        "magType": "mb",
        "type": "earthquake",
        "title": "M 4.5 - 10km ENE of Somewhere, CA"
      },
      "geometry": {
        "type": "Point",
        "coordinates": [-117.1234, 34.5678, 12.3]
      },
      "id": "us7000abcd"
    }
  ]
}
```

### Fault Data
Fault data is stored in GeoJSON format with the following structure:

```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "properties": {
        "id": "fault_001",
        "name": "San Andreas Fault",
        "type": "strike-slip",
        "slip_rate": 25.5,
        "slip_type": "right-lateral",
        "dip": 90.0,
        "rake": 0.0,
        "length": 1200.0,
        "width": 15.0,
        "max_magnitude": 8.0,
        "description": "Major strike-slip fault in California",
        "source": "EMSC-CSEM"
      },
      "geometry": {
        "type": "LineString",
        "coordinates": [
          [-121.0, 36.0],
          [-120.5, 36.2],
          [-120.0, 36.5]
        ]
      }
    }
  ]
}
```

## Development

### Project Structure

```
quakewatch-scraper/
├── cmd/
│   └── scraper/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── usgs.go             # USGS API client
│   │   └── emsc.go             # EMSC API client
│   ├── models/
│   │   ├── earthquake.go       # Earthquake data models
│   │   └── fault.go            # Fault data models
│   ├── collector/
│   │   ├── earthquake.go       # Earthquake data collector
│   │   └── fault.go            # Fault data collector
│   ├── storage/
│   │   └── json.go             # JSON file storage
│   └── utils/
│       └── logger.go           # Logging utilities
├── pkg/
│   └── cli/
│       └── commands.go         # CLI commands
├── configs/
│   └── config.yaml             # Configuration file
├── data/
│   ├── earthquakes/            # Earthquake data files
│   └── faults/                 # Fault data files
├── go.mod                      # Go module file
├── go.sum                      # Go module checksums
├── Makefile                    # Build and development tasks
└── README.md                   # This file
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Quality

```bash
# Format code
make fmt

# Lint code (requires golangci-lint)
make lint

# Generate documentation
make docs
```

### Quick Testing

```bash
# Test the application
make test-app

# Test earthquake collection
make test-earthquakes

# Test fault collection
make test-faults
```

## Examples

### Collect Recent Earthquakes

```bash
# Collect last 50 earthquakes
./bin/quakewatch-scraper earthquakes recent --limit 50 --filename recent_quakes
```

### Collect Historical Data

```bash
# Collect earthquakes for a specific date range
./bin/quakewatch-scraper earthquakes time-range \
  --start "2024-01-01" \
  --end "2024-01-31" \
  --limit 5000 \
  --filename january_2024_quakes
```

### Collect Significant Earthquakes

```bash
# Collect significant earthquakes for the past year
./bin/quakewatch-scraper earthquakes significant \
  --start "2023-01-01" \
  --end "2024-01-01" \
  --limit 1000 \
  --filename significant_2023
```

### Collect Regional Data

```bash
# Collect earthquakes in California region
./bin/quakewatch-scraper earthquakes region \
  --min-lat 32.0 \
  --max-lat 42.0 \
  --min-lon -125.0 \
  --max-lon -114.0 \
  --limit 1000 \
  --filename california_quakes
```

### Monitor System Health

```bash
# Check all system components
./bin/quakewatch-scraper health

# List available data files
./bin/quakewatch-scraper list

# Show data statistics
./bin/quakewatch-scraper stats --type earthquakes
```

## Troubleshooting

### Common Issues

1. **API Connection Errors**
   - Check internet connectivity
   - Verify API endpoints are accessible
   - Check rate limiting settings

2. **File Permission Errors**
   - Ensure write permissions to output directory
   - Check disk space availability

3. **Invalid Date Formats**
   - Use YYYY-MM-DD format for dates
   - Ensure dates are in chronological order

### Debug Mode

```bash
# Enable verbose logging
./bin/quakewatch-scraper earthquakes recent --verbose --log-level debug
```

### Health Check

```bash
# Run comprehensive health check
./bin/quakewatch-scraper health
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
- Create an issue in the repository
- Check the troubleshooting section
- Review the configuration options

## Roadmap

- [ ] Database integration (PostgreSQL/MySQL)
- [ ] Real-time monitoring capabilities
- [ ] Advanced filtering and processing
- [ ] API server functionality
- [ ] Comprehensive testing suite
- [ ] Docker containerization
- [ ] Kubernetes deployment
- [ ] Metrics and monitoring (Prometheus)
- [ ] Web interface
- [ ] Data visualization tools 
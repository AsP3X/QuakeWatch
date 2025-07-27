# API Reference

This document provides a comprehensive reference for all commands, options, and configurations available in the QuakeWatch Scraper.

## Table of Contents

- [Command Overview](#command-overview)
- [Global Options](#global-options)
- [Earthquake Commands](#earthquake-commands)
- [Fault Commands](#fault-commands)
- [Interval Commands](#interval-commands)
- [Database Commands](#database-commands)
- [Utility Commands](#utility-commands)
- [Configuration Reference](#configuration-reference)
- [Data Models](#data-models)

## Command Overview

The QuakeWatch Scraper uses a hierarchical command structure:

```bash
quakewatch-scraper [global-options] <command> [command-options] [arguments]
```

### Main Command Groups

- `earthquakes` - Earthquake data collection commands
- `faults` - Fault data collection commands
- `interval` - Scheduled collection commands
- `db` - Database management commands
- Utility commands: `version`, `health`, `stats`, `validate`, `list`, `purge`

## Global Options

These options are available for all commands:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--config` | string | `config.yaml` | Path to configuration file |
| `--output` | string | `stdout` | Output file path |
| `--format` | string | `json` | Output format (json, csv) |
| `--verbose` | bool | `false` | Enable verbose logging |
| `--quiet` | bool | `false` | Suppress output except errors |

## Earthquake Commands

### `earthquakes recent`

Collect recent earthquakes from the last hour.

```bash
quakewatch-scraper earthquakes recent [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--limit` | int | `1000` | Maximum number of earthquakes to collect |
| `--min-magnitude` | float | `0.0` | Minimum magnitude filter |
| `--max-magnitude` | float | `10.0` | Maximum magnitude filter |
| `--output` | string | `stdout` | Output file path |

#### Examples

```bash
# Collect last 50 earthquakes
quakewatch-scraper earthquakes recent --limit 50

# Collect earthquakes with magnitude >= 4.0
quakewatch-scraper earthquakes recent --min-magnitude 4.0

# Save to file
quakewatch-scraper earthquakes recent --output recent_quakes.json
```

### `earthquakes time-range`

Collect earthquakes within a specific time range.

```bash
quakewatch-scraper earthquakes time-range [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--start` | string | Yes | Start time (ISO 8601 or YYYY-MM-DD) |
| `--end` | string | Yes | End time (ISO 8601 or YYYY-MM-DD) |
| `--limit` | int | `1000` | Maximum number of earthquakes |
| `--min-magnitude` | float | `0.0` | Minimum magnitude filter |
| `--max-magnitude` | float | `10.0` | Maximum magnitude filter |

#### Examples

```bash
# Collect earthquakes for a specific date range
quakewatch-scraper earthquakes time-range \
    --start "2024-01-01" \
    --end "2024-01-31"

# Use ISO 8601 timestamps
quakewatch-scraper earthquakes time-range \
    --start "2024-01-01T00:00:00Z" \
    --end "2024-01-02T00:00:00Z"
```

### `earthquakes magnitude`

Collect earthquakes within a magnitude range.

```bash
quakewatch-scraper earthquakes magnitude [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--min` | float | Yes | Minimum magnitude |
| `--max` | float | Yes | Maximum magnitude |
| `--start` | string | Yes | Start time |
| `--end` | string | Yes | End time |
| `--limit` | int | `1000` | Maximum number of earthquakes |

#### Examples

```bash
# Collect significant earthquakes
quakewatch-scraper earthquakes magnitude \
    --min 4.5 \
    --max 10.0 \
    --start "2024-01-01" \
    --end "2024-01-31"
```

### `earthquakes significant`

Collect significant earthquakes (magnitude >= 4.5).

```bash
quakewatch-scraper earthquakes significant [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--start` | string | Yes | Start time |
| `--end` | string | Yes | End time |
| `--limit` | int | `1000` | Maximum number of earthquakes |

#### Examples

```bash
# Collect significant earthquakes for January 2024
quakewatch-scraper earthquakes significant \
    --start "2024-01-01" \
    --end "2024-01-31"
```

### `earthquakes region`

Collect earthquakes within a geographic region.

```bash
quakewatch-scraper earthquakes region [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--min-lat` | float | Yes | Minimum latitude |
| `--max-lat` | float | Yes | Maximum latitude |
| `--min-lon` | float | Yes | Minimum longitude |
| `--max-lon` | float | Yes | Maximum longitude |
| `--start` | string | Yes | Start time |
| `--end` | string | Yes | End time |
| `--limit` | int | `1000` | Maximum number of earthquakes |

#### Examples

```bash
# Collect earthquakes in California
quakewatch-scraper earthquakes region \
    --min-lat 32.0 \
    --max-lat 42.0 \
    --min-lon -125.0 \
    --max-lon -114.0 \
    --start "2024-01-01" \
    --end "2024-01-31"
```

### `earthquakes country`

Collect earthquakes for a specific country.

```bash
quakewatch-scraper earthquakes country [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--country` | string | Yes | Country name |
| `--start` | string | Yes | Start time |
| `--end` | string | Yes | End time |
| `--limit` | int | `1000` | Maximum number of earthquakes |

#### Examples

```bash
# Collect earthquakes in Japan
quakewatch-scraper earthquakes country \
    --country "Japan" \
    --start "2024-01-01" \
    --end "2024-01-31"
```

## Fault Commands

### `faults collect`

Collect fault data from EMSC-CSEM.

```bash
quakewatch-scraper faults collect [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--output` | string | `stdout` | Output file path |
| `--force` | bool | `false` | Force collection even if recent data exists |

#### Examples

```bash
# Collect fault data
quakewatch-scraper faults collect

# Save to specific file
quakewatch-scraper faults collect --output faults.json

# Force collection
quakewatch-scraper faults collect --force
```

### `faults update`

Update existing fault data.

```bash
quakewatch-scraper faults update [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--output` | string | `stdout` | Output file path |

#### Examples

```bash
# Update fault data
quakewatch-scraper faults update
```

## Interval Commands

### `interval earthquakes recent`

Schedule collection of recent earthquakes at regular intervals.

```bash
quakewatch-scraper interval earthquakes recent [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--interval` | duration | Yes | Collection interval (e.g., 5m, 1h, 24h) |
| `--duration` | duration | Yes | Total duration to run |
| `--limit` | int | `1000` | Maximum earthquakes per collection |
| `--output-dir` | string | `./data/earthquakes` | Output directory |

#### Examples

```bash
# Collect every 5 minutes for 1 hour
quakewatch-scraper interval earthquakes recent \
    --interval 5m \
    --duration 1h

# Collect every hour for 24 hours
quakewatch-scraper interval earthquakes recent \
    --interval 1h \
    --duration 24h
```

### `interval earthquakes significant`

Schedule collection of significant earthquakes.

```bash
quakewatch-scraper interval earthquakes significant [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--interval` | duration | Yes | Collection interval |
| `--duration` | duration | Yes | Total duration to run |
| `--start` | string | Yes | Start time for data collection |
| `--end` | string | Yes | End time for data collection |
| `--limit` | int | `1000` | Maximum earthquakes per collection |

#### Examples

```bash
# Collect significant earthquakes every hour for a month
quakewatch-scraper interval earthquakes significant \
    --interval 1h \
    --duration 720h \
    --start "2024-01-01" \
    --end "2024-01-31"
```

### `interval faults collect`

Schedule fault data collection.

```bash
quakewatch-scraper interval faults collect [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--interval` | duration | Yes | Collection interval |
| `--duration` | duration | Yes | Total duration to run |
| `--output-dir` | string | `./data/faults` | Output directory |

#### Examples

```bash
# Collect fault data daily for a week
quakewatch-scraper interval faults collect \
    --interval 24h \
    --duration 168h
```

## Database Commands

### `db init`

Initialize the database connection and create tables.

```bash
quakewatch-scraper db init [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--force` | bool | `false` | Force reinitialization |

#### Examples

```bash
# Initialize database
quakewatch-scraper db init

# Force reinitialization
quakewatch-scraper db init --force
```

### `db migrate up`

Apply database migrations.

```bash
quakewatch-scraper db migrate up [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--steps` | int | `0` | Number of migrations to apply (0 = all) |

#### Examples

```bash
# Apply all pending migrations
quakewatch-scraper db migrate up

# Apply next 2 migrations
quakewatch-scraper db migrate up --steps 2
```

### `db migrate down`

Rollback database migrations.

```bash
quakewatch-scraper db migrate down [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--steps` | int | `1` | Number of migrations to rollback |

#### Examples

```bash
# Rollback last migration
quakewatch-scraper db migrate down

# Rollback 3 migrations
quakewatch-scraper db migrate down --steps 3
```

### `db status`

Show database migration status.

```bash
quakewatch-scraper db status
```

#### Examples

```bash
# Check migration status
quakewatch-scraper db status
```

### `db force-version`

Force migration version.

```bash
quakewatch-scraper db force-version <version>
```

#### Arguments

| Argument | Type | Description |
|----------|------|-------------|
| `version` | int | Migration version to force |

#### Examples

```bash
# Force migration version to 1
quakewatch-scraper db force-version 1
```

## Utility Commands

### `version`

Display application version information.

```bash
quakewatch-scraper version
```

### `health`

Check application and database health.

```bash
quakewatch-scraper health [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--check-db` | bool | `false` | Include database health check |

#### Examples

```bash
# Check application health
quakewatch-scraper health

# Check application and database health
quakewatch-scraper health --check-db
```

### `stats`

Display collection statistics.

```bash
quakewatch-scraper stats [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--start` | string | | Start time for statistics |
| `--end` | string | | End time for statistics |
| `--type` | string | `all` | Data type (earthquakes, faults, all) |

#### Examples

```bash
# Show all statistics
quakewatch-scraper stats

# Show statistics for specific period
quakewatch-scraper stats \
    --start "2024-01-01" \
    --end "2024-01-31"

# Show earthquake statistics only
quakewatch-scraper stats --type earthquakes
```

### `validate`

Validate collected data.

```bash
quakewatch-scraper validate [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--file` | string | | Specific file to validate |
| `--type` | string | `all` | Data type to validate |

#### Examples

```bash
# Validate all data
quakewatch-scraper validate

# Validate specific file
quakewatch-scraper validate --file earthquakes.json

# Validate earthquake data only
quakewatch-scraper validate --type earthquakes
```

### `list`

List collected data files.

```bash
quakewatch-scraper list [options]
```

#### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `--type` | string | `all` | Data type to list |
| `--details` | bool | `false` | Show detailed information |

#### Examples

```bash
# List all data files
quakewatch-scraper list

# List earthquake files with details
quakewatch-scraper list --type earthquakes --details
```

### `purge`

Purge old data files.

```bash
quakewatch-scraper purge [options]
```

#### Options

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--older-than` | duration | Yes | Age threshold for deletion |
| `--dry-run` | bool | `false` | Show what would be deleted |
| `--type` | string | `all` | Data type to purge |

#### Examples

```bash
# Purge files older than 30 days
quakewatch-scraper purge --older-than 30d

# Dry run purge
quakewatch-scraper purge --older-than 30d --dry-run

# Purge earthquake files only
quakewatch-scraper purge --older-than 30d --type earthquakes
```

## Configuration Reference

### Configuration File Structure

```yaml
api:
    emsc:
        base_url: string          # EMSC API base URL
        timeout: duration         # Request timeout
    usgs:
        base_url: string          # USGS API base URL
        rate_limit: int           # Requests per minute
        timeout: duration         # Request timeout

collection:
    default_limit: int            # Default result limit
    max_limit: int               # Maximum result limit
    retry_attempts: int          # Number of retry attempts
    retry_delay: duration        # Delay between retries

database:
    type: string                 # Database type (postgres)
    host: string                 # Database host
    port: int                    # Database port
    username: string             # Database username
    password: string             # Database password
    database: string             # Database name
    ssl_mode: string             # SSL mode
    max_connections: int         # Maximum connections
    connection_timeout: duration # Connection timeout
    enabled: bool                # Enable database storage

logging:
    level: string                # Log level (debug, info, warn, error)
    format: string               # Log format (json, text)
    output: string               # Log output (stdout, stderr, file)

storage:
    output_dir: string           # Base output directory
    earthquakes_dir: string      # Earthquake data directory
    faults_dir: string           # Fault data directory
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `quakewatch` |
| `DB_SSL_MODE` | SSL mode | `disable` |
| `LOG_LEVEL` | Log level | `info` |
| `CONFIG_PATH` | Configuration file path | `config.yaml` |

## Data Models

### Earthquake Data Structure

```json
{
  "type": "Feature",
  "properties": {
    "mag": 4.5,
    "place": "10km ENE of Ridgecrest, CA",
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
    "net": "ci",
    "code": "12345678",
    "ids": ",ci12345678,",
    "sources": ",ci,",
    "types": ",dyfi,origin,phase-data,",
    "nst": 45,
    "dmin": 0.123,
    "rms": 0.45,
    "gap": 45.2,
    "magType": "ml",
    "type": "earthquake",
    "title": "M 4.5 - 10km ENE of Ridgecrest, CA"
  },
  "geometry": {
    "type": "Point",
    "coordinates": [-117.5, 35.7, 10.5]
  },
  "id": "ci12345678"
}
```

### Fault Data Structure

```json
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
  },
  "id": "fault_001"
}
```

### Collection Log Structure

```json
{
  "id": 1,
  "data_type": "earthquakes",
  "source": "usgs",
  "start_time": "2024-01-01T00:00:00Z",
  "end_time": "2024-01-01T00:05:00Z",
  "records_collected": 150,
  "status": "completed",
  "error_message": null,
  "created_at": "2024-01-01T00:05:00Z"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `E001` | Configuration file not found |
| `E002` | Invalid configuration format |
| `E003` | Database connection failed |
| `E004` | API request failed |
| `E005` | Data validation failed |
| `E006` | File operation failed |
| `E007` | Migration failed |
| `E008` | Invalid command arguments |
| `E009` | Health check failed |
| `E010` | Rate limit exceeded |

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Configuration error |
| `3` | Database error |
| `4` | API error |
| `5` | Validation error |
| `6` | File system error |
| `7` | Migration error |
| `8` | Invalid arguments |
| `9` | Health check failed | 
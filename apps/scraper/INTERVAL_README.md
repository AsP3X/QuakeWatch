# Interval Scraping Documentation

The interval scraping feature allows you to run data collection commands at specified intervals with comprehensive monitoring, error handling, and scheduling capabilities.

## Overview

Interval scraping is designed for continuous data collection scenarios where you need to:

- Collect earthquake data periodically (e.g., every 5 minutes, hourly, daily)
- Run fault data updates on a schedule
- Monitor collection processes with health checks
- Handle errors gracefully with retry logic
- Run in background as a daemon process
- Limit execution time and number of runs
- Use exponential backoff for failed operations

## Basic Usage

### Simple Interval Collection

```bash
# Collect recent earthquakes every 5 minutes
./bin/quakewatch-scraper interval earthquakes recent --interval 5m

# Collect significant earthquakes every hour
./bin/quakewatch-scraper interval earthquakes significant --interval 1h

# Collect fault data every 12 hours
./bin/quakewatch-scraper interval faults collect --interval 12h
```

### Time Format

The `--interval` flag accepts various time formats:

- `30s` - 30 seconds
- `5m` - 5 minutes
- `1h` - 1 hour
- `6h` - 6 hours
- `1d` - 1 day
- `1w` - 1 week

## Advanced Features

### Daemon Mode

Run the interval scraper in the background as a daemon process:

```bash
# Run in daemon mode
./bin/quakewatch-scraper interval earthquakes recent --interval 5m --daemon

# Daemon mode with custom PID file
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --daemon \
  --pid-file /var/run/quakewatch-scraper.pid
```

### Limited Runtime

Limit how long the interval scraper runs:

```bash
# Run for maximum 24 hours
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --max-runtime 24h

# Run for maximum 1 week
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 1h \
  --max-runtime 168h
```

### Maximum Executions

Limit the number of times the command runs:

```bash
# Run maximum 10 times
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --max-executions 10

# Run maximum 100 times
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 1h \
  --max-executions 100
```

### Error Handling

Configure how the scraper handles errors:

```bash
# Continue running even if commands fail
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --continue-on-error

# Skip empty collections (no data found)
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --skip-empty
```

### Backoff Strategies

Configure retry behavior when commands fail:

```bash
# Use exponential backoff
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --backoff exponential \
  --max-backoff 30m

# Use linear backoff
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --backoff linear \
  --max-backoff 10m

# No backoff (immediate retry)
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --backoff none
```

### Health Monitoring

Enable health checks during interval scraping:

```bash
# Health check every 10 minutes
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --health-check-interval 10m

# Health check every hour
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 1h \
  --health-check-interval 1h
```

## Command Types

### Earthquake Commands

All earthquake collection commands support interval mode:

```bash
# Recent earthquakes
./bin/quakewatch-scraper interval earthquakes recent --interval 5m

# Time range earthquakes
./bin/quakewatch-scraper interval earthquakes time-range \
  --interval 1h \
  --start "2024-01-01" \
  --end "2024-01-02"

# Magnitude range earthquakes
./bin/quakewatch-scraper interval earthquakes magnitude \
  --interval 6h \
  --min 4.5 \
  --max 10.0

# Significant earthquakes
./bin/quakewatch-scraper interval earthquakes significant \
  --interval 1h \
  --start "2024-01-01" \
  --end "2024-01-31"

# Regional earthquakes
./bin/quakewatch-scraper interval earthquakes region \
  --interval 2h \
  --min-lat 32 --max-lat 42 \
  --min-lon -125 --max-lon -114

# Country-specific earthquakes
./bin/quakewatch-scraper interval earthquakes country \
  --interval 4h \
  --country "Japan" \
  --min-mag 4.0
```

### Fault Commands

Fault collection commands also support interval mode:

```bash
# Collect fault data
./bin/quakewatch-scraper interval faults collect --interval 12h

# Update fault data
./bin/quakewatch-scraper interval faults update \
  --interval 24h \
  --retries 5 \
  --retry-delay 10s
```

### Custom Commands

Run multiple commands in sequence:

```bash
# Run multiple earthquake collection commands
./bin/quakewatch-scraper interval custom \
  --interval 1h \
  --commands "earthquakes recent,earthquakes significant --start 2024-01-01 --end 2024-01-31"

# Run earthquake and fault collection
./bin/quakewatch-scraper interval custom \
  --interval 6h \
  --commands "earthquakes recent,faults collect"
```

## Configuration Options

### Interval Configuration

The interval scraper supports extensive configuration through the `interval` section in `config.yaml`:

```yaml
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

### Command Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--interval` | string | 1h | Time interval between executions |
| `--max-runtime` | string | 24h | Maximum time to run |
| `--max-executions` | int | 0 | Maximum number of executions (0 = unlimited) |
| `--backoff` | string | exponential | Backoff strategy (exponential, linear, none) |
| `--max-backoff` | string | 30m | Maximum backoff time |
| `--continue-on-error` | bool | false | Continue running on errors |
| `--skip-empty` | bool | false | Skip empty collections |
| `--health-check-interval` | string | 5m | Health check interval |
| `--daemon` | bool | false | Run in daemon mode |
| `--pid-file` | string | ./quakewatch-scraper.pid | PID file location |
| `--log-file` | string | ./quakewatch-scraper.log | Log file location |

## Monitoring and Logging

### Log Output

Interval scraping provides detailed logging:

```
[2024-01-15 10:00:00] Starting interval scraper
[2024-01-15 10:00:00] Configuration: interval=5m, max-runtime=24h, backoff=exponential
[2024-01-15 10:00:00] Executing command: earthquakes recent
[2024-01-15 10:00:00] Found 15 earthquakes
[2024-01-15 10:00:00] Successfully collected and saved 15 earthquakes
[2024-01-15 10:00:00] Next execution in 5 minutes
[2024-01-15 10:05:00] Executing command: earthquakes recent
[2024-01-15 10:05:00] Found 3 earthquakes
[2024-01-15 10:05:00] Successfully collected and saved 3 earthquakes
```

### Health Checks

Health checks monitor system status:

```bash
# Check interval scraper health
./bin/quakewatch-scraper health

# Check specific components
./bin/quakewatch-scraper health --storage postgresql
```

### Metrics

The interval scraper collects various metrics:

- Execution count
- Success/failure rates
- Execution time
- Data collection statistics
- Error counts and types

## Error Handling

### Automatic Retries

When a command fails, the interval scraper can retry with backoff:

```bash
# Exponential backoff (default)
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --backoff exponential \
  --max-backoff 30m

# Linear backoff
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --backoff linear \
  --max-backoff 10m
```

### Error Recovery

The scraper can continue running even when commands fail:

```bash
# Continue on errors
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --continue-on-error
```

## Daemon Mode

### Running as Daemon

Daemon mode runs the interval scraper in the background:

```bash
# Start daemon
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --daemon \
  --pid-file /var/run/quakewatch-scraper.pid \
  --log-file /var/log/quakewatch-scraper.log

# Check if daemon is running
ps aux | grep quakewatch-scraper

# Stop daemon
kill $(cat /var/run/quakewatch-scraper.pid)
```

### PID File Management

The daemon creates a PID file for process management:

```bash
# Custom PID file location
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --daemon \
  --pid-file /tmp/quakewatch.pid
```

## Best Practices

### 1. Choose Appropriate Intervals

- **Recent earthquakes**: 5-15 minutes
- **Significant earthquakes**: 1-6 hours
- **Fault data**: 12-24 hours
- **Historical data**: Daily or weekly

### 2. Use Smart Collection

Combine interval scraping with smart collection to avoid duplicates:

```bash
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --smart \
  --storage postgresql
```

### 3. Monitor Resource Usage

- Use appropriate limits for `--max-runtime` and `--max-executions`
- Monitor disk space and database size
- Check log files regularly

### 4. Error Handling

- Use `--continue-on-error` for production environments
- Configure appropriate backoff strategies
- Monitor error logs

### 5. Health Monitoring

- Enable health checks with `--health-check-interval`
- Monitor system resources
- Set up alerts for failures

## Troubleshooting

### Common Issues

1. **High CPU Usage**
   - Increase interval time
   - Reduce data collection limits
   - Check for infinite loops

2. **Memory Issues**
   - Reduce batch sizes
   - Increase garbage collection frequency
   - Monitor memory usage

3. **Database Connection Issues**
   - Check connection pool settings
   - Verify database availability
   - Monitor connection limits

4. **File System Issues**
   - Check disk space
   - Verify file permissions
   - Monitor I/O performance

### Debug Mode

Enable debug logging for troubleshooting:

```bash
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --verbose \
  --log-level debug
```

### Log Analysis

Analyze logs for patterns and issues:

```bash
# Check for errors
grep ERROR /var/log/quakewatch-scraper.log

# Check execution times
grep "execution time" /var/log/quakewatch-scraper.log

# Check data collection statistics
grep "Found.*earthquakes" /var/log/quakewatch-scraper.log
```

## Examples

### Production Setup

```bash
# Production interval scraper with comprehensive monitoring
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 5m \
  --max-runtime 168h \
  --backoff exponential \
  --max-backoff 30m \
  --continue-on-error \
  --health-check-interval 10m \
  --daemon \
  --pid-file /var/run/quakewatch-scraper.pid \
  --log-file /var/log/quakewatch-scraper.log \
  --smart \
  --storage postgresql
```

### Development Setup

```bash
# Development interval scraper with quick feedback
./bin/quakewatch-scraper interval earthquakes recent \
  --interval 30s \
  --max-executions 10 \
  --verbose \
  --log-level debug
```

### Custom Workflow

```bash
# Custom workflow with multiple commands
./bin/quakewatch-scraper interval custom \
  --interval 1h \
  --commands "earthquakes recent,earthquakes significant --start 2024-01-01 --end 2024-01-31,faults collect" \
  --continue-on-error \
  --skip-empty
```

## Integration

### Systemd Service

Create a systemd service for automatic startup:

```ini
[Unit]
Description=QuakeWatch Scraper Interval
After=network.target

[Service]
Type=forking
User=quakewatch
ExecStart=/path/to/quakewatch-scraper interval earthquakes recent --interval 5m --daemon
PIDFile=/var/run/quakewatch-scraper.pid
Restart=always

[Install]
WantedBy=multi-user.target
```

### Cron Alternative

For simple scheduling, you can use cron instead:

```bash
# Add to crontab
*/5 * * * * /path/to/quakewatch-scraper earthquakes recent --smart --storage postgresql
0 * * * * /path/to/quakewatch-scraper earthquakes significant --start $(date -d '1 hour ago' +%Y-%m-%d) --end $(date +%Y-%m-%d)
0 0 * * * /path/to/quakewatch-scraper faults collect
```

The interval scraping feature provides a robust, production-ready solution for continuous data collection with comprehensive monitoring and error handling capabilities. 
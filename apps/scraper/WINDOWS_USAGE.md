# Windows Usage Guide for QuakeWatch Scraper

## Overview

The QuakeWatch Scraper is a command-line tool for collecting earthquake and fault data. This guide explains how to use it on Windows systems.

## Quick Start

### 1. Download and Extract

1. Download the Windows executable (`quakewatch-scraper-windows-amd64.exe`)
2. Extract it to a folder of your choice (e.g., `C:\quakewatch-scraper\`)

### 2. Basic Usage

Open Command Prompt or PowerShell in the folder containing the executable and run:

```cmd
# Get recent earthquakes for Japan (last 30 days)
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan

# Get earthquakes for Japan with magnitude 4.0+ (last 30 days)
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan --min-mag 4.0

# Get earthquakes for Japan with custom date range
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan --start 2024-01-01 --end 2024-12-31 --min-mag 4.0
```

## Output Files

By default, the tool saves data to:
- **Windows**: `%APPDATA%\QuakeWatch\data\` (e.g., `C:\Users\YourUsername\AppData\Roaming\QuakeWatch\data\`)
- **Fallback**: `./data/` (relative to the executable)

### Custom Output Directory

You can specify a custom output directory:

```cmd
# Save to a specific folder
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan --output-dir "C:\MyData\earthquakes"

# Save to current directory
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan --output-dir ".\"
```

## Configuration

The tool automatically creates configuration files in:
- **Windows**: `%APPDATA%\QuakeWatch\configs\`
- **Fallback**: `./configs/` (relative to the executable)

### First Run

On first run, if no configuration is found, the tool will:
1. Use default settings
2. Create necessary directories
3. Save data to the appropriate location

## Command Examples

### Earthquake Commands

```cmd
# Recent earthquakes (last hour)
quakewatch-scraper-windows-amd64.exe earthquakes recent

# Time range earthquakes
quakewatch-scraper-windows-amd64.exe earthquakes time-range --start 2024-01-01 --end 2024-01-31

# Magnitude range earthquakes
quakewatch-scraper-windows-amd64.exe earthquakes magnitude --min 4.0 --max 8.0

# Significant earthquakes (M4.5+)
quakewatch-scraper-windows-amd64.exe earthquakes significant --start 2024-01-01 --end 2024-01-31

# Regional earthquakes
quakewatch-scraper-windows-amd64.exe earthquakes region --min-lat 32 --max-lat 42 --min-lon -125 --max-lon -114

# Country-specific earthquakes (with defaults)
quakewatch-scraper-windows-amd64.exe earthquakes country --country "United States"
```

### Output Options

```cmd
# Output to stdout (no file saved)
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan --stdout

# Custom filename
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan --filename "japan_earthquakes"

# Limit number of records
quakewatch-scraper-windows-amd64.exe earthquakes country --country Japan --limit 100
```

## Troubleshooting

### "required flag(s) 'end', 'start' not set"

This error occurs when using commands that require date ranges. For the `country` command, you can:

1. **Use defaults** (recommended): Don't specify `--start` and `--end` - it will use the last 30 days
2. **Specify dates**: Use `--start` and `--end` flags with dates in YYYY-MM-DD format

### File Not Found

If you can't find your output files:

1. Check the console output - it shows the exact path where files are saved
2. Look in `%APPDATA%\QuakeWatch\data\` (Windows default)
3. Use `--output-dir` to specify a custom location

### Permission Errors

If you get permission errors:

1. Run Command Prompt as Administrator
2. Choose a different output directory with `--output-dir`
3. Ensure you have write permissions to the target directory

## Data File Format

Earthquake data is saved as JSON files with timestamps:
- `earthquakes_2024-01-15_14-30-25.json`
- `earthquakes_2024-01-15_14-35-10.json`

Each file contains GeoJSON format earthquake data from the USGS API.

## Advanced Usage

### Interval Collection

Run commands at regular intervals:

```cmd
# Collect Japan earthquakes every hour
quakewatch-scraper-windows-amd64.exe interval earthquakes country --interval 1h --country Japan

# Collect significant earthquakes every 6 hours
quakewatch-scraper-windows-amd64.exe interval earthquakes significant --interval 6h --start 2024-01-01 --end 2024-12-31
```

### Database Storage

For PostgreSQL storage (requires database setup):

```cmd
quakewatch-scraper-windows-amd64.exe earthquakes recent --storage postgresql
```

## Support

For issues or questions:
1. Check the console output for error messages
2. Verify your internet connection (required for API access)
3. Ensure you have sufficient disk space for data files 
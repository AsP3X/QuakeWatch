# QuakeWatch

A comprehensive system for gathering, storing, and displaying earthquake and fault data from multiple seismological sources.

## Overview

QuakeWatch is a monorepo containing applications for real-time earthquake data collection and web-based visualization. The system fetches data from USGS and EMSC-CSEM APIs, processes and stores it, and provides an interactive web interface for viewing earthquake information.

## Project Structure

```
QuakeWatch/
├── apps/
│   ├── quakewatch/          # Web application (Next.js)
│   └── scraper/            # Data collection service (Go)
└── LICENSE
```

## Applications

### Data Scraper (`apps/scraper/`)
A Go application that collects earthquake and fault data from:
- **USGS Earthquake API**: Real-time earthquake data
- **EMSC-CSEM API**: Fault data

Features:
- Command-line interface for data collection
- Country and region filtering
- JSON file storage
- Data validation and statistics

### Web Application (`apps/quakewatch/`)
A Next.js web application for displaying earthquake data with:
- Interactive maps
- Real-time data visualization
- Search and filtering capabilities
- User authentication

## Getting Started

### Prerequisites
- Go 1.21+ (for scraper)
- Node.js 18+ (for web app)
- PostgreSQL with PostGIS (for database)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd QuakeWatch
   ```

2. **Set up the data scraper**
   ```bash
   cd apps/scraper
   make install
   make build
   ```

3. **Run the scraper**
   ```bash
   ./bin/quakewatch-scraper earthquakes recent --stdout --limit 5
   ```

## Documentation

- [Scraper Documentation](apps/scraper/README.md)
- [Project Planning](apps/quakewatch/planning/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 
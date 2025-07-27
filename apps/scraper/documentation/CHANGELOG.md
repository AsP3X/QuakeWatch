# Changelog

All notable changes to the QuakeWatch Scraper project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation suite
- Database schema documentation
- Deployment guide with multiple deployment strategies
- API reference with complete command documentation
- Security considerations and best practices
- Monitoring and logging configuration examples

### Changed
- Enhanced project structure documentation
- Improved configuration examples
- Updated build and deployment instructions

## [1.0.0] - 2024-01-15

### Added
- Initial release of QuakeWatch Scraper
- Multi-source data collection from USGS FDSNWS API and EMSC-CSEM
- Comprehensive CLI interface with earthquake and fault data collection
- PostgreSQL database support with full CRUD operations
- JSON file storage for data persistence
- Scheduled collection with configurable intervals
- Database migration system using golang-migrate
- Cross-platform support (Linux, macOS, Windows)
- Health monitoring and status checking
- Data validation and cleaning capabilities
- Statistics and reporting functionality
- Data management utilities (list, purge, validate)

### Features
- **Earthquake Data Collection**:
  - Recent earthquakes (last hour)
  - Time range collection
  - Magnitude-based filtering
  - Significant earthquakes (M4.5+)
  - Geographic region filtering
  - Country-specific collection

- **Fault Data Collection**:
  - EMSC-CSEM fault data collection
  - Fault data updates
  - Geographic fault information

- **Scheduled Collection**:
  - Interval-based collection for earthquakes
  - Interval-based collection for faults
  - Custom interval scheduling
  - Duration-based execution

- **Database Operations**:
  - Database initialization
  - Migration management (up, down, status)
  - Force version control
  - Connection health checks

- **Utility Commands**:
  - Version information
  - Health checks
  - Statistics reporting
  - Data validation
  - File listing and management
  - Data purging with dry-run support

### Technical Details
- **Architecture**: Modular Go application with clean separation of concerns
- **Storage**: Dual storage support (JSON files and PostgreSQL)
- **API Integration**: USGS FDSNWS and EMSC-CSEM APIs
- **Configuration**: YAML-based configuration with environment variable support
- **Logging**: Structured logging with configurable levels and formats
- **Error Handling**: Comprehensive error handling with retry logic
- **Performance**: Optimized database queries with strategic indexing

### Database Schema
- **Earthquakes Table**: Complete earthquake event data with metadata
- **Faults Table**: Geological fault data with spatial information
- **Collection Logs Table**: Audit trail for data collection activities
- **Collection Metadata Table**: Tracking of last collection times

### Dependencies
- Go 1.24+
- PostgreSQL 12+ (optional)
- golang-migrate/v4
- sqlx
- lib/pq
- logrus
- cobra
- viper

## [0.9.0] - 2024-01-10

### Added
- Beta release with core functionality
- Basic earthquake data collection
- Simple fault data collection
- JSON file storage
- Basic CLI interface

### Changed
- Improved error handling
- Enhanced logging
- Better configuration management

## [0.8.0] - 2024-01-05

### Added
- Alpha release with initial features
- USGS API integration
- Basic data models
- Simple file storage

### Known Issues
- Limited error handling
- Basic logging only
- No database support
- Minimal configuration options

## [0.7.0] - 2024-01-01

### Added
- Initial project setup
- Basic project structure
- Go module configuration
- Makefile for build automation

### Technical Foundation
- Project scaffolding
- Dependency management
- Build system setup
- Development environment configuration

---

## Version History Summary

| Version | Release Date | Status | Key Features |
|---------|--------------|--------|--------------|
| 1.0.0 | 2024-01-15 | Stable | Full feature set, production ready |
| 0.9.0 | 2024-01-10 | Beta | Core functionality, testing phase |
| 0.8.0 | 2024-01-05 | Alpha | Basic features, development phase |
| 0.7.0 | 2024-01-01 | Pre-alpha | Project setup, foundation |

## Migration Guide

### Upgrading from 0.9.0 to 1.0.0

1. **Database Changes**:
   - Run database migrations: `./quakewatch-scraper db migrate up`
   - New tables added: collection_logs, collection_metadata
   - Enhanced indexes for performance

2. **Configuration Changes**:
   - New configuration options for logging and storage
   - Enhanced database configuration
   - Additional API configuration options

3. **CLI Changes**:
   - New commands: `interval`, `stats`, `validate`, `list`, `purge`
   - Enhanced existing commands with additional options
   - Improved error messages and help text

### Upgrading from 0.8.0 to 1.0.0

1. **Major Changes**:
   - Complete rewrite of data collection logic
   - New database schema
   - Enhanced CLI interface
   - Improved error handling

2. **Breaking Changes**:
   - Configuration file format changed
   - Data storage format updated
   - Command-line interface restructured

3. **Migration Steps**:
   - Backup existing data
   - Update configuration files
   - Run database migrations
   - Test new functionality

## Release Process

### Version Numbering
- **Major Version**: Breaking changes, new major features
- **Minor Version**: New features, backward compatible
- **Patch Version**: Bug fixes, minor improvements

### Release Checklist
- [ ] Update version in code
- [ ] Update documentation
- [ ] Run full test suite
- [ ] Build for all platforms
- [ ] Create release notes
- [ ] Tag release in Git
- [ ] Update changelog
- [ ] Deploy to package repositories

### Supported Platforms

#### Version 1.0.0
- **Linux**: Ubuntu 20.04+, CentOS 8+, RHEL 8+
- **macOS**: 10.15+ (Catalina and later)
- **Windows**: Windows 10+ (64-bit)

#### Build Targets
- Linux AMD64
- Linux ARM64
- macOS AMD64
- Windows AMD64
- Windows ARM64

## Contributing

### Development Workflow
1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Update documentation
5. Submit a pull request

### Code Standards
- Follow Go coding standards
- Add tests for new functionality
- Update documentation for API changes
- Use conventional commit messages

### Testing
- Unit tests for all packages
- Integration tests for API clients
- Database migration tests
- End-to-end CLI tests

## Support

### Version Support Policy
- **Current Version**: Full support, security updates, bug fixes
- **Previous Version**: Security updates and critical bug fixes only
- **Older Versions**: No official support

### Getting Help
- Check the documentation
- Review the troubleshooting guide
- Create an issue on GitHub
- Contact the development team

---

*This changelog is maintained by the QuakeWatch Scraper development team. For more information, see the [project documentation](README.md).* 
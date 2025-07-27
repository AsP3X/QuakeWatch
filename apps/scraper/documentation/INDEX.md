# Documentation Index

Welcome to the QuakeWatch Scraper documentation. This index provides an overview of all available documentation and guides you to the right resources for your needs.

## 📚 Documentation Overview

The QuakeWatch Scraper documentation is organized into several comprehensive guides, each focusing on different aspects of the tool:

### Core Documentation

| Document | Purpose | Audience |
|----------|---------|----------|
| [README.md](README.md) | Main project overview and quick start | All users |
| [API_REFERENCE.md](API_REFERENCE.md) | Complete command reference and options | Users, Developers |
| [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) | Database structure and management | Developers, DBAs |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Installation and deployment guides | System Administrators |
| [DEVELOPMENT.md](DEVELOPMENT.md) | Development setup and contribution guide | Developers |
| [CHANGELOG.md](CHANGELOG.md) | Version history and changes | All users |

## 🎯 Quick Navigation by Use Case

### I'm a New User
Start here to get up and running quickly:
1. [README.md](README.md) - Overview and quick start
2. [API_REFERENCE.md](API_REFERENCE.md) - Learn the commands
3. [DEPLOYMENT.md](DEPLOYMENT.md) - Installation instructions

### I'm a System Administrator
For deployment and production management:
1. [DEPLOYMENT.md](DEPLOYMENT.md) - Production deployment guide
2. [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) - Database setup and management
3. [API_REFERENCE.md](API_REFERENCE.md) - Command reference for operations

### I'm a Developer
For contributing to the project:
1. [DEVELOPMENT.md](DEVELOPMENT.md) - Development environment setup
2. [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) - Database development
3. [API_REFERENCE.md](API_REFERENCE.md) - API and CLI development

### I'm a Data Scientist/Researcher
For using the collected data:
1. [README.md](README.md) - Understanding the tool
2. [API_REFERENCE.md](API_REFERENCE.md) - Data collection commands
3. [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) - Data structure and queries

## 📖 Detailed Documentation Guide

### [README.md](README.md) - Main Documentation
**What you'll find:**
- Project overview and features
- Installation instructions
- Quick start guide
- Basic usage examples
- Architecture overview
- Data sources information

**Best for:** Getting started, understanding the project, basic usage

### [API_REFERENCE.md](API_REFERENCE.md) - Complete API Reference
**What you'll find:**
- All CLI commands and options
- Configuration reference
- Data models and structures
- Error codes and exit codes
- Examples for every command

**Best for:** Daily usage, command reference, troubleshooting

### [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) - Database Documentation
**What you'll find:**
- Complete database schema
- Table structures and relationships
- Indexes and performance optimization
- Migration system
- Backup and recovery procedures
- Security considerations

**Best for:** Database administration, performance tuning, data analysis

### [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment Guide
**What you'll find:**
- Development environment setup
- Production deployment strategies
- Docker deployment
- Systemd service configuration
- Monitoring and logging
- Security hardening
- Troubleshooting guide

**Best for:** System administrators, DevOps engineers, production deployment

### [DEVELOPMENT.md](DEVELOPMENT.md) - Development Guide
**What you'll find:**
- Development environment setup
- Coding standards and conventions
- Testing strategies and examples
- Database development
- API development
- CLI development
- Contributing guidelines
- Release process

**Best for:** Developers, contributors, code maintainers

### [CHANGELOG.md](CHANGELOG.md) - Version History
**What you'll find:**
- Complete version history
- Feature additions and changes
- Bug fixes and improvements
- Migration guides between versions
- Release process information
- Support policy

**Best for:** Understanding changes, upgrading, release planning

## 🔍 Finding Specific Information

### Common Tasks Quick Reference

| Task | Primary Document | Secondary Document |
|------|------------------|-------------------|
| Install the tool | [README.md](README.md) | [DEPLOYMENT.md](DEPLOYMENT.md) |
| Learn commands | [API_REFERENCE.md](API_REFERENCE.md) | [README.md](README.md) |
| Set up database | [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) | [DEPLOYMENT.md](DEPLOYMENT.md) |
| Deploy to production | [DEPLOYMENT.md](DEPLOYMENT.md) | [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) |
| Contribute code | [DEVELOPMENT.md](DEVELOPMENT.md) | [API_REFERENCE.md](API_REFERENCE.md) |
| Understand data structure | [DATABASE_SCHEMA.md](DATABASE_SCHEMA.md) | [API_REFERENCE.md](API_REFERENCE.md) |
| Troubleshoot issues | [DEPLOYMENT.md](DEPLOYMENT.md) | [API_REFERENCE.md](API_REFERENCE.md) |
| Upgrade version | [CHANGELOG.md](CHANGELOG.md) | [DEPLOYMENT.md](DEPLOYMENT.md) |

### Command Reference Quick Links

#### Earthquake Data Collection
- [Recent earthquakes](API_REFERENCE.md#earthquakes-recent)
- [Time range collection](API_REFERENCE.md#earthquakes-time-range)
- [Magnitude filtering](API_REFERENCE.md#earthquakes-magnitude)
- [Significant earthquakes](API_REFERENCE.md#earthquakes-significant)
- [Geographic filtering](API_REFERENCE.md#earthquakes-region)
- [Country-specific collection](API_REFERENCE.md#earthquakes-country)

#### Fault Data Collection
- [Collect fault data](API_REFERENCE.md#faults-collect)
- [Update fault data](API_REFERENCE.md#faults-update)

#### Scheduled Collection
- [Interval-based collection](API_REFERENCE.md#interval-commands)
- [Earthquake scheduling](API_REFERENCE.md#interval-earthquakes-recent)
- [Fault scheduling](API_REFERENCE.md#interval-faults-collect)

#### Database Operations
- [Database initialization](API_REFERENCE.md#db-init)
- [Migration management](API_REFERENCE.md#db-migrate-up)
- [Database status](API_REFERENCE.md#db-status)

#### Utility Commands
- [Health checks](API_REFERENCE.md#health)
- [Statistics](API_REFERENCE.md#stats)
- [Data validation](API_REFERENCE.md#validate)
- [Data management](API_REFERENCE.md#list)
- [Data purging](API_REFERENCE.md#purge)

## 🛠️ Getting Help

### Documentation Issues
If you find issues with the documentation:
1. Check the [CHANGELOG.md](CHANGELOG.md) for recent updates
2. Search for similar issues in the project repository
3. Create an issue with the `documentation` label

### Tool Issues
If you encounter problems with the tool:
1. Check the [troubleshooting section](DEPLOYMENT.md#troubleshooting)
2. Review the [error codes](API_REFERENCE.md#error-codes)
3. Check the [health command](API_REFERENCE.md#health)
4. Create an issue with detailed information

### Feature Requests
For new features or improvements:
1. Check the [CHANGELOG.md](CHANGELOG.md) for planned features
2. Review existing issues to avoid duplicates
3. Create a feature request with detailed use case

## 📝 Contributing to Documentation

### Documentation Standards
- Use clear, concise language
- Include practical examples
- Maintain consistent formatting
- Update related documents when making changes
- Test all code examples

### Documentation Structure
- Each document should have a clear purpose
- Include table of contents for long documents
- Use consistent heading levels
- Cross-reference related information
- Include version information where relevant

### Updating Documentation
1. Follow the [development workflow](DEVELOPMENT.md#development-workflow)
2. Update relevant documentation when adding features
3. Test all examples and commands
4. Update the changelog for significant changes
5. Review documentation in pull requests

## 🔗 External Resources

### Official Documentation
- [Go Programming Language](https://golang.org/doc/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Documentation](https://docs.docker.com/)

### API References
- [USGS FDSNWS API](https://earthquake.usgs.gov/fdsnws/event/1/)
- [EMSC-CSEM API](https://www.emsc-csem.org/)

### Development Tools
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Viper Configuration](https://github.com/spf13/viper)
- [Golang Migrate](https://github.com/golang-migrate/migrate)

## 📊 Documentation Metrics

This documentation suite includes:
- **6 comprehensive guides** covering all aspects of the tool
- **100+ code examples** demonstrating real usage
- **Complete API reference** with all commands and options
- **Production-ready deployment** instructions
- **Developer-friendly** contribution guidelines

## 🎉 Getting Started

Ready to dive in? Here's a recommended reading order:

1. **Start with [README.md](README.md)** - Get the big picture
2. **Try the [quick start guide](README.md#quick-start)** - Get hands-on experience
3. **Explore [API_REFERENCE.md](API_REFERENCE.md)** - Learn the commands
4. **Set up your environment** - Choose from [DEPLOYMENT.md](DEPLOYMENT.md) or [DEVELOPMENT.md](DEVELOPMENT.md)
5. **Dive deeper** - Explore specific areas based on your needs

Happy exploring! 🚀

---

*This documentation index is maintained by the QuakeWatch Scraper development team. For questions or suggestions, please create an issue in the project repository.* 
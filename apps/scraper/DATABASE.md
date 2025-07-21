# PostgreSQL Database Implementation

This document outlines the PostgreSQL database implementation for the QuakeWatch Data Scraper.

## Overview

The QuakeWatch Scraper now supports PostgreSQL as a storage backend alongside the existing JSON file storage. This provides:

- **Structured Data Storage**: Relational database with proper indexing
- **Advanced Querying**: SQL-based queries with geographic and temporal filtering
- **Data Integrity**: ACID compliance and constraint enforcement
- **Scalability**: Better performance for large datasets
- **Concurrent Access**: Multiple application instances can access the same data

## Architecture

### Storage Interface

The application uses a common `Storage` interface that both JSON and PostgreSQL implementations satisfy:

```go
type Storage interface {
    // Earthquake operations
    SaveEarthquakes(ctx context.Context, earthquakes *models.USGSResponse) error
    LoadEarthquakes(ctx context.Context, limit int, offset int) (*models.USGSResponse, error
    // ... additional methods
}
```

### Database Schema

#### Earthquakes Table
```sql
CREATE TABLE earthquakes (
    id SERIAL PRIMARY KEY,
    usgs_id VARCHAR(255) UNIQUE NOT NULL,
    magnitude DECIMAL(4,2) NOT NULL,
    magnitude_type VARCHAR(10),
    place TEXT NOT NULL,
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    updated TIMESTAMP WITH TIME ZONE NOT NULL,
    -- ... additional fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Faults Table
```sql
CREATE TABLE faults (
    id SERIAL PRIMARY KEY,
    fault_id VARCHAR(255) UNIQUE NOT NULL,
    name TEXT NOT NULL,
    fault_type VARCHAR(100),
    -- ... additional fields
    coordinates JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Collection Logs Table
```sql
CREATE TABLE collection_logs (
    id SERIAL PRIMARY KEY,
    data_type VARCHAR(50) NOT NULL,
    source VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    records_collected INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Setup Instructions

### 1. Install PostgreSQL

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

#### macOS (using Homebrew)
```bash
brew install postgresql
brew services start postgresql
```

#### Windows
Download and install from [PostgreSQL official website](https://www.postgresql.org/download/windows/)

### 2. Create Database and User

```bash
# Connect to PostgreSQL as superuser
sudo -u postgres psql

# Create database and user
CREATE DATABASE quakewatch;
CREATE USER quakewatch_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE quakewatch TO quakewatch_user;
\q
```

### 3. Configure Environment Variables

Create a `.env` file in the project root:

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

### 4. Run Database Migrations

```bash
# Build the application
make build

# Run migrations
make db-migrate-up

# Check migration status
make db-status
```

### 5. Using Docker (Alternative)

If you prefer using Docker for PostgreSQL:

```bash
# Start PostgreSQL container
make db-setup-docker

# Run migrations
make db-migrate-up

# Stop container when done
make db-stop-docker
```

## Usage

### Basic Database Operations

```bash
# Initialize database (creates tables and indexes)
./bin/quakewatch-scraper db init

# Check database status
./bin/quakewatch-scraper db status

# Run migrations
./bin/quakewatch-scraper db migrate up

# Rollback migrations
./bin/quakewatch-scraper db migrate down

# Migrate to specific version
./bin/quakewatch-scraper db migrate to 1
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

### Advanced Queries

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

## Performance Considerations

### Indexing Strategy

The database includes several indexes for optimal performance:

- **Time-based queries**: `idx_earthquakes_time`
- **Magnitude filtering**: `idx_earthquakes_magnitude`
- **Geographic queries**: `idx_earthquakes_location`
- **Significance filtering**: `idx_earthquakes_significance`
- **Unique constraints**: `idx_earthquakes_usgs_id`

### Connection Pooling

The application uses connection pooling with configurable settings:

- **Max Open Connections**: 25 (default)
- **Max Idle Connections**: 5 (default)
- **Connection Lifetime**: 5 minutes (default)
- **Idle Timeout**: 5 minutes (default)

### Batch Operations

For bulk data insertion, the application uses:

- **Transactions**: All operations within a single transaction
- **Upsert Logic**: `ON CONFLICT` clauses for duplicate handling
- **Batch Processing**: Multiple records processed in single queries

## Monitoring and Maintenance

### Database Statistics

```bash
# Get database statistics
./bin/quakewatch-scraper db stats

# Monitor collection logs
./bin/quakewatch-scraper db logs --limit 10
```

### Backup and Restore

```bash
# Create database backup
pg_dump -h localhost -U quakewatch_user -d quakewatch > backup.sql

# Restore from backup
psql -h localhost -U quakewatch_user -d quakewatch < backup.sql
```

### Data Migration from JSON

```bash
# Migrate existing JSON data to PostgreSQL
./bin/quakewatch-scraper migrate json-to-postgresql \
  --json-dir ./data \
  --storage postgresql
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure PostgreSQL is running
   - Check host and port configuration
   - Verify firewall settings

2. **Authentication Failed**
   - Verify username and password
   - Check pg_hba.conf configuration
   - Ensure user has proper permissions

3. **Migration Errors**
   - Check migration file syntax
   - Verify database user permissions
   - Review migration logs

### Debug Mode

Enable debug logging for database operations:

```bash
export LOG_LEVEL=debug
./bin/quakewatch-scraper earthquakes recent --storage postgresql
```

## Security Considerations

1. **Connection Security**
   - Use SSL connections in production
   - Implement connection encryption
   - Use strong passwords

2. **Access Control**
   - Limit database user permissions
   - Use read-only users for queries
   - Implement connection pooling limits

3. **Data Protection**
   - Regular backups
   - Encryption at rest
   - Audit logging

## Future Enhancements

1. **PostGIS Integration**
   - Geographic indexing
   - Spatial queries
   - Distance calculations

2. **Partitioning**
   - Time-based partitioning
   - Geographic partitioning
   - Improved query performance

3. **Replication**
   - Read replicas
   - High availability
   - Load balancing

4. **Advanced Analytics**
   - Materialized views
   - Statistical functions
   - Trend analysis

## API Reference

### Database Configuration

```go
type DatabaseConfig struct {
    Host            string
    Port            int
    User            string
    Password        string
    Database        string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}
```

### Storage Interface

```go
type Storage interface {
    // Earthquake operations
    SaveEarthquakes(ctx context.Context, earthquakes *models.USGSResponse) error
    LoadEarthquakes(ctx context.Context, limit int, offset int) (*models.USGSResponse, error)
    GetEarthquakeByID(ctx context.Context, usgsID string) (*models.Earthquake, error)
    // ... additional methods
}
```

### Migration Manager

```go
type MigrationManager struct {
    db     *sqlx.DB
    config *config.DatabaseConfig
}

func (m *MigrationManager) MigrateUp() error
func (m *MigrationManager) MigrateDown() error
func (m *MigrationManager) GetVersion() (uint, bool, error)
```

## Contributing

When contributing to the database implementation:

1. **Schema Changes**: Always create migration files
2. **Testing**: Include database tests
3. **Documentation**: Update this document
4. **Performance**: Consider query optimization
5. **Security**: Follow security best practices

## Support

For database-related issues:

1. Check the troubleshooting section
2. Review PostgreSQL logs
3. Enable debug logging
4. Create detailed issue reports
5. Include relevant configuration and error messages 
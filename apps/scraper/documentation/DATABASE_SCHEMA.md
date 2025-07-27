# Database Schema Documentation

This document provides detailed information about the database schema, tables, relationships, and migration system used by the QuakeWatch Scraper.

## Table of Contents

- [Overview](#overview)
- [Database Tables](#database-tables)
- [Indexes](#indexes)
- [Triggers](#triggers)
- [Data Types](#data-types)
- [Migrations](#migrations)
- [Performance Considerations](#performance-considerations)
- [Backup and Recovery](#backup-and-recovery)

## Overview

The QuakeWatch Scraper uses PostgreSQL as its primary database system. The schema is designed to efficiently store and query earthquake and fault data with support for:

- **Earthquake Events**: Complete earthquake data with metadata
- **Fault Information**: Geological fault data with spatial information
- **Collection Logs**: Audit trail of data collection activities
- **Collection Metadata**: Tracking of last collection times

### Database Features

- **Spatial Data Support**: Uses JSONB for storing geographical coordinates
- **Audit Trail**: Automatic tracking of creation and update times
- **Performance Optimization**: Strategic indexing for common queries
- **Data Integrity**: Foreign key constraints and data validation
- **Migration System**: Version-controlled schema changes

## Database Tables

### 1. Earthquakes Table

The main table for storing earthquake event data.

```sql
CREATE TABLE earthquakes (
    id SERIAL PRIMARY KEY,
    usgs_id VARCHAR(255) UNIQUE NOT NULL,
    magnitude DECIMAL(4,2) NOT NULL,
    magnitude_type VARCHAR(10),
    place TEXT NOT NULL,
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    updated TIMESTAMP WITH TIME ZONE NOT NULL,
    url TEXT,
    detail_url TEXT,
    felt_count INTEGER,
    cdi DECIMAL(3,1),
    mmi DECIMAL(3,1),
    alert VARCHAR(50),
    status VARCHAR(50) NOT NULL,
    tsunami BOOLEAN DEFAULT FALSE,
    sig INTEGER NOT NULL,
    net VARCHAR(10) NOT NULL,
    code VARCHAR(50) NOT NULL,
    ids TEXT,
    sources TEXT,
    types TEXT,
    nst INTEGER,
    dmin DECIMAL(10,6),
    rms DECIMAL(10,6),
    gap DECIMAL(5,2),
    latitude DECIMAL(10,6) NOT NULL,
    longitude DECIMAL(10,6) NOT NULL,
    depth DECIMAL(10,6),
    title TEXT,
    felt INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Column Descriptions

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key, auto-incrementing |
| `usgs_id` | VARCHAR(255) | Unique USGS identifier |
| `magnitude` | DECIMAL(4,2) | Earthquake magnitude |
| `magnitude_type` | VARCHAR(10) | Type of magnitude measurement |
| `place` | TEXT | Human-readable location description |
| `time` | TIMESTAMP WITH TIME ZONE | Earthquake occurrence time |
| `updated` | TIMESTAMP WITH TIME ZONE | Last update time from USGS |
| `url` | TEXT | USGS event page URL |
| `detail_url` | TEXT | USGS detail API URL |
| `felt_count` | INTEGER | Number of people who felt the earthquake |
| `cdi` | DECIMAL(3,1) | Community Intensity (CDI) |
| `mmi` | DECIMAL(3,1) | Modified Mercalli Intensity |
| `alert` | VARCHAR(50) | Alert level (green, yellow, orange, red) |
| `status` | VARCHAR(50) | Review status |
| `tsunami` | BOOLEAN | Tsunami warning flag |
| `sig` | INTEGER | Significance value |
| `net` | VARCHAR(10) | Network identifier |
| `code` | VARCHAR(50) | Event code |
| `ids` | TEXT | Comma-separated list of IDs |
| `sources` | TEXT | Comma-separated list of sources |
| `types` | TEXT | Comma-separated list of data types |
| `nst` | INTEGER | Number of seismic stations |
| `dmin` | DECIMAL(10,6) | Distance to nearest station |
| `rms` | DECIMAL(10,6) | Root mean square of travel time residuals |
| `gap` | DECIMAL(5,2) | Largest azimuthal gap between stations |
| `latitude` | DECIMAL(10,6) | Earthquake latitude |
| `longitude` | DECIMAL(10,6) | Earthquake longitude |
| `depth` | DECIMAL(10,6) | Earthquake depth in kilometers |
| `title` | TEXT | Human-readable title |
| `felt` | INTEGER | Number of felt reports |
| `created_at` | TIMESTAMP WITH TIME ZONE | Record creation time |
| `updated_at` | TIMESTAMP WITH TIME ZONE | Record last update time |

### 2. Faults Table

Table for storing geological fault data.

```sql
CREATE TABLE faults (
    id SERIAL PRIMARY KEY,
    fault_id VARCHAR(255) UNIQUE NOT NULL,
    name TEXT NOT NULL,
    fault_type VARCHAR(100),
    slip_rate DECIMAL(10,2),
    slip_type VARCHAR(50),
    dip DECIMAL(5,2),
    rake DECIMAL(5,2),
    length DECIMAL(10,2),
    width DECIMAL(10,2),
    max_magnitude DECIMAL(4,2),
    description TEXT,
    source VARCHAR(255),
    geometry_type VARCHAR(50) NOT NULL,
    coordinates JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Column Descriptions

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key, auto-incrementing |
| `fault_id` | VARCHAR(255) | Unique fault identifier |
| `name` | TEXT | Fault name |
| `fault_type` | VARCHAR(100) | Type of fault (strike-slip, normal, reverse) |
| `slip_rate` | DECIMAL(10,2) | Slip rate in mm/year |
| `slip_type` | VARCHAR(50) | Type of slip (left-lateral, right-lateral) |
| `dip` | DECIMAL(5,2) | Fault dip angle in degrees |
| `rake` | DECIMAL(5,2) | Fault rake angle in degrees |
| `length` | DECIMAL(10,2) | Fault length in kilometers |
| `width` | DECIMAL(10,2) | Fault width in kilometers |
| `max_magnitude` | DECIMAL(4,2) | Maximum expected magnitude |
| `description` | TEXT | Fault description |
| `source` | VARCHAR(255) | Data source |
| `geometry_type` | VARCHAR(50) | Geometry type (LineString, Polygon) |
| `coordinates` | JSONB | Geographical coordinates |
| `created_at` | TIMESTAMP WITH TIME ZONE | Record creation time |
| `updated_at` | TIMESTAMP WITH TIME ZONE | Record last update time |

### 3. Collection Logs Table

Audit trail for data collection activities.

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

#### Column Descriptions

| Column | Type | Description |
|--------|------|-------------|
| `id` | SERIAL | Primary key, auto-incrementing |
| `data_type` | VARCHAR(50) | Type of data collected (earthquakes, faults) |
| `source` | VARCHAR(100) | Data source (usgs, emsc) |
| `start_time` | TIMESTAMP WITH TIME ZONE | Collection start time |
| `end_time` | TIMESTAMP WITH TIME ZONE | Collection end time |
| `records_collected` | INTEGER | Number of records collected |
| `status` | VARCHAR(50) | Collection status (completed, failed, partial) |
| `error_message` | TEXT | Error message if collection failed |
| `created_at` | TIMESTAMP WITH TIME ZONE | Record creation time |

### 4. Collection Metadata Table

Tracks last collection times for different data types.

```sql
CREATE TABLE collection_metadata (
    data_type VARCHAR(50) PRIMARY KEY,
    last_collection_time BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Column Descriptions

| Column | Type | Description |
|--------|------|-------------|
| `data_type` | VARCHAR(50) | Primary key, data type identifier |
| `last_collection_time` | BIGINT | Unix timestamp of last collection |
| `created_at` | TIMESTAMP WITH TIME ZONE | Record creation time |
| `updated_at` | TIMESTAMP WITH TIME ZONE | Record last update time |

## Indexes

### Performance Indexes

The database includes strategic indexes to optimize common query patterns:

```sql
-- Earthquake indexes
CREATE INDEX idx_earthquakes_time ON earthquakes(time);
CREATE INDEX idx_earthquakes_magnitude ON earthquakes(magnitude);
CREATE INDEX idx_earthquakes_location ON earthquakes(latitude, longitude);
CREATE INDEX idx_earthquakes_sig ON earthquakes(sig);
CREATE INDEX idx_earthquakes_usgs_id ON earthquakes(usgs_id);

-- Fault indexes
CREATE INDEX idx_faults_name ON faults(name);
CREATE INDEX idx_faults_type ON faults(fault_type);
CREATE INDEX idx_faults_fault_id ON faults(fault_id);

-- Collection log indexes
CREATE INDEX idx_collection_logs_type_time ON collection_logs(data_type, start_time);
CREATE INDEX idx_collection_logs_status ON collection_logs(status);
CREATE INDEX idx_collection_logs_created_at ON collection_logs(created_at);

-- Collection metadata indexes
CREATE INDEX idx_collection_metadata_data_type ON collection_metadata(data_type);
```

### Index Usage Patterns

| Index | Purpose | Query Patterns |
|-------|---------|----------------|
| `idx_earthquakes_time` | Time-based queries | Recent earthquakes, date ranges |
| `idx_earthquakes_magnitude` | Magnitude filtering | Significant earthquakes, magnitude ranges |
| `idx_earthquakes_location` | Spatial queries | Geographic region searches |
| `idx_earthquakes_sig` | Significance filtering | High-significance events |
| `idx_earthquakes_usgs_id` | Unique lookups | Duplicate prevention |
| `idx_faults_name` | Fault name searches | Fault lookup by name |
| `idx_faults_type` | Fault type filtering | Filtering by fault type |
| `idx_collection_logs_type_time` | Collection history | Collection timeline queries |
| `idx_collection_logs_status` | Status filtering | Failed collection analysis |

## Triggers

### Automatic Timestamp Updates

The database uses triggers to automatically update the `updated_at` timestamp:

```sql
-- Trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers
CREATE TRIGGER update_earthquakes_updated_at 
    BEFORE UPDATE ON earthquakes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_faults_updated_at 
    BEFORE UPDATE ON faults
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_collection_metadata_updated_at 
    BEFORE UPDATE ON collection_metadata
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

## Data Types

### PostgreSQL Extensions

```sql
-- Enable UUID extension for future use
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

### Custom Data Types

The schema uses standard PostgreSQL data types optimized for the data:

- **DECIMAL(4,2)**: For magnitude values (e.g., 4.5, 8.2)
- **DECIMAL(10,6)**: For precise coordinates and depths
- **JSONB**: For flexible coordinate storage
- **TIMESTAMP WITH TIME ZONE**: For timezone-aware timestamps
- **TEXT**: For variable-length string data
- **VARCHAR**: For fixed-length string data with constraints

## Migrations

### Migration System

The QuakeWatch Scraper uses the `golang-migrate` library for database migrations. Migrations are stored in the `migrations/` directory.

### Current Migration

**File**: `000001_combined_schema.up.sql`

This migration creates the complete initial schema including:
- All tables (earthquakes, faults, collection_logs, collection_metadata)
- All indexes for performance optimization
- All triggers for automatic timestamp updates
- PostgreSQL extensions

### Migration Commands

```bash
# Apply all pending migrations
./bin/quakewatch-scraper db migrate up

# Apply specific number of migrations
./bin/quakewatch-scraper db migrate up --steps 2

# Rollback last migration
./bin/quakewatch-scraper db migrate down

# Rollback specific number of migrations
./bin/quakewatch-scraper db migrate down --steps 3

# Check migration status
./bin/quakewatch-scraper db status

# Force migration version
./bin/quakewatch-scraper db force-version 1
```

### Migration Best Practices

1. **Always backup before migrations**: Use `pg_dump` before applying migrations
2. **Test migrations**: Test migrations on a copy of production data
3. **Version control**: Keep migrations in version control
4. **Rollback planning**: Ensure down migrations are properly tested
5. **Performance impact**: Monitor performance during large migrations

## Performance Considerations

### Query Optimization

#### Common Query Patterns

1. **Recent Earthquakes**
```sql
SELECT * FROM earthquakes 
WHERE time >= NOW() - INTERVAL '1 hour' 
ORDER BY time DESC;
```

2. **Significant Earthquakes**
```sql
SELECT * FROM earthquakes 
WHERE magnitude >= 4.5 
AND time BETWEEN '2024-01-01' AND '2024-01-31'
ORDER BY magnitude DESC;
```

3. **Geographic Region**
```sql
SELECT * FROM earthquakes 
WHERE latitude BETWEEN 32.0 AND 42.0 
AND longitude BETWEEN -125.0 AND -114.0
AND time BETWEEN '2024-01-01' AND '2024-01-31';
```

4. **Collection Statistics**
```sql
SELECT data_type, COUNT(*) as total_collections,
       SUM(records_collected) as total_records
FROM collection_logs 
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY data_type;
```

### Performance Tips

1. **Use appropriate indexes**: The schema includes indexes for common query patterns
2. **Limit result sets**: Use LIMIT clauses for large datasets
3. **Partition large tables**: Consider table partitioning for very large datasets
4. **Monitor query performance**: Use EXPLAIN ANALYZE for slow queries
5. **Regular maintenance**: Run VACUUM and ANALYZE regularly

### Database Configuration

Recommended PostgreSQL configuration for production:

```ini
# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# Checkpoint settings
checkpoint_completion_target = 0.9
wal_buffers = 16MB

# Logging
log_statement = 'all'
log_duration = on
log_min_duration_statement = 1000

# Connection settings
max_connections = 100
```

## Backup and Recovery

### Backup Strategies

#### Automated Backups

```bash
#!/bin/bash
# Daily backup script
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups/quakewatch"
DB_NAME="quakewatch"

# Create backup directory
mkdir -p $BACKUP_DIR

# Full database backup
pg_dump -h localhost -U postgres -d $DB_NAME \
    --format=custom --compress=9 \
    --file="$BACKUP_DIR/quakewatch_$DATE.dump"

# Keep only last 7 days of backups
find $BACKUP_DIR -name "quakewatch_*.dump" -mtime +7 -delete
```

#### Backup Types

1. **Full Backup**: Complete database dump
2. **Incremental Backup**: Only changed data since last backup
3. **Point-in-Time Recovery**: WAL-based recovery

### Recovery Procedures

#### Full Database Recovery

```bash
# Stop the application
systemctl stop quakewatch-scraper

# Restore from backup
pg_restore -h localhost -U postgres -d quakewatch \
    --clean --if-exists \
    /backups/quakewatch/quakewatch_20240101_120000.dump

# Restart the application
systemctl start quakewatch-scraper
```

#### Point-in-Time Recovery

```bash
# Create recovery.conf
cat > recovery.conf << EOF
restore_command = 'cp /var/lib/postgresql/wal/%f %p'
recovery_target_time = '2024-01-01 12:00:00'
EOF

# Start recovery
pg_ctl start -D /var/lib/postgresql/data
```

### Monitoring and Maintenance

#### Regular Maintenance Tasks

```sql
-- Update table statistics
ANALYZE earthquakes;
ANALYZE faults;
ANALYZE collection_logs;

-- Clean up old data (optional)
DELETE FROM collection_logs 
WHERE created_at < NOW() - INTERVAL '1 year';

-- Vacuum tables
VACUUM earthquakes;
VACUUM faults;
```

#### Health Checks

```sql
-- Check table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;

-- Check slow queries
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

## Security Considerations

### Access Control

```sql
-- Create application user
CREATE USER quakewatch_app WITH PASSWORD 'secure_password';

-- Grant necessary permissions
GRANT CONNECT ON DATABASE quakewatch TO quakewatch_app;
GRANT USAGE ON SCHEMA public TO quakewatch_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO quakewatch_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO quakewatch_app;
```

### Data Encryption

- **At Rest**: Use PostgreSQL's built-in encryption or filesystem encryption
- **In Transit**: Use SSL/TLS connections
- **Sensitive Data**: Consider encrypting sensitive fields

### Audit Logging

```sql
-- Enable audit logging
CREATE EXTENSION IF NOT EXISTS pgaudit;

-- Configure audit settings
ALTER SYSTEM SET pgaudit.log = 'write, ddl';
ALTER SYSTEM SET pgaudit.log_relation = on;
```

This comprehensive database schema documentation provides all the information needed to understand, maintain, and optimize the QuakeWatch Scraper database system. 
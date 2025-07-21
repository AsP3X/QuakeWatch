# QuakeWatch - Database Structure Plan

## Project Overview
The QuakeWatch Database is a PostgreSQL-based system with PostGIS extension designed to store, manage, and query fault data with geographical capabilities. This database serves as the central data layer for both the data scraper and web application components.

## Database Technology Stack

### Core Database
- **Primary Database**: PostgreSQL 15+
- **Geospatial Extension**: PostGIS 3.3+
- **Connection Pooling**: PgBouncer (production)
- **Backup Solution**: pg_dump with WAL archiving
- **Monitoring**: pg_stat_statements, pg_stat_monitor

### Development Environment
- **Container**: PostgreSQL in Docker (already configured in devcontainer)
- **Database Name**: quakewatch
- **Default User**: quakewatch_user
- **Port**: 5432

## Database Schema Design

### 1. Core Fault Data Tables

#### Faults Table
```sql
-- Main faults table storing fault line data
CREATE TABLE faults (
    id SERIAL PRIMARY KEY,
    catalog_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255),
    catalog_name VARCHAR(100),
    slip_type VARCHAR(50),
    
    -- Slip rate data (parsed from string format)
    net_slip_rate_min DECIMAL(10,3),
    net_slip_rate_max DECIMAL(10,3),
    net_slip_rate_avg DECIMAL(10,3),
    net_slip_rate_unit VARCHAR(10) DEFAULT 'mm/yr',
    
    -- Dip angle data
    average_dip_min DECIMAL(5,2),
    average_dip_max DECIMAL(5,2),
    average_dip_avg DECIMAL(5,2),
    dip_direction VARCHAR(10),
    
    -- Rake angle data
    average_rake_min DECIMAL(5,2),
    average_rake_max DECIMAL(5,2),
    average_rake_avg DECIMAL(5,2),
    
    -- Seismic depth data
    upper_seis_depth_min DECIMAL(8,3),
    upper_seis_depth_max DECIMAL(8,3),
    upper_seis_depth_avg DECIMAL(8,3),
    lower_seis_depth_min DECIMAL(8,3),
    lower_seis_depth_max DECIMAL(8,3),
    lower_seis_depth_avg DECIMAL(8,3),
    depth_unit VARCHAR(10) DEFAULT 'km',
    
    -- Geographical data
    geometry GEOMETRY(LINESTRING, 4326),
    bbox GEOMETRY(POLYGON, 4326),
    centroid GEOMETRY(POINT, 4326),
    
    -- Metadata
    raw_properties JSONB,
    confidence_score DECIMAL(3,2),
    data_quality VARCHAR(20),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_geometry CHECK (ST_IsValid(geometry)),
    CONSTRAINT valid_slip_rate CHECK (net_slip_rate_min <= net_slip_rate_max),
    CONSTRAINT valid_dip CHECK (average_dip_min <= average_dip_max AND average_dip_min >= 0 AND average_dip_max <= 90),
    CONSTRAINT valid_rake CHECK (average_rake_min <= average_rake_max AND average_rake_min >= -180 AND average_rake_max <= 180),
    CONSTRAINT valid_depth CHECK (upper_seis_depth_min <= upper_seis_depth_max AND lower_seis_depth_min <= lower_seis_depth_max)
);

-- Indexes for performance
CREATE INDEX idx_faults_catalog_id ON faults(catalog_id);
CREATE INDEX idx_faults_name ON faults(name);
CREATE INDEX idx_faults_slip_type ON faults(slip_type);
CREATE INDEX idx_faults_geometry ON faults USING GIST(geometry);
CREATE INDEX idx_faults_bbox ON faults USING GIST(bbox);
CREATE INDEX idx_faults_centroid ON faults USING GIST(centroid);
CREATE INDEX idx_faults_created_at ON faults(created_at);
CREATE INDEX idx_faults_updated_at ON faults(updated_at);
```

#### Fault Segments Table
```sql
-- Store individual fault segments for complex fault systems
CREATE TABLE fault_segments (
    id SERIAL PRIMARY KEY,
    fault_id INTEGER REFERENCES faults(id) ON DELETE CASCADE,
    segment_name VARCHAR(255),
    segment_order INTEGER,
    
    -- Segment-specific properties
    segment_slip_rate_min DECIMAL(10,3),
    segment_slip_rate_max DECIMAL(10,3),
    segment_slip_rate_avg DECIMAL(10,3),
    
    -- Geometry for individual segment
    geometry GEOMETRY(LINESTRING, 4326),
    
    -- Metadata
    confidence_score DECIMAL(3,2),
    notes TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_segment_geometry CHECK (ST_IsValid(geometry))
);

CREATE INDEX idx_fault_segments_fault_id ON fault_segments(fault_id);
CREATE INDEX idx_fault_segments_geometry ON fault_segments USING GIST(geometry);
```

### 2. Data Collection and Logging Tables

#### Collection Logs Table
```sql
-- Track data collection operations
CREATE TABLE collection_logs (
    id SERIAL PRIMARY KEY,
    collection_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('success', 'error', 'partial', 'skipped')),
    
    -- Collection statistics
    records_fetched INTEGER,
    records_processed INTEGER,
    records_stored INTEGER,
    records_updated INTEGER,
    records_deleted INTEGER,
    errors_count INTEGER,
    
    -- Performance metrics
    execution_time_ms INTEGER,
    api_response_time_ms INTEGER,
    database_time_ms INTEGER,
    
    -- Error details
    error_details TEXT,
    error_stack TEXT,
    
    -- Source information
    source_url VARCHAR(500),
    source_version VARCHAR(50),
    collection_method VARCHAR(50)
);

CREATE INDEX idx_collection_logs_date ON collection_logs(collection_date);
CREATE INDEX idx_collection_logs_status ON collection_logs(status);
```

#### Data Quality Logs Table
```sql
-- Track data quality issues
CREATE TABLE data_quality_logs (
    id SERIAL PRIMARY KEY,
    fault_id INTEGER REFERENCES faults(id) ON DELETE CASCADE,
    issue_type VARCHAR(50),
    severity VARCHAR(20) CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT,
    field_name VARCHAR(100),
    expected_value TEXT,
    actual_value TEXT,
    resolution_status VARCHAR(20) DEFAULT 'open',
    resolved_at TIMESTAMP,
    resolved_by VARCHAR(100),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_data_quality_logs_fault_id ON data_quality_logs(fault_id);
CREATE INDEX idx_data_quality_logs_issue_type ON data_quality_logs(issue_type);
CREATE INDEX idx_data_quality_logs_severity ON data_quality_logs(severity);
CREATE INDEX idx_data_quality_logs_status ON data_quality_logs(resolution_status);
```

### 3. Geographical and Reference Tables

#### Regions Table
```sql
-- Define geographical regions for grouping and analysis
CREATE TABLE regions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    code VARCHAR(10) UNIQUE,
    description TEXT,
    
    -- Geographical boundary
    geometry GEOMETRY(POLYGON, 4326),
    bbox GEOMETRY(POLYGON, 4326),
    
    -- Metadata
    country VARCHAR(100),
    continent VARCHAR(50),
    tectonic_plate VARCHAR(100),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT valid_region_geometry CHECK (ST_IsValid(geometry))
);

CREATE INDEX idx_regions_name ON regions(name);
CREATE INDEX idx_regions_code ON regions(code);
CREATE INDEX idx_regions_geometry ON regions USING GIST(geometry);
```

#### Fault Types Reference Table
```sql
-- Reference table for fault types
CREATE TABLE fault_types (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    
    -- Color coding for visualization
    color_hex VARCHAR(7),
    color_rgb VARCHAR(20),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert common fault types
INSERT INTO fault_types (code, name, description, category, color_hex) VALUES
('REVERSE', 'Reverse Fault', 'Compressional fault where hanging wall moves up', 'Compressional', '#DC143C'),
('NORMAL', 'Normal Fault', 'Extensional fault where hanging wall moves down', 'Extensional', '#4169E1'),
('STRIKE_SLIP', 'Strike-Slip Fault', 'Horizontal fault with lateral movement', 'Strike-Slip', '#FF8C00'),
('THRUST', 'Thrust Fault', 'Low-angle reverse fault', 'Compressional', '#8B0000'),
('OBLIQUE', 'Oblique Fault', 'Combination of strike-slip and dip-slip movement', 'Mixed', '#9932CC');
```

### 4. Analytics and Statistics Tables

#### Fault Statistics Table
```sql
-- Pre-calculated statistics for performance
CREATE TABLE fault_statistics (
    id SERIAL PRIMARY KEY,
    region_id INTEGER REFERENCES regions(id),
    fault_type_id INTEGER REFERENCES fault_types(id),
    
    -- Counts
    total_faults INTEGER,
    active_faults INTEGER,
    high_risk_faults INTEGER,
    
    -- Averages
    avg_slip_rate DECIMAL(10,3),
    avg_dip DECIMAL(5,2),
    avg_depth DECIMAL(8,3),
    
    -- Ranges
    min_slip_rate DECIMAL(10,3),
    max_slip_rate DECIMAL(10,3),
    min_depth DECIMAL(8,3),
    max_depth DECIMAL(8,3),
    
    -- Total lengths
    total_fault_length DECIMAL(12,3),
    avg_fault_length DECIMAL(8,3),
    
    -- Calculation metadata
    calculation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    data_freshness_hours INTEGER,
    
    UNIQUE(region_id, fault_type_id)
);

CREATE INDEX idx_fault_statistics_region ON fault_statistics(region_id);
CREATE INDEX idx_fault_statistics_type ON fault_statistics(fault_type_id);
```

#### Search Index Table
```sql
-- Full-text search index for fault names and descriptions
CREATE TABLE fault_search_index (
    id SERIAL PRIMARY KEY,
    fault_id INTEGER REFERENCES faults(id) ON DELETE CASCADE,
    search_vector tsvector,
    
    UNIQUE(fault_id)
);

CREATE INDEX idx_fault_search_vector ON fault_search_index USING GIN(search_vector);

-- Trigger to update search vector
CREATE OR REPLACE FUNCTION update_fault_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.catalog_name, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.slip_type, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_fault_search_vector
    BEFORE INSERT OR UPDATE ON faults
    FOR EACH ROW
    EXECUTE FUNCTION update_fault_search_vector();
```

## Database Functions and Views

### 1. Geographical Functions

#### Distance Calculation Function
```sql
-- Calculate distance between fault and point
CREATE OR REPLACE FUNCTION fault_distance_to_point(
    fault_geom GEOMETRY,
    point_lat DECIMAL,
    point_lon DECIMAL
) RETURNS DECIMAL AS $$
BEGIN
    RETURN ST_Distance(
        fault_geom,
        ST_SetSRID(ST_MakePoint(point_lon, point_lat), 4326)
    );
END;
$$ LANGUAGE plpgsql;
```

#### Fault Intersection Function
```sql
-- Find faults within specified radius of point
CREATE OR REPLACE FUNCTION find_faults_near_point(
    point_lat DECIMAL,
    point_lon DECIMAL,
    radius_km DECIMAL DEFAULT 100
) RETURNS TABLE (
    fault_id INTEGER,
    fault_name VARCHAR,
    distance_km DECIMAL,
    slip_type VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        f.id,
        f.name,
        ST_Distance(
            f.geometry,
            ST_SetSRID(ST_MakePoint(point_lon, point_lat), 4326)
        ) / 1000 as distance_km,
        f.slip_type
    FROM faults f
    WHERE ST_DWithin(
        f.geometry,
        ST_SetSRID(ST_MakePoint(point_lon, point_lat), 4326),
        radius_km * 1000
    )
    ORDER BY distance_km;
END;
$$ LANGUAGE plpgsql;
```

### 2. Data Quality Functions

#### Data Completeness Check
```sql
-- Check data completeness for a fault
CREATE OR REPLACE FUNCTION check_fault_data_completeness(fault_id_param INTEGER)
RETURNS TABLE (
    field_name VARCHAR,
    is_populated BOOLEAN,
    completeness_percentage DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'name' as field_name,
        (f.name IS NOT NULL AND f.name != '') as is_populated,
        CASE WHEN f.name IS NOT NULL AND f.name != '' THEN 100.0 ELSE 0.0 END as completeness_percentage
    FROM faults f WHERE f.id = fault_id_param
    
    UNION ALL
    
    SELECT 
        'slip_rate' as field_name,
        (f.net_slip_rate_avg IS NOT NULL) as is_populated,
        CASE WHEN f.net_slip_rate_avg IS NOT NULL THEN 100.0 ELSE 0.0 END as completeness_percentage
    FROM faults f WHERE f.id = fault_id_param
    
    UNION ALL
    
    SELECT 
        'geometry' as field_name,
        (f.geometry IS NOT NULL) as is_populated,
        CASE WHEN f.geometry IS NOT NULL THEN 100.0 ELSE 0.0 END as completeness_percentage
    FROM faults f WHERE f.id = fault_id_param;
END;
$$ LANGUAGE plpgsql;
```

### 3. Analytics Views

#### Fault Summary View
```sql
-- Comprehensive fault summary view
CREATE VIEW fault_summary AS
SELECT 
    f.id,
    f.catalog_id,
    f.name,
    f.catalog_name,
    f.slip_type,
    ft.name as fault_type_name,
    ft.color_hex as fault_color,
    
    -- Slip rate information
    f.net_slip_rate_min,
    f.net_slip_rate_max,
    f.net_slip_rate_avg,
    
    -- Geometry information
    ST_Length(f.geometry) as fault_length_km,
    ST_X(ST_Centroid(f.geometry)) as centroid_lon,
    ST_Y(ST_Centroid(f.geometry)) as centroid_lat,
    
    -- Data quality
    CASE 
        WHEN f.name IS NOT NULL AND f.geometry IS NOT NULL AND f.net_slip_rate_avg IS NOT NULL THEN 'complete'
        WHEN f.name IS NOT NULL AND f.geometry IS NOT NULL THEN 'partial'
        ELSE 'incomplete'
    END as data_quality,
    
    f.created_at,
    f.updated_at
    
FROM faults f
LEFT JOIN fault_types ft ON f.slip_type = ft.code
WHERE f.geometry IS NOT NULL;
```

#### Regional Statistics View
```sql
-- Regional fault statistics
CREATE VIEW regional_fault_stats AS
SELECT 
    r.id as region_id,
    r.name as region_name,
    r.country,
    r.continent,
    
    COUNT(f.id) as total_faults,
    COUNT(CASE WHEN f.net_slip_rate_avg > 5 THEN 1 END) as high_slip_rate_faults,
    COUNT(CASE WHEN f.net_slip_rate_avg > 10 THEN 1 END) as very_high_slip_rate_faults,
    
    AVG(f.net_slip_rate_avg) as avg_slip_rate,
    MAX(f.net_slip_rate_avg) as max_slip_rate,
    MIN(f.net_slip_rate_avg) as min_slip_rate,
    
    AVG(ST_Length(f.geometry)) as avg_fault_length_km,
    SUM(ST_Length(f.geometry)) as total_fault_length_km
    
FROM regions r
LEFT JOIN faults f ON ST_Intersects(r.geometry, f.geometry)
WHERE f.geometry IS NOT NULL
GROUP BY r.id, r.name, r.country, r.continent;
```

## Database Configuration

### 1. PostGIS Setup
```sql
-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- Enable additional useful extensions
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gin;
```

### 2. Performance Configuration
```sql
-- Set appropriate configuration for geospatial queries
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
```

### 3. Connection Pooling (Production)
```ini
# pgbouncer.ini configuration
[databases]
quakewatch = host=localhost port=5432 dbname=quakewatch

[pgbouncer]
listen_addr = 127.0.0.1
listen_port = 6432
auth_type = md5
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 20
```

## Data Migration and Versioning

### 1. Migration Scripts
```sql
-- Migration script template
-- migration_001_initial_schema.sql

BEGIN;

-- Create initial tables
CREATE TABLE IF NOT EXISTS faults (
    -- table definition
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_faults_catalog_id ON faults(catalog_id);

-- Insert initial data
INSERT INTO fault_types (code, name, description) VALUES
('REVERSE', 'Reverse Fault', 'Compressional fault');

COMMIT;
```

### 2. Version Control
```sql
-- Schema version tracking
CREATE TABLE schema_versions (
    id SERIAL PRIMARY KEY,
    version VARCHAR(20) UNIQUE NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT,
    checksum VARCHAR(64)
);

-- Function to apply migrations
CREATE OR REPLACE FUNCTION apply_migration(
    version_param VARCHAR,
    description_param TEXT,
    checksum_param VARCHAR
) RETURNS VOID AS $$
BEGIN
    INSERT INTO schema_versions (version, description, checksum)
    VALUES (version_param, description_param, checksum_param);
END;
$$ LANGUAGE plpgsql;
```

## Backup and Recovery Strategy

### 1. Backup Configuration
```bash
#!/bin/bash
# backup_script.sh

# Daily full backup
pg_dump -h localhost -U quakewatch_user -d quakewatch \
    --format=custom --compress=9 \
    --file=/backups/quakewatch_$(date +%Y%m%d).dump

# WAL archiving for point-in-time recovery
# Add to postgresql.conf:
# archive_mode = on
# archive_command = 'cp %p /backups/wal/%f'
```

### 2. Recovery Procedures
```sql
-- Point-in-time recovery example
-- pg_restore -h localhost -U quakewatch_user -d quakewatch \
--     --clean --if-exists \
--     /backups/quakewatch_20240101.dump

-- Recover to specific point in time
-- pg_restore -h localhost -U quakewatch_user -d quakewatch \
--     --clean --if-exists \
--     --recovery-target-time="2024-01-01 12:00:00" \
--     /backups/quakewatch_20240101.dump
```

## Monitoring and Maintenance

### 1. Performance Monitoring
```sql
-- Query performance monitoring
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements
WHERE query LIKE '%faults%'
ORDER BY total_time DESC
LIMIT 10;

-- Table size monitoring
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### 2. Maintenance Tasks
```sql
-- Regular maintenance procedures
-- 1. Update statistics
ANALYZE faults;
ANALYZE fault_segments;
ANALYZE collection_logs;

-- 2. Vacuum tables
VACUUM ANALYZE faults;
VACUUM ANALYZE fault_segments;

-- 3. Reindex if needed
REINDEX INDEX CONCURRENTLY idx_faults_geometry;
```

## Security Considerations

### 1. User Management
```sql
-- Create application user with limited privileges
CREATE USER quakewatch_app WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE quakewatch TO quakewatch_app;
GRANT USAGE ON SCHEMA public TO quakewatch_app;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO quakewatch_app;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO quakewatch_app;

-- Create read-only user for analytics
CREATE USER quakewatch_readonly WITH PASSWORD 'readonly_password';
GRANT CONNECT ON DATABASE quakewatch TO quakewatch_readonly;
GRANT USAGE ON SCHEMA public TO quakewatch_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO quakewatch_readonly;
```

### 2. Row Level Security (if needed)
```sql
-- Example RLS for sensitive data
ALTER TABLE faults ENABLE ROW LEVEL SECURITY;

CREATE POLICY fault_access_policy ON faults
    FOR ALL
    TO quakewatch_app
    USING (true);
```

## Implementation Timeline

### Phase 1: Database Setup (Week 1)
1. **Environment Setup**
   - Configure PostgreSQL with PostGIS
   - Set up development database
   - Configure connection pooling

2. **Schema Implementation**
   - Create core tables (faults, fault_segments)
   - Implement indexes and constraints
   - Set up data quality logging

### Phase 2: Advanced Features (Week 2)
1. **Geographical Functions**
   - Implement spatial queries
   - Create distance calculation functions
   - Set up regional analysis

2. **Analytics and Views**
   - Create summary views
   - Implement statistics tables
   - Set up search functionality

### Phase 3: Optimization (Week 3)
1. **Performance Tuning**
   - Optimize query performance
   - Implement caching strategies
   - Set up monitoring

2. **Backup and Recovery**
   - Configure backup procedures
   - Test recovery processes
   - Document maintenance procedures

## Success Metrics

### Performance Metrics
- **Query Response Time**: < 100ms for common queries
- **Spatial Query Performance**: < 500ms for complex spatial operations
- **Index Efficiency**: > 95% index usage for common queries
- **Storage Efficiency**: Optimized data storage with compression

### Data Quality Metrics
- **Data Completeness**: > 95% of required fields populated
- **Geometrical Validity**: 100% valid geometries
- **Referential Integrity**: No orphaned records
- **Data Freshness**: < 6 hours old data

### Operational Metrics
- **Uptime**: > 99.9% availability
- **Backup Success Rate**: 100% successful backups
- **Recovery Time**: < 30 minutes for full recovery
- **Migration Success Rate**: 100% successful migrations

## Conclusion

The QuakeWatch Database provides a robust, scalable foundation for storing and querying fault data with advanced geographical capabilities. The comprehensive schema design supports both the data scraper and web application while maintaining data integrity and performance.

The database architecture is designed for future growth and can accommodate additional data sources, enhanced analytics, and advanced geographical features as the project evolves. 
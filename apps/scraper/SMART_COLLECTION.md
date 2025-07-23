# Smart Collection Feature

The smart collection feature allows you to collect earthquake data on intervals without producing duplicates. This is achieved through time-based filtering and collection metadata tracking.

## How It Works

1. **Collection Tracking**: The system tracks the last collection time for each data type
2. **Time-Based Filtering**: New collections only fetch data since the last collection time
3. **Database Upsert**: PostgreSQL storage uses upsert logic to prevent duplicate records
4. **Collection Logging**: All collection operations are logged with metadata

## Features

### Smart Collection Methods

- `CollectRecentEarthquakesSmart()`: Collects earthquakes since last collection time
- `CollectRecentEarthquakes()`: Collects earthquakes for a specified time range
- Collection metadata tracking with `GetLastCollectionTime()` and `UpdateLastCollectionTime()`
- Collection logging with `LogCollection()` and `GetCollectionLogs()`

### Storage Support

- **PostgreSQL**: Full support with upsert logic and collection metadata tables
- **JSON**: Basic support with file-based metadata tracking

## Usage Examples

### Command Line Interface

```bash
# Smart collection (avoids duplicates)
./quakewatch-scraper earthquakes recent --smart --storage postgresql

# Time-based collection (last 3 hours)
./quakewatch-scraper earthquakes recent --hours-back 3 --storage postgresql

# Smart collection with JSON storage
./quakewatch-scraper earthquakes recent --smart --storage json
```

### Programmatic Usage

```go
// Initialize storage
storage := storage.NewPostgreSQLStorage(dbConfig)
collector := collector.NewEarthquakeCollector(usgsClient, nil)

ctx := context.Background()

// Smart collection
err := collector.CollectRecentEarthquakesSmart(ctx, storage)

// Time-based collection
earthquakes, err := collector.CollectRecentEarthquakes(ctx, 2) // 2 hours
err = storage.SaveEarthquakes(ctx, earthquakes)

// Check collection logs
logs, err := storage.GetCollectionLogs(ctx, "earthquakes", 10)

// Get last collection time
lastTime, err := storage.GetLastCollectionTime(ctx, "earthquakes")
```

## Database Schema

### Collection Metadata Table

```sql
CREATE TABLE collection_metadata (
    data_type VARCHAR(50) PRIMARY KEY,
    last_collection_time BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Collection Logs Table

```sql
CREATE TABLE collection_logs (
    id BIGSERIAL PRIMARY KEY,
    data_type VARCHAR(50) NOT NULL,
    source VARCHAR(100) NOT NULL,
    start_time BIGINT NOT NULL,
    end_time BIGINT,
    records_collected INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    created_at BIGINT NOT NULL
);
```

## Configuration

### CLI Flags

- `--smart`: Enable smart collection mode
- `--storage`: Storage backend (json, postgresql)
- `--hours-back`: Number of hours to look back for time-based collection

### Environment Variables

For PostgreSQL storage:
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_SSL_MODE`: SSL mode

## Best Practices

1. **Use Smart Collection for Intervals**: Always use `--smart` flag for interval-based collection
2. **Monitor Collection Logs**: Regularly check collection logs for errors or issues
3. **Database Backend**: Use PostgreSQL for production environments with high data volumes
4. **Backup Metadata**: Regularly backup collection metadata tables
5. **Error Handling**: Implement proper error handling for collection failures

## Migration

Run the database migration to create the required tables:

```bash
# Using golang-migrate
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up

# Or manually
psql -d your_database -f migrations/001_create_collection_metadata.up.sql
```

## Example Output

```
Smart collection: collecting earthquakes from 2024-01-15 10:00:00 to 2024-01-15 11:00:00 (since last collection: 2024-01-15 10:00:00)...
Found 15 earthquakes
Successfully collected and saved 15 earthquakes
```

## Troubleshooting

### Common Issues

1. **No Data Collected**: Check if the time range is appropriate
2. **Database Connection**: Verify database credentials and connectivity
3. **Migration Issues**: Ensure collection metadata tables exist
4. **Permission Errors**: Check file/database permissions

### Debug Commands

```bash
# Check collection logs
./quakewatch-scraper stats --storage postgresql

# Verify database connection
./quakewatch-scraper health --storage postgresql

# List recent collections
./quakewatch-scraper list --storage postgresql
``` 
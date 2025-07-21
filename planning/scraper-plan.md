# QuakeWatch - Data Scraper Plan

## Project Overview
The QuakeWatch Data Scraper is a dedicated Node.js application designed to fetch, validate, clean, and store fault data from the EMSC-CSEM (European-Mediterranean Seismological Centre) API. This component serves as the data ingestion layer for the entire QuakeWatch system.

## Data Source
- **URL**: https://www.emsc-csem.org/javascript/gem_active_faults.geojson
- **Format**: GeoJSON
- **Content**: Active fault data with geographical coordinates and fault properties
- **Update Frequency**: Variable (to be determined based on API availability)

## Scraper Architecture

### Core Components

#### 1. Data Fetcher
**Purpose**: Retrieve GeoJSON data from EMSC-CSEM API

**Features**:
- HTTP client using Axios for reliable data fetching
- Retry mechanism with exponential backoff
- Rate limiting to respect API constraints
- Request timeout handling
- User-Agent and proper headers

**Implementation**:
```javascript
// Example structure
class DataFetcher {
  async fetchFaultData()
  async validateResponse()
  async handleRetries()
  async applyRateLimiting()
}
```

#### 2. Data Validator
**Purpose**: Validate incoming data structure and integrity

**Features**:
- Schema validation using Joi
- GeoJSON format validation
- Coordinate range validation
- Required field checks
- Data type validation

**Validation Schema**:
```javascript
const faultSchema = Joi.object({
  type: Joi.string().valid('FeatureCollection').required(),
  features: Joi.array().items(
    Joi.object({
      type: Joi.string().valid('Feature').required(),
      properties: Joi.object({
        name: Joi.string(),
        catalog_id: Joi.string(),
        catalog_name: Joi.string(),
        slip_type: Joi.string(),
        net_slip_rate: Joi.string(),
        average_dip: Joi.string(),
        average_rake: Joi.string(),
        upper_seis_depth: Joi.string(),
        lower_seis_depth: Joi.string()
      }),
      geometry: Joi.object({
        type: Joi.string().valid('LineString').required(),
        coordinates: Joi.array().items(
          Joi.array().items(Joi.number()).min(2).max(2)
        ).min(2)
      })
    })
  ).min(1)
});
```

#### 3. Data Cleaner
**Purpose**: Transform and normalize data for database storage

**Features**:
- Parse string values to appropriate data types
- Normalize coordinate systems
- Handle missing or null values
- Generate unique identifiers
- Extract and validate fault properties

**Data Transformations**:
```javascript
// Example transformations
const cleanFaultData = (rawData) => {
  return rawData.features.map(feature => ({
    id: generateUniqueId(feature.properties.catalog_id),
    name: feature.properties.name || 'Unnamed Fault',
    catalogId: feature.properties.catalog_id,
    catalogName: feature.properties.catalog_name,
    slipType: feature.properties.slip_type,
    netSlipRate: parseSlipRate(feature.properties.net_slip_rate),
    averageDip: parseAngle(feature.properties.average_dip),
    averageRake: parseAngle(feature.properties.average_rake),
    upperSeisDepth: parseDepth(feature.properties.upper_seis_depth),
    lowerSeisDepth: parseDepth(feature.properties.lower_seis_depth),
    coordinates: feature.geometry.coordinates,
    geometry: feature.geometry,
    createdAt: new Date(),
    updatedAt: new Date()
  }));
};
```

#### 4. Scheduler
**Purpose**: Automate data collection at regular intervals

**Features**:
- Configurable scheduling using node-cron
- Multiple schedule options (hourly, daily, weekly)
- Manual trigger capability
- Schedule management and monitoring

**Scheduling Options**:
```javascript
// Example schedules
const schedules = {
  hourly: '0 * * * *',
  daily: '0 0 * * *',
  weekly: '0 0 * * 0',
  custom: '0 */6 * * *' // Every 6 hours
};
```

#### 5. Logger
**Purpose**: Comprehensive logging for monitoring and debugging

**Features**:
- Structured logging with Winston
- Multiple log levels (error, warn, info, debug)
- Log rotation and archiving
- Performance metrics logging
- Error tracking and reporting

**Log Categories**:
- Data fetching operations
- Validation results
- Cleaning statistics
- Database operations
- System health metrics

### Database Integration

#### Storage Strategy
- **Primary Database**: PostgreSQL with PostGIS extension
- **Backup Strategy**: Automated backups with point-in-time recovery
- **Data Versioning**: Track changes and maintain historical data
- **Indexing**: Geospatial indexes for efficient querying

#### Schema Design
```sql
-- Faults table
CREATE TABLE faults (
  id SERIAL PRIMARY KEY,
  catalog_id VARCHAR(50) UNIQUE NOT NULL,
  name VARCHAR(255),
  catalog_name VARCHAR(100),
  slip_type VARCHAR(50),
  net_slip_rate_min DECIMAL(10,3),
  net_slip_rate_max DECIMAL(10,3),
  net_slip_rate_avg DECIMAL(10,3),
  average_dip_min DECIMAL(5,2),
  average_dip_max DECIMAL(5,2),
  average_dip_avg DECIMAL(5,2),
  average_rake_min DECIMAL(5,2),
  average_rake_max DECIMAL(5,2),
  average_rake_avg DECIMAL(5,2),
  upper_seis_depth_min DECIMAL(8,3),
  upper_seis_depth_max DECIMAL(8,3),
  upper_seis_depth_avg DECIMAL(8,3),
  lower_seis_depth_min DECIMAL(8,3),
  lower_seis_depth_max DECIMAL(8,3),
  lower_seis_depth_avg DECIMAL(8,3),
  geometry GEOMETRY(LINESTRING, 4326),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Data collection logs
CREATE TABLE collection_logs (
  id SERIAL PRIMARY KEY,
  collection_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  status VARCHAR(20), -- success, error, partial
  records_fetched INTEGER,
  records_processed INTEGER,
  records_stored INTEGER,
  errors_count INTEGER,
  execution_time_ms INTEGER,
  error_details TEXT
);
```

## Implementation Phases

### Phase 1: Core Scraper Development (Week 1-2)

#### Week 1: Foundation
1. **Project Setup**
   - Initialize Node.js project with TypeScript
   - Install dependencies (axios, joi, winston, node-cron, pg)
   - Configure ESLint, Prettier, and Jest
   - Set up environment configuration

2. **Data Fetcher Implementation**
   - Create HTTP client wrapper
   - Implement retry logic with exponential backoff
   - Add rate limiting functionality
   - Create response validation

3. **Data Validator Implementation**
   - Define Joi schemas for data validation
   - Implement coordinate validation
   - Add data type checking
   - Create validation error reporting

#### Week 2: Data Processing
1. **Data Cleaner Implementation**
   - Create data transformation functions
   - Implement coordinate normalization
   - Add data type parsing
   - Handle edge cases and missing data

2. **Logging System**
   - Set up Winston logger configuration
   - Create structured logging format
   - Implement log rotation
   - Add performance metrics

3. **Error Handling**
   - Implement comprehensive error handling
   - Create error recovery mechanisms
   - Add error reporting and alerting
   - Build error categorization

### Phase 2: Database Integration (Week 3)

#### Database Setup
1. **Database Design**
   - Set up PostgreSQL with PostGIS
   - Create database schema
   - Implement geospatial indexing
   - Set up connection pooling

2. **Data Storage Implementation**
   - Create database connection module
   - Implement data insertion logic
   - Add upsert functionality for updates
   - Create data migration scripts

3. **Integration Testing**
   - Test end-to-end data flow
   - Validate data integrity
   - Performance testing
   - Error scenario testing

### Phase 3: Automation and Monitoring (Week 4)

#### Scheduling and Automation
1. **Scheduler Implementation**
   - Configure cron jobs for data collection
   - Implement manual trigger functionality
   - Add schedule management
   - Create monitoring dashboard

2. **Health Monitoring**
   - Implement health checks
   - Add performance monitoring
   - Create alerting system
   - Build status reporting

3. **Deployment and DevOps**
   - Create Docker containerization
   - Set up CI/CD pipeline
   - Configure production environment
   - Implement backup strategies

## Technical Stack

### Core Dependencies
- **Runtime**: Node.js 18+
- **Language**: TypeScript
- **HTTP Client**: Axios
- **Data Validation**: Joi
- **Logging**: Winston
- **Scheduling**: node-cron
- **Database**: PostgreSQL with PostGIS
- **ORM**: Prisma or TypeORM
- **Testing**: Jest

### Development Tools
- **Code Quality**: ESLint, Prettier
- **Type Checking**: TypeScript
- **Package Management**: npm or yarn
- **Version Control**: Git

### Monitoring and Operations
- **Process Management**: PM2
- **Containerization**: Docker
- **CI/CD**: GitHub Actions
- **Monitoring**: Application performance monitoring

## Configuration

### Environment Variables
```bash
# Database Configuration
DATABASE_URL=postgresql://user:password@localhost:5432/quakewatch
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=quakewatch
POSTGRES_USER=quakewatch_user
POSTGRES_PASSWORD=secure_password

# API Configuration
EMSC_API_URL=https://www.emsc-csem.org/javascript/gem_active_faults.geojson
API_TIMEOUT=30000
API_RETRY_ATTEMPTS=3
API_RATE_LIMIT=1000

# Scheduling Configuration
COLLECTION_SCHEDULE="0 */6 * * *"  # Every 6 hours
MANUAL_TRIGGER_ENABLED=true

# Logging Configuration
LOG_LEVEL=info
LOG_FILE_PATH=./logs/scraper.log
LOG_MAX_SIZE=10m
LOG_MAX_FILES=5

# Monitoring Configuration
HEALTH_CHECK_INTERVAL=300000  # 5 minutes
ALERT_EMAIL=admin@quakewatch.com
```

### Configuration Files
- `config/database.ts` - Database configuration
- `config/api.ts` - API client configuration
- `config/scheduler.ts` - Scheduling configuration
- `config/logging.ts` - Logging configuration

## Error Handling and Recovery

### Error Categories
1. **Network Errors**: API unavailability, timeout, rate limiting
2. **Data Errors**: Invalid format, missing fields, corrupted data
3. **Database Errors**: Connection issues, constraint violations
4. **System Errors**: Memory issues, disk space, process crashes

### Recovery Strategies
- **Automatic Retry**: Exponential backoff for transient errors
- **Fallback Data**: Use cached data when API is unavailable
- **Partial Processing**: Continue processing valid records when some fail
- **Alert System**: Notify administrators of critical failures

## Testing Strategy

### Unit Tests
- Data fetcher functionality
- Validation logic
- Data cleaning transformations
- Error handling scenarios

### Integration Tests
- End-to-end data flow
- Database operations
- API interactions
- Scheduling functionality

### Performance Tests
- Large dataset processing
- Database query performance
- Memory usage optimization
- Concurrent operation handling

## Monitoring and Alerting

### Key Metrics
- **Data Collection Success Rate**: Target > 95%
- **Processing Time**: Target < 5 minutes for full dataset
- **Error Rate**: Target < 1%
- **Database Performance**: Query response times
- **System Resources**: CPU, memory, disk usage

### Alerting Rules
- Data collection failures
- High error rates
- Performance degradation
- System resource issues
- Database connectivity problems

## Deployment

### Containerization
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY dist ./dist
COPY config ./config
EXPOSE 3000
CMD ["node", "dist/index.js"]
```

### Production Considerations
- **High Availability**: Multiple instances with load balancing
- **Data Backup**: Automated daily backups with retention policy
- **Security**: Environment variable management, network security
- **Scaling**: Horizontal scaling capabilities
- **Monitoring**: Comprehensive logging and metrics

## Maintenance and Updates

### Regular Maintenance
- **Data Quality Audits**: Monthly review of data integrity
- **Performance Optimization**: Quarterly performance reviews
- **Security Updates**: Regular dependency updates
- **Backup Verification**: Weekly backup restoration tests

### Update Procedures
- **Schema Migrations**: Version-controlled database changes
- **Code Deployments**: Blue-green deployment strategy
- **Configuration Updates**: Environment-specific configurations
- **Rollback Procedures**: Quick rollback mechanisms

## Success Metrics

### Technical Metrics
- **Data Collection Success Rate**: > 95%
- **Processing Time**: < 5 minutes for full dataset
- **Error Rate**: < 1%
- **System Uptime**: > 99.9%
- **Data Freshness**: < 6 hours old

### Quality Metrics
- **Data Completeness**: > 98% of expected fields populated
- **Data Accuracy**: Coordinate validation success rate > 99%
- **Duplicate Detection**: < 0.1% duplicate records
- **Schema Compliance**: 100% valid GeoJSON structure

## Conclusion

The QuakeWatch Data Scraper provides a robust, scalable foundation for collecting and processing fault data. The modular architecture ensures maintainability and extensibility, while comprehensive error handling and monitoring guarantee reliable operation.

The scraper serves as the critical data ingestion layer for the entire QuakeWatch system, ensuring that the web application always has access to the most current and accurate fault data available. 
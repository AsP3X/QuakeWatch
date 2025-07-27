# Development Guide

This guide provides comprehensive information for developers contributing to the QuakeWatch Scraper project, including setup, coding standards, testing, and contribution workflows.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Database Development](#database-development)
- [API Development](#api-development)
- [CLI Development](#cli-development)
- [Contributing](#contributing)
- [Release Process](#release-process)

## Development Environment Setup

### Prerequisites

- **Go**: 1.24 or higher
- **Git**: Latest version
- **PostgreSQL**: 12.0 or higher (for database development)
- **Docker**: Latest version (optional, for containerized development)
- **Make**: For build automation
- **Code Editor**: VS Code, GoLand, or Vim with Go support

### Initial Setup

1. **Clone the Repository**

```bash
git clone <repository-url>
cd quakewatch-scraper
```

2. **Install Dependencies**

```bash
# Install Go dependencies
make install
# or
go mod download
go mod tidy
```

3. **Set Up Development Environment**

```bash
# Create development directories
make setup

# Set up environment variables
cp .env.example .env
# Edit .env with your local settings
```

4. **Install Development Tools**

```bash
# Install Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/lint/golint@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# Install pre-commit hooks (optional)
go install github.com/pre-commit/pre-commit@latest
pre-commit install
```

### IDE Configuration

#### VS Code

Install the following extensions:
- Go (official)
- Go Test Explorer
- GitLens
- YAML
- Docker

VS Code settings for Go development:

```json
{
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.formatTool": "goimports",
    "go.testFlags": ["-v"],
    "go.buildTags": "",
    "go.toolsManagement.autoUpdate": true,
    "go.gopath": "",
    "go.goroot": "",
    "go.inferGopath": true,
    "go.buildOnSave": "package",
    "go.lintOnSave": "package",
    "go.vetOnSave": "package",
    "go.coverOnSave": false,
    "go.testOnSave": false,
    "go.gocodeAutoBuild": false
}
```

#### GoLand

Configure GoLand for optimal Go development:
- Enable Go modules
- Configure code formatting with goimports
- Set up linting with golangci-lint
- Configure test coverage display

### Database Setup for Development

1. **Local PostgreSQL Setup**

```bash
# Install PostgreSQL
sudo apt install postgresql postgresql-contrib

# Start PostgreSQL service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create development database
sudo -u postgres createdb quakewatch_dev
sudo -u postgres psql -c "CREATE USER quakewatch_dev WITH PASSWORD 'dev_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE quakewatch_dev TO quakewatch_dev;"
```

2. **Docker Database Setup**

```bash
# Start PostgreSQL with Docker
docker run --name quakewatch-postgres-dev \
    -e POSTGRES_DB=quakewatch_dev \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -p 5432:5432 \
    -d postgres:15

# Initialize database
make db-init-dev
make db-migrate-up-dev
```

## Project Structure

### Directory Organization

```
quakewatch-scraper/
├── cmd/                    # Application entry points
│   └── scraper/
│       └── main.go        # Main application entry point
├── internal/              # Private application code
│   ├── api/              # API clients and integrations
│   ├── collector/        # Data collection logic
│   ├── config/           # Configuration management
│   ├── models/           # Data models and structures
│   ├── scheduler/        # Scheduling and daemon logic
│   ├── storage/          # Data storage implementations
│   └── utils/            # Utility functions
├── pkg/                  # Public packages
│   └── cli/              # Command-line interface
├── configs/              # Configuration files
├── migrations/           # Database migrations
├── data/                 # Data storage directory
├── bin/                  # Compiled binaries
├── docs/                 # Documentation
├── scripts/              # Build and deployment scripts
├── tests/                # Integration tests
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Makefile              # Build automation
├── Dockerfile            # Container definition
├── docker-compose.yml    # Container orchestration
└── README.md             # Project documentation
```

### Code Organization Principles

1. **Separation of Concerns**: Each package has a single responsibility
2. **Dependency Injection**: Use interfaces for loose coupling
3. **Error Handling**: Consistent error handling patterns
4. **Configuration**: External configuration management
5. **Testing**: Comprehensive test coverage

### Package Responsibilities

#### `cmd/scraper/`
- Application entry point
- Command-line argument parsing
- Application lifecycle management

#### `internal/api/`
- External API integrations
- HTTP client implementations
- API response handling

#### `internal/collector/`
- Data collection orchestration
- Data validation and cleaning
- Collection strategies

#### `internal/config/`
- Configuration loading and validation
- Environment variable handling
- Default configuration management

#### `internal/models/`
- Data structures and types
- JSON serialization/deserialization
- Data validation methods

#### `internal/scheduler/`
- Interval-based execution
- Daemon process management
- Health monitoring

#### `internal/storage/`
- Data persistence implementations
- Database operations
- File system operations

#### `internal/utils/`
- Common utility functions
- Platform-specific code
- Helper functions

#### `pkg/cli/`
- Command-line interface
- Command implementations
- User interaction handling

## Coding Standards

### Go Code Style

Follow the official Go coding standards:

1. **Formatting**: Use `gofmt` or `goimports`
2. **Naming**: Follow Go naming conventions
3. **Comments**: Document exported functions and types
4. **Error Handling**: Always check and handle errors
5. **Package Organization**: Keep packages focused and cohesive

### Code Formatting

```bash
# Format code
make fmt
# or
go fmt ./...
goimports -w .

# Check code style
make lint
# or
golangci-lint run
```

### Naming Conventions

```go
// Package names: lowercase, single word
package collector

// Function names: camelCase
func collectEarthquakes() error

// Variable names: camelCase
var earthquakeData []Earthquake

// Constant names: camelCase or UPPER_CASE
const (
    defaultTimeout = 30 * time.Second
    MAX_RETRIES    = 3
)

// Type names: PascalCase
type EarthquakeCollector struct{}

// Interface names: PascalCase, often with 'er' suffix
type DataCollector interface{}
```

### Error Handling

```go
// Always check errors
if err != nil {
    return fmt.Errorf("failed to collect data: %w", err)
}

// Use custom error types for specific errors
type CollectionError struct {
    Source string
    Err    error
}

func (e *CollectionError) Error() string {
    return fmt.Sprintf("collection failed for %s: %v", e.Source, e.Err)
}

func (e *CollectionError) Unwrap() error {
    return e.Err
}
```

### Logging

```go
// Use structured logging
log.WithFields(log.Fields{
    "source": "usgs",
    "count":  len(earthquakes),
    "duration": duration,
}).Info("Earthquake collection completed")

// Use appropriate log levels
log.Debug("Processing earthquake data")
log.Info("Collection started")
log.Warn("Rate limit approaching")
log.Error("Collection failed")
```

### Configuration

```go
// Use environment variables with defaults
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
    API      APIConfig      `mapstructure:"api"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host" env:"DB_HOST" env-default:"localhost"`
    Port     int    `mapstructure:"port" env:"DB_PORT" env-default:"5432"`
    Username string `mapstructure:"username" env:"DB_USER" env-default:"postgres"`
    Password string `mapstructure:"password" env:"DB_PASSWORD"`
}
```

## Testing

### Testing Strategy

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows
4. **Performance Tests**: Test performance characteristics

### Running Tests

```bash
# Run all tests
make test
# or
go test ./...

# Run tests with coverage
make test-coverage
# or
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific test
go test ./internal/collector -v

# Run benchmark tests
go test -bench=. ./internal/collector
```

### Test Structure

```go
// internal/collector/earthquake_test.go
package collector

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestEarthquakeCollector_CollectRecent(t *testing.T) {
    // Arrange
    collector := NewEarthquakeCollector()
    
    // Act
    earthquakes, err := collector.CollectRecent(10)
    
    // Assert
    require.NoError(t, err)
    assert.Len(t, earthquakes, 10)
    assert.True(t, time.Since(earthquakes[0].Time) < time.Hour)
}

func TestEarthquakeCollector_CollectRecent_InvalidLimit(t *testing.T) {
    // Arrange
    collector := NewEarthquakeCollector()
    
    // Act
    _, err := collector.CollectRecent(-1)
    
    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid limit")
}
```

### Mock Testing

```go
// internal/api/mock_usgs.go
package api

import (
    "github.com/stretchr/testify/mock"
)

type MockUSGSClient struct {
    mock.Mock
}

func (m *MockUSGSClient) GetEarthquakes(params EarthquakeParams) (*USGSResponse, error) {
    args := m.Called(params)
    return args.Get(0).(*USGSResponse), args.Error(1)
}

// Test usage
func TestCollector_WithMockAPI(t *testing.T) {
    mockAPI := &MockUSGSClient{}
    mockAPI.On("GetEarthquakes", mock.Anything).Return(&USGSResponse{
        Features: []Earthquake{},
    }, nil)
    
    collector := NewEarthquakeCollector(mockAPI)
    // ... test implementation
}
```

### Integration Tests

```go
// tests/integration/collector_test.go
package integration

import (
    "testing"
    "time"
    
    "quakewatch-scraper/internal/collector"
    "quakewatch-scraper/internal/storage"
)

func TestEarthquakeCollection_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Set up test database
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    // Create collector with real dependencies
    collector := collector.NewEarthquakeCollector()
    storage := storage.NewPostgreSQLStorage(db)
    
    // Test complete workflow
    earthquakes, err := collector.CollectRecent(5)
    require.NoError(t, err)
    
    err = storage.StoreEarthquakes(earthquakes)
    require.NoError(t, err)
    
    // Verify data was stored
    stored, err := storage.GetEarthquakes(10)
    require.NoError(t, err)
    assert.Len(t, stored, 5)
}
```

### Performance Testing

```go
// internal/collector/earthquake_bench_test.go
package collector

import (
    "testing"
)

func BenchmarkEarthquakeCollector_CollectRecent(b *testing.B) {
    collector := NewEarthquakeCollector()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := collector.CollectRecent(100)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkEarthquakeCollector_ProcessData(b *testing.B) {
    collector := NewEarthquakeCollector()
    earthquakes := generateTestEarthquakes(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        collector.ProcessData(earthquakes)
    }
}
```

## Database Development

### Migration Development

1. **Create New Migration**

```bash
# Create new migration file
migrate create -ext sql -dir migrations -seq add_new_table
```

2. **Migration File Structure**

```sql
-- migrations/000002_add_new_table.up.sql
CREATE TABLE new_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_new_table_name ON new_table(name);

-- migrations/000002_add_new_table.down.sql
DROP INDEX IF EXISTS idx_new_table_name;
DROP TABLE IF EXISTS new_table;
```

3. **Test Migrations**

```bash
# Test migration up
make db-migrate-up-test

# Test migration down
make db-migrate-down-test

# Verify migration
make db-status-test
```

### Database Testing

```go
// internal/storage/postgresql_test.go
package storage

import (
    "testing"
    "database/sql"
    
    _ "github.com/lib/pq"
    "github.com/stretchr/testify/require"
)

func setupTestDatabase(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/quakewatch_test?sslmode=disable")
    require.NoError(t, err)
    
    // Run migrations
    err = runMigrations(db, "up")
    require.NoError(t, err)
    
    return db
}

func cleanupTestDatabase(t *testing.T, db *sql.DB) {
    // Run down migrations
    err := runMigrations(db, "down")
    require.NoError(t, err)
    
    db.Close()
}

func TestPostgreSQLStorage_StoreEarthquakes(t *testing.T) {
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    storage := NewPostgreSQLStorage(db)
    
    earthquakes := []Earthquake{
        // Test data
    }
    
    err := storage.StoreEarthquakes(earthquakes)
    require.NoError(t, err)
    
    // Verify data was stored
    stored, err := storage.GetEarthquakes(10)
    require.NoError(t, err)
    require.Len(t, stored, len(earthquakes))
}
```

## API Development

### Adding New API Endpoints

1. **Define API Client Interface**

```go
// internal/api/interface.go
type EarthquakeAPI interface {
    GetRecent(limit int) ([]Earthquake, error)
    GetByTimeRange(start, end time.Time) ([]Earthquake, error)
    GetByMagnitude(min, max float64) ([]Earthquake, error)
}
```

2. **Implement API Client**

```go
// internal/api/usgs.go
type USGSClient struct {
    baseURL    string
    httpClient *http.Client
    rateLimiter *rate.Limiter
}

func (c *USGSClient) GetRecent(limit int) ([]Earthquake, error) {
    // Implementation
}

func (c *USGSClient) GetByTimeRange(start, end time.Time) ([]Earthquake, error) {
    // Implementation
}
```

3. **Add Tests**

```go
// internal/api/usgs_test.go
func TestUSGSClient_GetRecent(t *testing.T) {
    client := NewUSGSClient("https://earthquake.usgs.gov/fdsnws/event/1")
    
    earthquakes, err := client.GetRecent(10)
    require.NoError(t, err)
    assert.Len(t, earthquakes, 10)
}
```

### Error Handling

```go
// internal/api/errors.go
type APIError struct {
    StatusCode int
    Message    string
    URL        string
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error %d: %s (URL: %s)", e.StatusCode, e.Message, e.URL)
}

func (e *APIError) IsRateLimit() bool {
    return e.StatusCode == 429
}

func (e *APIError) IsNotFound() bool {
    return e.StatusCode == 404
}
```

## CLI Development

### Adding New Commands

1. **Define Command Structure**

```go
// pkg/cli/commands.go
func (a *App) newCustomCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "custom",
        Short: "Custom command description",
        Long:  `Detailed description of the custom command`,
        RunE:  a.runCustom,
    }
    
    // Add flags
    cmd.Flags().String("param", "", "Parameter description")
    cmd.Flags().Int("limit", 100, "Limit description")
    
    return cmd
}
```

2. **Implement Command Logic**

```go
func (a *App) runCustom(cmd *cobra.Command, args []string) error {
    // Get flags
    param, _ := cmd.Flags().GetString("param")
    limit, _ := cmd.Flags().GetInt("limit")
    
    // Validate input
    if param == "" {
        return fmt.Errorf("param is required")
    }
    
    // Execute command logic
    result, err := a.executeCustomLogic(param, limit)
    if err != nil {
        return fmt.Errorf("failed to execute custom logic: %w", err)
    }
    
    // Output result
    return a.outputToStdout(result)
}
```

3. **Add Command Tests**

```go
// pkg/cli/commands_test.go
func TestCustomCommand(t *testing.T) {
    app := NewApp()
    
    // Test command execution
    cmd := app.newCustomCmd()
    cmd.SetArgs([]string{"--param", "test", "--limit", "5"})
    
    err := cmd.Execute()
    assert.NoError(t, err)
}

func TestCustomCommand_Validation(t *testing.T) {
    app := NewApp()
    
    cmd := app.newCustomCmd()
    cmd.SetArgs([]string{"--limit", "5"}) // Missing required param
    
    err := cmd.Execute()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "param is required")
}
```

### Command Output Formatting

```go
// pkg/cli/output.go
type OutputFormatter interface {
    Format(data interface{}) ([]byte, error)
}

type JSONFormatter struct{}

func (f *JSONFormatter) Format(data interface{}) ([]byte, error) {
    return json.MarshalIndent(data, "", "  ")
}

type CSVFormatter struct{}

func (f *CSVFormatter) Format(data interface{}) ([]byte, error) {
    // CSV formatting implementation
}

func (a *App) outputToStdout(data interface{}) error {
    var formatter OutputFormatter
    
    switch a.cfg.Output.Format {
    case "json":
        formatter = &JSONFormatter{}
    case "csv":
        formatter = &CSVFormatter{}
    default:
        formatter = &JSONFormatter{}
    }
    
    output, err := formatter.Format(data)
    if err != nil {
        return fmt.Errorf("failed to format output: %w", err)
    }
    
    fmt.Println(string(output))
    return nil
}
```

## Contributing

### Development Workflow

1. **Fork and Clone**

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/your-username/quakewatch-scraper.git
cd quakewatch-scraper

# Add upstream remote
git remote add upstream https://github.com/original-org/quakewatch-scraper.git
```

2. **Create Feature Branch**

```bash
# Create and switch to feature branch
git checkout -b feature/your-feature-name

# Or use conventional branch naming
git checkout -b feat/add-new-api-endpoint
git checkout -b fix/database-connection-issue
git checkout -b docs/update-api-documentation
```

3. **Make Changes**

```bash
# Make your changes
# Follow coding standards
# Add tests for new functionality
# Update documentation

# Stage changes
git add .

# Commit with conventional commit message
git commit -m "feat: add new earthquake filtering by depth

- Add depth-based filtering to earthquake collector
- Implement depth range validation
- Add tests for depth filtering functionality
- Update CLI documentation for new feature"
```

4. **Test Changes**

```bash
# Run all tests
make test

# Run linting
make lint

# Run integration tests
make test-integration

# Build application
make build
```

5. **Push and Create Pull Request**

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
# Include:
# - Description of changes
# - Testing performed
# - Screenshots (if applicable)
# - Related issues
```

### Commit Message Convention

Use conventional commit messages:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(collector): add magnitude range filtering

fix(database): resolve connection timeout issue

docs(api): update USGS API documentation

test(storage): add integration tests for PostgreSQL
```

### Code Review Process

1. **Self-Review**: Review your own code before submitting
2. **Peer Review**: At least one other developer must review
3. **Automated Checks**: All CI checks must pass
4. **Documentation**: Update relevant documentation
5. **Testing**: Ensure adequate test coverage

### Review Checklist

- [ ] Code follows project standards
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No breaking changes (or documented)
- [ ] Performance impact considered
- [ ] Security implications reviewed
- [ ] Error handling is appropriate
- [ ] Logging is adequate

## Release Process

### Version Management

1. **Update Version**

```bash
# Update version in code
# internal/config/version.go
const Version = "1.1.0"

# Update go.mod if needed
go mod edit -go=1.24
```

2. **Create Release Branch**

```bash
git checkout -b release/v1.1.0
git push origin release/v1.1.0
```

3. **Update Documentation**

```bash
# Update CHANGELOG.md
# Update README.md if needed
# Update API documentation
```

4. **Create Release**

```bash
# Tag the release
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# Create GitHub release
# Include:
# - Release notes
# - Binary downloads
# - Docker images
```

### Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated
- [ ] Binary builds successful
- [ ] Docker images built
- [ ] Release notes prepared
- [ ] GitHub release created
- [ ] Announcement sent

### Hotfix Process

1. **Create Hotfix Branch**

```bash
git checkout -b hotfix/critical-bug-fix
```

2. **Fix and Test**

```bash
# Make minimal changes to fix the issue
# Add tests to prevent regression
# Test thoroughly
```

3. **Release Hotfix**

```bash
# Update version (patch increment)
# Create hotfix release
# Merge to main and develop branches
```

This comprehensive development guide provides all the information needed for developers to contribute effectively to the QuakeWatch Scraper project, from initial setup to release management. 
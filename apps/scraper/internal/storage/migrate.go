package storage

import (
	"fmt"
	"log"

	"quakewatch-scraper/internal/config"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db     *sqlx.DB
	config *config.DatabaseConfig
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(config *config.DatabaseConfig) (*MigrationManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	db, err := sqlx.Connect("postgres", config.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &MigrationManager{
		db:     db,
		config: config,
	}, nil
}

// MigrateUp runs all pending migrations
func (m *MigrationManager) MigrateUp() error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// MigrateDown rolls back all migrations
func (m *MigrationManager) MigrateDown() error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	log.Println("Database migrations rolled back successfully")
	return nil
}

// MigrateToVersion migrates to a specific version
func (m *MigrationManager) MigrateToVersion(version uint) error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Migrate(version); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	log.Printf("Database migrated to version %d successfully", version)
	return nil
}

// GetVersion returns the current migration version
func (m *MigrationManager) GetVersion() (uint, bool, error) {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

// GetVersionWithoutClose returns the current migration version without closing the connection
func (m *MigrationManager) GetVersionWithoutClose() (uint, bool, error) {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	// Note: We don't defer migrator.Close() here to avoid closing the connection

	version, dirty, err := migrator.Version()
	if err != nil {
		migrator.Close() // Close on error
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	migrator.Close() // Close after successful operation
	return version, dirty, nil
}

// ForceVersion forces the migration version
func (m *MigrationManager) ForceVersion(version uint) error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Force(int(version)); err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}

	log.Printf("Database version forced to %d", version)
	return nil
}

// TestConnection tests the database connection
func (m *MigrationManager) TestConnection() error {
	return m.db.Ping()
}

// TableExists checks if a table exists in the database
func (m *MigrationManager) TableExists(tableName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = $1
	)`
	err := m.db.Get(&exists, query, tableName)
	return exists, err
}

// Close closes the database connection
func (m *MigrationManager) Close() error {
	return m.db.Close()
}

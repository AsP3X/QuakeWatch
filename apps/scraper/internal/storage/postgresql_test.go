package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"quakewatch-scraper/internal/config"
	"quakewatch-scraper/internal/models"
)

func TestPostgreSQLStorage_Integration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run")
	}

	// Create test configuration
	config := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Database: "quakewatch_test",
		SSLMode:  "disable",
	}

	// Create storage instance
	storage, err := NewPostgreSQLStorage(config)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL storage: %v", err)
	}
	defer storage.Close()

	// Test earthquake operations
	t.Run("EarthquakeOperations", func(t *testing.T) {
		testEarthquakeOperations(t, storage)
	})

	// Test fault operations
	t.Run("FaultOperations", func(t *testing.T) {
		testFaultOperations(t, storage)
	})

	// Test statistics
	t.Run("Statistics", func(t *testing.T) {
		testStatistics(t, storage)
	})
}

func testEarthquakeOperations(t *testing.T, storage *PostgreSQLStorage) {
	ctx := context.Background()

	// Create test earthquake data
	earthquakes := &models.USGSResponse{
		Type: "FeatureCollection",
		Features: []models.Earthquake{
			{
				Type: "Feature",
				ID:   "test-earthquake-1",
				Properties: models.EarthquakeProperties{
					Mag:     5.5,
					Place:   "Test Location",
					Time:    time.Now().UnixMilli(),
					Updated: time.Now().UnixMilli(),
					Status:  "reviewed",
					Sig:     100,
					Net:     "us",
					Code:    "test123",
					Title:   "Test Earthquake",
				},
				Geometry: models.Geometry{
					Type:        "Point",
					Coordinates: []float64{-122.4194, 37.7749, 10.0},
				},
			},
		},
	}

	// Test saving earthquakes
	err := storage.SaveEarthquakes(ctx, earthquakes)
	if err != nil {
		t.Fatalf("Failed to save earthquakes: %v", err)
	}

	// Test loading earthquakes
	loaded, err := storage.LoadEarthquakes(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to load earthquakes: %v", err)
	}

	if len(loaded.Features) == 0 {
		t.Error("Expected to load at least one earthquake")
	}

	// Verify the loaded earthquake
	found := false
	for _, eq := range loaded.Features {
		if eq.ID == "test-earthquake-1" {
			found = true
			if eq.Properties.Mag != 5.5 {
				t.Errorf("Expected magnitude 5.5, got %f", eq.Properties.Mag)
			}
			break
		}
	}

	if !found {
		t.Error("Test earthquake not found in loaded data")
	}
}

func testFaultOperations(t *testing.T, storage *PostgreSQLStorage) {
	ctx := context.Background()

	// Create test fault data
	faults := &models.Fault{
		Type: "FeatureCollection",
		Features: []models.FaultFeature{
			{
				Type: "Feature",
				ID:   "test-fault-1",
				Properties: models.FaultProperties{
					ID:   "test-fault-1",
					Name: "Test Fault",
					Type: "strike-slip",
				},
				Geometry: models.FaultGeometry{
					Type: "LineString",
					Coordinates: [][]float64{
						{-122.4194, 37.7749},
						{-122.4195, 37.7750},
					},
				},
			},
		},
	}

	// Test saving faults
	err := storage.SaveFaults(ctx, faults)
	if err != nil {
		t.Fatalf("Failed to save faults: %v", err)
	}

	// Test loading faults
	loaded, err := storage.LoadFaults(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to load faults: %v", err)
	}

	if len(loaded.Features) == 0 {
		t.Error("Expected to load at least one fault")
	}

	// Verify the loaded fault
	found := false
	for _, fault := range loaded.Features {
		if fault.ID == "test-fault-1" {
			found = true
			if fault.Properties.Name != "Test Fault" {
				t.Errorf("Expected name 'Test Fault', got %s", fault.Properties.Name)
			}
			break
		}
	}

	if !found {
		t.Error("Test fault not found in loaded data")
	}
}

func testStatistics(t *testing.T, storage *PostgreSQLStorage) {
	ctx := context.Background()

	// Test getting statistics
	stats, err := storage.GetStatistics(ctx)
	if err != nil {
		t.Fatalf("Failed to get statistics: %v", err)
	}

	// Basic validation
	if stats.TotalEarthquakes < 0 {
		t.Error("Total earthquakes should be non-negative")
	}

	if stats.TotalFaults < 0 {
		t.Error("Total faults should be non-negative")
	}
}

func TestDatabaseConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DatabaseConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Database: "testdb",
			},
			wantErr: false,
		},
		{
			name: "Missing host",
			config: &config.DatabaseConfig{
				Port:     5432,
				User:     "testuser",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "Missing user",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "Missing database",
			config: &config.DatabaseConfig{
				Host: "localhost",
				Port: 5432,
				User: "testuser",
			},
			wantErr: true,
		},
		{
			name: "Invalid port",
			config: &config.DatabaseConfig{
				Host:     "localhost",
				Port:     70000, // Invalid port
				User:     "testuser",
				Database: "testdb",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDatabaseConfig_ConnectionStrings(t *testing.T) {
	config := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	// Test connection string
	connStr := config.GetConnectionString()
	expectedConnStr := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	if connStr != expectedConnStr {
		t.Errorf("GetConnectionString() = %v, want %v", connStr, expectedConnStr)
	}

	// Test DSN
	dsn := config.GetDSN()
	expectedDSN := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	if dsn != expectedDSN {
		t.Errorf("GetDSN() = %v, want %v", dsn, expectedDSN)
	}
}

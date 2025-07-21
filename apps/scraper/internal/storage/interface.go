package storage

import (
	"context"
	"quakewatch-scraper/internal/models"
)

// Storage defines the interface for data storage operations
type Storage interface {
	// Earthquake operations
	SaveEarthquakes(ctx context.Context, earthquakes *models.USGSResponse) error
	LoadEarthquakes(ctx context.Context, limit int, offset int) (*models.USGSResponse, error)
	GetEarthquakeByID(ctx context.Context, usgsID string) (*models.Earthquake, error)
	GetEarthquakesByTimeRange(ctx context.Context, startTime, endTime int64) ([]models.Earthquake, error)
	GetEarthquakesByMagnitudeRange(ctx context.Context, minMag, maxMag float64) ([]models.Earthquake, error)
	GetEarthquakesByLocation(ctx context.Context, minLat, maxLat, minLon, maxLon float64) ([]models.Earthquake, error)
	GetSignificantEarthquakes(ctx context.Context, startTime, endTime int64) ([]models.Earthquake, error)
	DeleteEarthquake(ctx context.Context, usgsID string) error

	// Fault operations
	SaveFaults(ctx context.Context, faults *models.Fault) error
	LoadFaults(ctx context.Context, limit int, offset int) (*models.Fault, error)
	GetFaultByID(ctx context.Context, faultID string) (*models.FaultFeature, error)
	GetFaultsByType(ctx context.Context, faultType string) ([]models.FaultFeature, error)
	GetFaultsByLocation(ctx context.Context, minLat, maxLat, minLon, maxLon float64) ([]models.FaultFeature, error)
	DeleteFault(ctx context.Context, faultID string) error

	// Collection logging
	LogCollection(ctx context.Context, dataType, source string, startTime int64, recordsCollected int, status string, errorMsg string) error
	GetCollectionLogs(ctx context.Context, dataType string, limit int) ([]CollectionLog, error)

	// Statistics and metadata
	GetStatistics(ctx context.Context) (*Statistics, error)
	GetFileStats(ctx context.Context, dataType string) (map[string]interface{}, error)

	// Maintenance operations
	PurgeAll(ctx context.Context) error
	PurgeByType(ctx context.Context, dataType string) error
	Close() error
}

// CollectionLog represents a data collection operation log
type CollectionLog struct {
	ID               int64  `db:"id"`
	DataType         string `db:"data_type"`
	Source           string `db:"source"`
	StartTime        int64  `db:"start_time"`
	EndTime          *int64 `db:"end_time"`
	RecordsCollected int    `db:"records_collected"`
	Status           string `db:"status"`
	ErrorMessage     string `db:"error_message"`
	CreatedAt        int64  `db:"created_at"`
}

// Statistics represents database statistics
type Statistics struct {
	TotalEarthquakes       int64  `db:"total_earthquakes"`
	TotalFaults            int64  `db:"total_faults"`
	RecentEarthquakes      int64  `db:"recent_earthquakes"`
	SignificantEarthquakes int64  `db:"significant_earthquakes"`
	LastCollection         *int64 `db:"last_collection"`
}

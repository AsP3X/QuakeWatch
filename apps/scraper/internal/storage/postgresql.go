package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"quakewatch-scraper/internal/config"
	"quakewatch-scraper/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgreSQLStorage implements the Storage interface for PostgreSQL
type PostgreSQLStorage struct {
	db     *sqlx.DB
	config *config.DatabaseConfig
}

// NewPostgreSQLStorage creates a new PostgreSQL storage instance
func NewPostgreSQLStorage(config *config.DatabaseConfig) (*PostgreSQLStorage, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	db, err := sqlx.Connect("postgres", config.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgreSQLStorage{
		db:     db,
		config: config,
	}, nil
}

// SaveEarthquakes saves earthquake data to the database
func (s *PostgreSQLStorage) SaveEarthquakes(ctx context.Context, earthquakes *models.USGSResponse) error {
	if earthquakes == nil || len(earthquakes.Features) == 0 {
		return nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO earthquakes (
			usgs_id, magnitude, magnitude_type, place, time, updated, url, detail_url,
			felt_count, cdi, mmi, alert, status, tsunami, significance, network, code,
			ids, sources, types, nst, dmin, rms, gap, latitude, longitude, depth, title
		) VALUES (
			:usgs_id, :magnitude, :magnitude_type, :place, :time, :updated, :url, :detail_url,
			:felt_count, :cdi, :mmi, :alert, :status, :tsunami, :significance, :network, :code,
			:ids, :sources, :types, :nst, :dmin, :rms, :gap, :latitude, :longitude, :depth, :title
		) ON CONFLICT (usgs_id) DO UPDATE SET
			magnitude = EXCLUDED.magnitude,
			magnitude_type = EXCLUDED.magnitude_type,
			place = EXCLUDED.place,
			updated = EXCLUDED.updated,
			url = EXCLUDED.url,
			detail_url = EXCLUDED.detail_url,
			felt_count = EXCLUDED.felt_count,
			cdi = EXCLUDED.cdi,
			mmi = EXCLUDED.mmi,
			alert = EXCLUDED.alert,
			status = EXCLUDED.status,
			tsunami = EXCLUDED.tsunami,
			significance = EXCLUDED.significance,
			network = EXCLUDED.network,
			code = EXCLUDED.code,
			ids = EXCLUDED.ids,
			sources = EXCLUDED.sources,
			types = EXCLUDED.types,
			nst = EXCLUDED.nst,
			dmin = EXCLUDED.dmin,
			rms = EXCLUDED.rms,
			gap = EXCLUDED.gap,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			depth = EXCLUDED.depth,
			title = EXCLUDED.title,
			updated_at = NOW()
	`

	for _, earthquake := range earthquakes.Features {
		// Extract coordinates
		var latitude, longitude, depth float64
		if len(earthquake.Geometry.Coordinates) >= 3 {
			longitude = earthquake.Geometry.Coordinates[0]
			latitude = earthquake.Geometry.Coordinates[1]
			depth = earthquake.Geometry.Coordinates[2]
		} else if len(earthquake.Geometry.Coordinates) >= 2 {
			longitude = earthquake.Geometry.Coordinates[0]
			latitude = earthquake.Geometry.Coordinates[1]
		}

		// Convert tsunami int to boolean
		tsunami := earthquake.Properties.Tsunami > 0

		params := map[string]interface{}{
			"usgs_id":        earthquake.ID,
			"magnitude":      earthquake.Properties.Mag,
			"magnitude_type": earthquake.Properties.MagType,
			"place":          earthquake.Properties.Place,
			"time":           earthquake.Properties.GetTime(),
			"updated":        earthquake.Properties.GetUpdated(),
			"url":            earthquake.Properties.URL,
			"detail_url":     earthquake.Properties.Detail,
			"felt_count":     earthquake.Properties.Felt,
			"cdi":            earthquake.Properties.CDI,
			"mmi":            earthquake.Properties.MMI,
			"alert":          earthquake.Properties.Alert,
			"status":         earthquake.Properties.Status,
			"tsunami":        tsunami,
			"significance":   earthquake.Properties.Sig,
			"network":        earthquake.Properties.Net,
			"code":           earthquake.Properties.Code,
			"ids":            earthquake.Properties.IDs,
			"sources":        earthquake.Properties.Sources,
			"types":          earthquake.Properties.Types,
			"nst":            earthquake.Properties.Nst,
			"dmin":           earthquake.Properties.Dmin,
			"rms":            earthquake.Properties.RMS,
			"gap":            earthquake.Properties.Gap,
			"latitude":       latitude,
			"longitude":      longitude,
			"depth":          depth,
			"title":          earthquake.Properties.Title,
		}

		_, err := tx.NamedExecContext(ctx, query, params)
		if err != nil {
			return fmt.Errorf("failed to insert earthquake %s: %w", earthquake.ID, err)
		}
	}

	return tx.Commit()
}

// LoadEarthquakes loads earthquakes from the database
func (s *PostgreSQLStorage) LoadEarthquakes(ctx context.Context, limit int, offset int) (*models.USGSResponse, error) {
	query := `
		SELECT 
			id, usgs_id, magnitude, magnitude_type, place, time, updated, url, detail_url,
			felt_count, cdi, mmi, alert, status, tsunami, significance, network, code,
			ids, sources, types, nst, dmin, rms, gap, latitude, longitude, depth, title
		FROM earthquakes 
		ORDER BY time DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query earthquakes: %w", err)
	}
	defer rows.Close()

	var earthquakes []models.Earthquake
	for rows.Next() {
		var eq struct {
			ID            int       `db:"id"`
			USGSID        string    `db:"usgs_id"`
			Magnitude     float64   `db:"magnitude"`
			MagnitudeType string    `db:"magnitude_type"`
			Place         string    `db:"place"`
			Time          time.Time `db:"time"`
			Updated       time.Time `db:"updated"`
			URL           string    `db:"url"`
			DetailURL     string    `db:"detail_url"`
			FeltCount     *int      `db:"felt_count"`
			CDI           *float64  `db:"cdi"`
			MMI           *float64  `db:"mmi"`
			Alert         string    `db:"alert"`
			Status        string    `db:"status"`
			Tsunami       bool      `db:"tsunami"`
			Significance  int       `db:"significance"`
			Network       string    `db:"network"`
			Code          string    `db:"code"`
			IDs           string    `db:"ids"`
			Sources       string    `db:"sources"`
			Types         string    `db:"types"`
			Nst           *int      `db:"nst"`
			Dmin          *float64  `db:"dmin"`
			RMS           *float64  `db:"rms"`
			Gap           *float64  `db:"gap"`
			Latitude      float64   `db:"latitude"`
			Longitude     float64   `db:"longitude"`
			Depth         *float64  `db:"depth"`
			Title         string    `db:"title"`
		}

		if err := rows.StructScan(&eq); err != nil {
			return nil, fmt.Errorf("failed to scan earthquake: %w", err)
		}

		// Convert tsunami boolean to int
		tsunami := 0
		if eq.Tsunami {
			tsunami = 1
		}

		earthquake := models.Earthquake{
			Type: "Feature",
			ID:   eq.USGSID,
			Properties: models.EarthquakeProperties{
				Mag:     eq.Magnitude,
				Place:   eq.Place,
				Time:    eq.Time.UnixMilli(),
				Updated: eq.Updated.UnixMilli(),
				URL:     eq.URL,
				Detail:  eq.DetailURL,
				Felt:    eq.FeltCount,
				CDI:     eq.CDI,
				MMI:     eq.MMI,
				Alert:   eq.Alert,
				Status:  eq.Status,
				Tsunami: tsunami,
				Sig:     eq.Significance,
				Net:     eq.Network,
				Code:    eq.Code,
				IDs:     eq.IDs,
				Sources: eq.Sources,
				Types:   eq.Types,
				Nst:     eq.Nst,
				Dmin:    eq.Dmin,
				RMS:     eq.RMS,
				Gap:     eq.Gap,
				MagType: eq.MagnitudeType,
				Type:    "earthquake",
				Title:   eq.Title,
			},
			Geometry: models.Geometry{
				Type:        "Point",
				Coordinates: []float64{eq.Longitude, eq.Latitude},
			},
		}

		// Add depth if available
		if eq.Depth != nil {
			earthquake.Geometry.Coordinates = append(earthquake.Geometry.Coordinates, *eq.Depth)
		}

		earthquakes = append(earthquakes, earthquake)
	}

	return &models.USGSResponse{
		Type:     "FeatureCollection",
		Features: earthquakes,
	}, nil
}

// SaveFaults saves fault data to the database
func (s *PostgreSQLStorage) SaveFaults(ctx context.Context, faults *models.Fault) error {
	if faults == nil || len(faults.Features) == 0 {
		return nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO faults (
			fault_id, name, fault_type, slip_rate, slip_type, dip, rake, length, width,
			max_magnitude, description, source, geometry_type, coordinates
		) VALUES (
			:fault_id, :name, :fault_type, :slip_rate, :slip_type, :dip, :rake, :length, :width,
			:max_magnitude, :description, :source, :geometry_type, :coordinates
		) ON CONFLICT (fault_id) DO UPDATE SET
			name = EXCLUDED.name,
			fault_type = EXCLUDED.fault_type,
			slip_rate = EXCLUDED.slip_rate,
			slip_type = EXCLUDED.slip_type,
			dip = EXCLUDED.dip,
			rake = EXCLUDED.rake,
			length = EXCLUDED.length,
			width = EXCLUDED.width,
			max_magnitude = EXCLUDED.max_magnitude,
			description = EXCLUDED.description,
			source = EXCLUDED.source,
			geometry_type = EXCLUDED.geometry_type,
			coordinates = EXCLUDED.coordinates,
			updated_at = NOW()
	`

	for _, fault := range faults.Features {
		coordinates, err := json.Marshal(fault.Geometry.Coordinates)
		if err != nil {
			return fmt.Errorf("failed to marshal coordinates for fault %s: %w", fault.Properties.ID, err)
		}

		params := map[string]interface{}{
			"fault_id":      fault.Properties.ID,
			"name":          fault.Properties.Name,
			"fault_type":    fault.Properties.Type,
			"slip_rate":     fault.Properties.SlipRate,
			"slip_type":     fault.Properties.SlipType,
			"dip":           fault.Properties.Dip,
			"rake":          fault.Properties.Rake,
			"length":        fault.Properties.Length,
			"width":         fault.Properties.Width,
			"max_magnitude": fault.Properties.MaxMagnitude,
			"description":   fault.Properties.Description,
			"source":        fault.Properties.Source,
			"geometry_type": fault.Geometry.Type,
			"coordinates":   coordinates,
		}

		_, err = tx.NamedExecContext(ctx, query, params)
		if err != nil {
			return fmt.Errorf("failed to insert fault %s: %w", fault.Properties.ID, err)
		}
	}

	return tx.Commit()
}

// LoadFaults loads faults from the database
func (s *PostgreSQLStorage) LoadFaults(ctx context.Context, limit int, offset int) (*models.Fault, error) {
	query := `
		SELECT 
			id, fault_id, name, fault_type, slip_rate, slip_type, dip, rake, length, width,
			max_magnitude, description, source, geometry_type, coordinates
		FROM faults 
		ORDER BY name 
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query faults: %w", err)
	}
	defer rows.Close()

	var faultFeatures []models.FaultFeature
	for rows.Next() {
		var f struct {
			ID           int             `db:"id"`
			FaultID      string          `db:"fault_id"`
			Name         string          `db:"name"`
			FaultType    string          `db:"fault_type"`
			SlipRate     *float64        `db:"slip_rate"`
			SlipType     string          `db:"slip_type"`
			Dip          *float64        `db:"dip"`
			Rake         *float64        `db:"rake"`
			Length       *float64        `db:"length"`
			Width        *float64        `db:"width"`
			MaxMagnitude *float64        `db:"max_magnitude"`
			Description  string          `db:"description"`
			Source       string          `db:"source"`
			GeometryType string          `db:"geometry_type"`
			Coordinates  json.RawMessage `db:"coordinates"`
		}

		if err := rows.StructScan(&f); err != nil {
			return nil, fmt.Errorf("failed to scan fault: %w", err)
		}

		var coordinates [][]float64
		if err := json.Unmarshal(f.Coordinates, &coordinates); err != nil {
			return nil, fmt.Errorf("failed to unmarshal coordinates for fault %s: %w", f.FaultID, err)
		}

		faultFeature := models.FaultFeature{
			Type: "Feature",
			ID:   f.FaultID,
			Properties: models.FaultProperties{
				ID:           f.FaultID,
				Name:         f.Name,
				Type:         f.FaultType,
				SlipRate:     f.SlipRate,
				SlipType:     f.SlipType,
				Dip:          f.Dip,
				Rake:         f.Rake,
				Length:       f.Length,
				Width:        f.Width,
				MaxMagnitude: f.MaxMagnitude,
				Description:  f.Description,
				Source:       f.Source,
			},
			Geometry: models.FaultGeometry{
				Type:        f.GeometryType,
				Coordinates: coordinates,
			},
		}

		faultFeatures = append(faultFeatures, faultFeature)
	}

	return &models.Fault{
		Type:     "FeatureCollection",
		Features: faultFeatures,
	}, nil
}

// LogCollection logs a data collection operation
func (s *PostgreSQLStorage) LogCollection(ctx context.Context, dataType, source string, startTime int64, recordsCollected int, status string, errorMsg string) error {
	query := `
		INSERT INTO collection_logs (
			data_type, source, start_time, end_time, records_collected, status, error_message
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	startTimeObj := time.Unix(startTime/1000, 0)
	var endTimeObj *time.Time
	if status == "completed" || status == "failed" {
		now := time.Now()
		endTimeObj = &now
	}

	_, err := s.db.ExecContext(ctx, query, dataType, source, startTimeObj, endTimeObj, recordsCollected, status, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to log collection: %w", err)
	}

	return nil
}

// GetStatistics returns database statistics
func (s *PostgreSQLStorage) GetStatistics(ctx context.Context) (*Statistics, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM earthquakes) as total_earthquakes,
			(SELECT COUNT(*) FROM faults) as total_faults,
			(SELECT COUNT(*) FROM earthquakes WHERE time > NOW() - INTERVAL '24 hours') as recent_earthquakes,
			(SELECT COUNT(*) FROM earthquakes WHERE magnitude >= 4.5) as significant_earthquakes,
			(SELECT EXTRACT(EPOCH FROM MAX(created_at)) FROM collection_logs) as last_collection
	`

	var stats Statistics
	err := s.db.GetContext(ctx, &stats, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	return &stats, nil
}

// Close closes the database connection
func (s *PostgreSQLStorage) Close() error {
	return s.db.Close()
}

// Implement remaining interface methods with placeholder implementations
func (s *PostgreSQLStorage) GetEarthquakeByID(ctx context.Context, usgsID string) (*models.Earthquake, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetEarthquakesByTimeRange(ctx context.Context, startTime, endTime int64) ([]models.Earthquake, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetEarthquakesByMagnitudeRange(ctx context.Context, minMag, maxMag float64) ([]models.Earthquake, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetEarthquakesByLocation(ctx context.Context, minLat, maxLat, minLon, maxLon float64) ([]models.Earthquake, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetSignificantEarthquakes(ctx context.Context, startTime, endTime int64) ([]models.Earthquake, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) DeleteEarthquake(ctx context.Context, usgsID string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetFaultByID(ctx context.Context, faultID string) (*models.FaultFeature, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetFaultsByType(ctx context.Context, faultType string) ([]models.FaultFeature, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetFaultsByLocation(ctx context.Context, minLat, maxLat, minLon, maxLon float64) ([]models.FaultFeature, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) DeleteFault(ctx context.Context, faultID string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetCollectionLogs(ctx context.Context, dataType string, limit int) ([]CollectionLog, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) GetFileStats(ctx context.Context, dataType string) (map[string]interface{}, error) {
	// TODO: Implement
	return nil, fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) PurgeAll(ctx context.Context) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

func (s *PostgreSQLStorage) PurgeByType(ctx context.Context, dataType string) error {
	// TODO: Implement
	return fmt.Errorf("not implemented")
}

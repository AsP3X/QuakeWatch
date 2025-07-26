package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"quakewatch-scraper/internal/models"
)

// JSONStorage implements the Storage interface for JSON file storage
type JSONStorage struct {
	outputDir string
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(outputDir string) *JSONStorage {
	return &JSONStorage{
		outputDir: outputDir,
	}
}

// SaveEarthquakes saves earthquake data to a JSON file
func (s *JSONStorage) SaveEarthquakes(ctx context.Context, earthquakes *models.USGSResponse) error {
	if earthquakes == nil || len(earthquakes.Features) == 0 {
		return nil
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("earthquakes_%s.json", timestamp)
	filePath := filepath.Join(s.outputDir, "earthquakes", filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(earthquakes); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Print the actual file path that was saved
	fmt.Printf("Saved earthquakes to %s\n", filePath)
	return nil
}

// SaveFaults saves fault data to a JSON file
func (s *JSONStorage) SaveFaults(ctx context.Context, faults *models.Fault) error {
	if faults == nil || len(faults.Features) == 0 {
		return nil
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("faults_%s.json", timestamp)
	filePath := filepath.Join(s.outputDir, "faults", filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(faults); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// ListFiles returns a list of JSON files for a specific data type
func (s *JSONStorage) ListFiles(dataType string) ([]string, error) {
	var dir string
	switch dataType {
	case "earthquakes":
		dir = filepath.Join(s.outputDir, "earthquakes")
	case "faults":
		dir = filepath.Join(s.outputDir, "faults")
	default:
		return nil, fmt.Errorf("unknown data type: %s", dataType)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var filenames []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filenames = append(filenames, file.Name())
		}
	}

	return filenames, nil
}

// LoadEarthquakes loads earthquake data from a JSON file
func (s *JSONStorage) LoadEarthquakes(ctx context.Context, limit int, offset int) (*models.USGSResponse, error) {
	files, err := s.ListFiles("earthquakes")
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return &models.USGSResponse{
			Type:     "FeatureCollection",
			Features: []models.Earthquake{},
		}, nil
	}

	// Load the most recent file
	latestFile := files[len(files)-1]
	filePath := filepath.Join(s.outputDir, "earthquakes", latestFile)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var earthquakes models.USGSResponse
	if err := json.NewDecoder(file).Decode(&earthquakes); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Apply limit and offset
	start := offset
	end := start + limit
	if start >= len(earthquakes.Features) {
		start = len(earthquakes.Features)
	}
	if end > len(earthquakes.Features) {
		end = len(earthquakes.Features)
	}

	earthquakes.Features = earthquakes.Features[start:end]
	return &earthquakes, nil
}

// LoadFaults loads fault data from a JSON file
func (s *JSONStorage) LoadFaults(ctx context.Context, limit int, offset int) (*models.Fault, error) {
	files, err := s.ListFiles("faults")
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return &models.Fault{
			Type:     "FeatureCollection",
			Features: []models.FaultFeature{},
		}, nil
	}

	// Load the most recent file
	latestFile := files[len(files)-1]
	filePath := filepath.Join(s.outputDir, "faults", latestFile)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var faults models.Fault
	if err := json.NewDecoder(file).Decode(&faults); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Apply limit and offset
	start := offset
	end := start + limit
	if start >= len(faults.Features) {
		start = len(faults.Features)
	}
	if end > len(faults.Features) {
		end = len(faults.Features)
	}

	faults.Features = faults.Features[start:end]
	return &faults, nil
}

// GetFileStats returns statistics about a specific file
func (s *JSONStorage) GetFileStats(ctx context.Context, dataType string) (map[string]interface{}, error) {
	files, err := s.ListFiles(dataType)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return map[string]interface{}{
			"data_type": dataType,
			"count":     0,
			"files":     []string{},
		}, nil
	}

	// Get stats for the most recent file
	latestFile := files[len(files)-1]
	var data interface{}
	var err2 error

	switch dataType {
	case "earthquakes":
		data, err2 = s.LoadEarthquakes(ctx, 1000, 0) // Load all for stats
	case "faults":
		data, err2 = s.LoadFaults(ctx, 1000, 0) // Load all for stats
	default:
		return nil, fmt.Errorf("unknown data type: %s", dataType)
	}

	if err2 != nil {
		return nil, err2
	}

	stats := make(map[string]interface{})
	stats["filename"] = latestFile
	stats["data_type"] = dataType
	stats["loaded_at"] = time.Now().Format(time.RFC3339)
	stats["total_files"] = len(files)

	switch v := data.(type) {
	case *models.USGSResponse:
		stats["count"] = len(v.Features)
		stats["metadata"] = v.Metadata
	case *models.Fault:
		stats["count"] = len(v.Features)
		stats["type"] = v.Type
	}

	return stats, nil
}

// Collection tracking methods for JSON storage (simplified implementation)
func (s *JSONStorage) GetLastCollectionTime(ctx context.Context, dataType string) (int64, error) {
	// For JSON storage, we'll use a simple file-based approach
	metadataFile := filepath.Join(s.outputDir, fmt.Sprintf("%s_collection_metadata.json", dataType))

	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		// Return a default time if no metadata file exists (24 hours ago)
		return time.Now().Add(-24 * time.Hour).Unix(), nil
	}

	file, err := os.Open(metadataFile)
	if err != nil {
		return 0, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	var metadata struct {
		LastCollectionTime int64 `json:"last_collection_time"`
	}

	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return 0, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return metadata.LastCollectionTime, nil
}

func (s *JSONStorage) UpdateLastCollectionTime(ctx context.Context, dataType string, collectionTime int64) error {
	metadataFile := filepath.Join(s.outputDir, fmt.Sprintf("%s_collection_metadata.json", dataType))

	metadata := struct {
		LastCollectionTime int64  `json:"last_collection_time"`
		UpdatedAt          string `json:"updated_at"`
	}{
		LastCollectionTime: collectionTime,
		UpdatedAt:          time.Now().Format(time.RFC3339),
	}

	file, err := os.Create(metadataFile)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return nil
}

// LogCollection logs a collection operation
func (s *JSONStorage) LogCollection(ctx context.Context, dataType, source string, startTime int64, recordsCollected int, status string, errorMsg string) error {
	logFile := filepath.Join(s.outputDir, fmt.Sprintf("%s_collection_log.json", dataType))

	logEntry := struct {
		DataType         string `json:"data_type"`
		Source           string `json:"source"`
		StartTime        int64  `json:"start_time"`
		EndTime          int64  `json:"end_time"`
		RecordsCollected int    `json:"records_collected"`
		Status           string `json:"status"`
		ErrorMessage     string `json:"error_message"`
		CreatedAt        string `json:"created_at"`
	}{
		DataType:         dataType,
		Source:           source,
		StartTime:        startTime,
		EndTime:          time.Now().Unix(),
		RecordsCollected: recordsCollected,
		Status:           status,
		ErrorMessage:     errorMsg,
		CreatedAt:        time.Now().Format(time.RFC3339),
	}

	// Read existing logs
	var logs []interface{}
	if _, err := os.Stat(logFile); err == nil {
		file, err := os.Open(logFile)
		if err == nil {
			json.NewDecoder(file).Decode(&logs)
			file.Close()
		}
	}

	// Add new log entry
	logs = append(logs, logEntry)

	// Write back to file
	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(logs); err != nil {
		return fmt.Errorf("failed to encode logs: %w", err)
	}

	return nil
}

// GetCollectionLogs retrieves collection logs
func (s *JSONStorage) GetCollectionLogs(ctx context.Context, dataType string, limit int) ([]CollectionLog, error) {
	logFile := filepath.Join(s.outputDir, fmt.Sprintf("%s_collection_log.json", dataType))

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return []CollectionLog{}, nil
	}

	file, err := os.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var logs []CollectionLog
	if err := json.NewDecoder(file).Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	// Apply limit
	if limit > 0 && len(logs) > limit {
		logs = logs[len(logs)-limit:]
	}

	return logs, nil
}

// GetStatistics returns basic statistics
func (s *JSONStorage) GetStatistics(ctx context.Context) (*Statistics, error) {
	earthquakeFiles, _ := s.ListFiles("earthquakes")
	faultFiles, _ := s.ListFiles("faults")

	stats := &Statistics{
		TotalEarthquakes: int64(len(earthquakeFiles)),
		TotalFaults:      int64(len(faultFiles)),
	}

	return stats, nil
}

// Close closes the storage (no-op for JSON storage)
func (s *JSONStorage) Close() error {
	return nil
}

// Placeholder implementations for interface compatibility
func (s *JSONStorage) GetEarthquakeByID(ctx context.Context, usgsID string) (*models.Earthquake, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) GetEarthquakesByTimeRange(ctx context.Context, startTime, endTime int64) ([]models.Earthquake, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) GetEarthquakesByMagnitudeRange(ctx context.Context, minMag, maxMag float64) ([]models.Earthquake, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) GetEarthquakesByLocation(ctx context.Context, minLat, maxLat, minLon, maxLon float64) ([]models.Earthquake, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) GetSignificantEarthquakes(ctx context.Context, startTime, endTime int64) ([]models.Earthquake, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) DeleteEarthquake(ctx context.Context, usgsID string) error {
	return fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) GetFaultByID(ctx context.Context, faultID string) (*models.FaultFeature, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) GetFaultsByType(ctx context.Context, faultType string) ([]models.FaultFeature, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) GetFaultsByLocation(ctx context.Context, minLat, maxLat, minLon, maxLon float64) ([]models.FaultFeature, error) {
	return nil, fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) DeleteFault(ctx context.Context, faultID string) error {
	return fmt.Errorf("not implemented for JSON storage")
}

func (s *JSONStorage) PurgeAll(ctx context.Context) error {
	// Purge earthquake files
	earthquakeFiles, err := s.ListFiles("earthquakes")
	if err != nil {
		return fmt.Errorf("failed to list earthquake files: %w", err)
	}

	for _, filename := range earthquakeFiles {
		filePath := filepath.Join(s.outputDir, "earthquakes", filename)
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove earthquake file %s: %w", filename, err)
		}
	}

	// Purge fault files
	faultFiles, err := s.ListFiles("faults")
	if err != nil {
		return fmt.Errorf("failed to list fault files: %w", err)
	}

	for _, filename := range faultFiles {
		filePath := filepath.Join(s.outputDir, "faults", filename)
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove fault file %s: %w", filename, err)
		}
	}

	return nil
}

func (s *JSONStorage) PurgeByType(ctx context.Context, dataType string) error {
	files, err := s.ListFiles(dataType)
	if err != nil {
		return fmt.Errorf("failed to list %s files: %w", dataType, err)
	}

	for _, filename := range files {
		filePath := filepath.Join(s.outputDir, dataType, filename)
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove %s file %s: %w", dataType, filename, err)
		}
	}

	return nil
}

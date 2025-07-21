package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"quakewatch-scraper/internal/models"
)

// JSONStorage handles saving data to JSON files
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
func (s *JSONStorage) SaveEarthquakes(earthquakes *models.USGSResponse, filename string) error {
	if filename == "" {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename = fmt.Sprintf("earthquakes_%s.json", timestamp)
	} else if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

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

	return nil
}

// SaveFaults saves fault data to a JSON file
func (s *JSONStorage) SaveFaults(faults *models.Fault, filename string) error {
	if filename == "" {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename = fmt.Sprintf("faults_%s.json", timestamp)
	} else if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

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

// ListFiles lists all JSON files in a specific data type directory
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
		return nil, fmt.Errorf("failed to read directory: %w", err)
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
func (s *JSONStorage) LoadEarthquakes(filename string) (*models.USGSResponse, error) {
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	filePath := filepath.Join(s.outputDir, "earthquakes", filename)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var earthquakes models.USGSResponse
	if err := json.NewDecoder(file).Decode(&earthquakes); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &earthquakes, nil
}

// LoadFaults loads fault data from a JSON file
func (s *JSONStorage) LoadFaults(filename string) (*models.Fault, error) {
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	filePath := filepath.Join(s.outputDir, "faults", filename)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var faults models.Fault
	if err := json.NewDecoder(file).Decode(&faults); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &faults, nil
}

// GetFileStats returns statistics about a specific file
func (s *JSONStorage) GetFileStats(dataType, filename string) (map[string]interface{}, error) {
	var data interface{}
	var err error

	switch dataType {
	case "earthquakes":
		data, err = s.LoadEarthquakes(filename)
	case "faults":
		data, err = s.LoadFaults(filename)
	default:
		return nil, fmt.Errorf("unknown data type: %s", dataType)
	}

	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["filename"] = filename
	stats["data_type"] = dataType
	stats["loaded_at"] = time.Now().Format(time.RFC3339)

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

// PurgeAll deletes all JSON files from both earthquakes and faults directories
func (s *JSONStorage) PurgeAll() error {
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

// PurgeByType deletes all JSON files of a specific data type
func (s *JSONStorage) PurgeByType(dataType string) error {
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

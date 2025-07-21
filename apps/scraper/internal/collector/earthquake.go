package collector

import (
	"fmt"
	"time"

	"quakewatch-scraper/internal/api"
	"quakewatch-scraper/internal/storage"
)

// EarthquakeCollector handles collecting earthquake data
type EarthquakeCollector struct {
	usgsClient *api.USGSClient
	storage    *storage.JSONStorage
}

// NewEarthquakeCollector creates a new earthquake collector
func NewEarthquakeCollector(usgsClient *api.USGSClient, storage *storage.JSONStorage) *EarthquakeCollector {
	return &EarthquakeCollector{
		usgsClient: usgsClient,
		storage:    storage,
	}
}

// CollectRecent collects recent earthquakes (last hour)
func (c *EarthquakeCollector) CollectRecent(limit int, filename string) error {
	fmt.Printf("Collecting recent earthquakes (last hour, limit: %d)...\n", limit)

	earthquakes, err := c.usgsClient.GetRecentEarthquakes(limit)
	if err != nil {
		return fmt.Errorf("failed to fetch recent earthquakes: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectByTimeRange collects earthquakes within a specific time range
func (c *EarthquakeCollector) CollectByTimeRange(startTime, endTime time.Time, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRange(startTime, endTime, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by time range: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectByMagnitude collects earthquakes within a magnitude range
func (c *EarthquakeCollector) CollectByMagnitude(minMag, maxMag float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes with magnitude %.1f to %.1f (limit: %d)...\n", minMag, maxMag, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByMagnitude(minMag, maxMag, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by magnitude: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectSignificant collects significant earthquakes (M4.5+)
func (c *EarthquakeCollector) CollectSignificant(startTime, endTime time.Time, limit int, filename string) error {
	fmt.Printf("Collecting significant earthquakes (M4.5+) from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetSignificantEarthquakes(startTime, endTime, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch significant earthquakes: %w", err)
	}

	fmt.Printf("Found %d significant earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved significant earthquakes to %s\n", filename)
	return nil
}

// CollectByRegion collects earthquakes within a geographic region
func (c *EarthquakeCollector) CollectByRegion(minLat, maxLat, minLon, maxLon float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes in region (%.2f,%.2f) to (%.2f,%.2f) (limit: %d)...\n",
		minLat, minLon, maxLat, maxLon, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByRegion(minLat, maxLat, minLon, maxLon, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by region: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

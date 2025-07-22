package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"quakewatch-scraper/internal/models"
)

// USGSClient handles communication with the USGS Earthquake API
type USGSClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewUSGSClient creates a new USGS API client
func NewUSGSClient(baseURL string, timeout time.Duration) *USGSClient {
	return &USGSClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetEarthquakes fetches earthquake data from USGS API
func (c *USGSClient) GetEarthquakes(params map[string]string) (*models.USGSResponse, error) {
	u, err := url.Parse(c.baseURL + "/query")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()

	// Set default format to geojson
	q.Set("format", "geojson")

	// Add custom parameters
	for key, value := range params {
		q.Set(key, value)
	}

	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var response models.USGSResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// GetRecentEarthquakes fetches earthquakes from the last hour
func (c *USGSClient) GetRecentEarthquakes(limit int) (*models.USGSResponse, error) {
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	params := map[string]string{
		"starttime": startTime.Format("2006-01-02T15:04:05"),
		"endtime":   endTime.Format("2006-01-02T15:04:05"),
		"limit":     strconv.Itoa(limit),
	}

	return c.GetEarthquakes(params)
}

// GetEarthquakesByTimeRange fetches earthquakes within a specific time range
func (c *USGSClient) GetEarthquakesByTimeRange(startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"starttime": startTime.Format("2006-01-02T15:04:05"),
		"endtime":   endTime.Format("2006-01-02T15:04:05"),
		"limit":     strconv.Itoa(limit),
	}

	return c.GetEarthquakes(params)
}

// GetEarthquakesByMagnitude fetches earthquakes within a magnitude range
func (c *USGSClient) GetEarthquakesByMagnitude(minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"minmagnitude": strconv.FormatFloat(minMag, 'f', 1, 64),
		"maxmagnitude": strconv.FormatFloat(maxMag, 'f', 1, 64),
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(params)
}

// GetSignificantEarthquakes fetches significant earthquakes (M4.5+)
func (c *USGSClient) GetSignificantEarthquakes(startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"starttime":    startTime.Format("2006-01-02T15:04:05"),
		"endtime":      endTime.Format("2006-01-02T15:04:05"),
		"minmagnitude": "4.5",
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(params)
}

// GetEarthquakesByRegion fetches earthquakes within a geographic region
func (c *USGSClient) GetEarthquakesByRegion(minLat, maxLat, minLon, maxLon float64, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"minlatitude":  strconv.FormatFloat(minLat, 'f', 2, 64),
		"maxlatitude":  strconv.FormatFloat(maxLat, 'f', 2, 64),
		"minlongitude": strconv.FormatFloat(minLon, 'f', 2, 64),
		"maxlongitude": strconv.FormatFloat(maxLon, 'f', 2, 64),
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(params)
}

// GetEarthquakesByTimeRangeAndMagnitude fetches earthquakes within a time range and magnitude range
func (c *USGSClient) GetEarthquakesByTimeRangeAndMagnitude(startTime, endTime time.Time, minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"starttime":    startTime.Format("2006-01-02T15:04:05"),
		"endtime":      endTime.Format("2006-01-02T15:04:05"),
		"minmagnitude": strconv.FormatFloat(minMag, 'f', 1, 64),
		"maxmagnitude": strconv.FormatFloat(maxMag, 'f', 1, 64),
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(params)
}

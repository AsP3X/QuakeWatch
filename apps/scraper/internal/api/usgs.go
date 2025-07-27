package api

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"quakewatch-scraper/internal/models"
	"quakewatch-scraper/internal/utils"
)

// USGSClient handles communication with the USGS Earthquake API
type USGSClient struct {
	baseURL        string
	enhancedClient *EnhancedAPIClient
	metrics        *utils.CollectionMetrics
	logger         *utils.StructuredLogger
	validator      *utils.DataValidator
}

// NewUSGSClient creates a new USGS API client
func NewUSGSClient(baseURL string, timeout time.Duration, metrics *utils.CollectionMetrics, logger *utils.StructuredLogger) *USGSClient {
	retryStrategy := utils.DefaultRetryStrategy()
	enhancedClient := NewEnhancedAPIClient(
		timeout,
		5,              // circuit breaker threshold
		30*time.Second, // circuit breaker timeout
		retryStrategy,
		metrics,
		logger,
		60, // rate limit per minute
	)

	// Set up data validation rules
	validator := utils.NewDataValidator()
	// Only validate that we have a response structure - the USGS API is reliable
	// so we don't need overly strict validation rules that cause false warnings
	validator.AddRule(utils.NewRequiredFieldRule("Type", 0.3))
	validator.AddRule(utils.NewRequiredFieldRule("Features", 0.7))

	return &USGSClient{
		baseURL:        baseURL,
		enhancedClient: enhancedClient,
		metrics:        metrics,
		logger:         logger,
		validator:      validator,
	}
}

// GetEarthquakes fetches earthquake data from USGS API with enhanced error handling
func (c *USGSClient) GetEarthquakes(ctx context.Context, params map[string]string) (*models.USGSResponse, error) {
	u, err := url.Parse(c.baseURL + "/query")
	if err != nil {
		return nil, utils.NewCollectionError(
			utils.ErrorTypeConfiguration,
			"usgs_client",
			"failed to parse URL",
			false,
			err,
		)
	}

	q := u.Query()
	q.Set("format", "geojson")

	// Add custom parameters
	for key, value := range params {
		q.Set(key, value)
	}

	u.RawQuery = q.Encode()

	c.metrics.RecordCollectionStart("usgs_api")
	defer func() {
		c.metrics.RecordCollectionEnd("usgs_api", 0, 0, 0)
	}()

	resp, err := c.enhancedClient.GetWithRetry(ctx, u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response models.USGSResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, utils.NewCollectionError(
			utils.ErrorTypeValidation,
			"usgs_client",
			"failed to decode response",
			false,
			err,
		)
	}

	// Validate the response
	valid, errors, score := c.validator.Validate(&response)
	c.metrics.RecordDataQuality("usgs_api", score)

	if !valid {
		c.logger.LogValidation(ctx, "usgs_api", false, score, utils.ConvertValidationErrors(errors))
	} else {
		c.logger.LogValidation(ctx, "usgs_api", true, score, nil)
	}

	// Log collection event
	c.logger.LogCollection(ctx, utils.CollectionEvent{
		Source:           "usgs_api",
		RecordsCollected: len(response.Features),
		Duration:         0, // Will be set by caller
		Status:           "success",
		QualityScore:     score,
		Metadata: map[string]interface{}{
			"url":    u.String(),
			"params": params,
		},
	})

	return &response, nil
}

// GetRecentEarthquakes fetches earthquakes from the last hour
func (c *USGSClient) GetRecentEarthquakes(ctx context.Context, limit int) (*models.USGSResponse, error) {
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	params := map[string]string{
		"starttime": startTime.Format("2006-01-02T15:04:05"),
		"endtime":   endTime.Format("2006-01-02T15:04:05"),
		"limit":     strconv.Itoa(limit),
	}

	return c.GetEarthquakes(ctx, params)
}

// GetEarthquakesByTimeRange fetches earthquakes within a specific time range
func (c *USGSClient) GetEarthquakesByTimeRange(ctx context.Context, startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"starttime": startTime.Format("2006-01-02T15:04:05"),
		"endtime":   endTime.Format("2006-01-02T15:04:05"),
		"limit":     strconv.Itoa(limit),
	}

	return c.GetEarthquakes(ctx, params)
}

// GetEarthquakesByMagnitude fetches earthquakes within a magnitude range
func (c *USGSClient) GetEarthquakesByMagnitude(ctx context.Context, minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"minmagnitude": strconv.FormatFloat(minMag, 'f', 1, 64),
		"maxmagnitude": strconv.FormatFloat(maxMag, 'f', 1, 64),
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(ctx, params)
}

// GetSignificantEarthquakes fetches significant earthquakes (M4.5+)
func (c *USGSClient) GetSignificantEarthquakes(ctx context.Context, startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"starttime":    startTime.Format("2006-01-02T15:04:05"),
		"endtime":      endTime.Format("2006-01-02T15:04:05"),
		"minmagnitude": "4.5",
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(ctx, params)
}

// GetEarthquakesByRegion fetches earthquakes within a geographic region
func (c *USGSClient) GetEarthquakesByRegion(ctx context.Context, minLat, maxLat, minLon, maxLon float64, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"minlatitude":  strconv.FormatFloat(minLat, 'f', 2, 64),
		"maxlatitude":  strconv.FormatFloat(maxLat, 'f', 2, 64),
		"minlongitude": strconv.FormatFloat(minLon, 'f', 2, 64),
		"maxlongitude": strconv.FormatFloat(maxLon, 'f', 2, 64),
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(ctx, params)
}

// GetEarthquakesByTimeRangeAndMagnitude fetches earthquakes by time range and magnitude
func (c *USGSClient) GetEarthquakesByTimeRangeAndMagnitude(ctx context.Context, startTime, endTime time.Time, minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	params := map[string]string{
		"starttime":    startTime.Format("2006-01-02T15:04:05"),
		"endtime":      endTime.Format("2006-01-02T15:04:05"),
		"minmagnitude": strconv.FormatFloat(minMag, 'f', 1, 64),
		"maxmagnitude": strconv.FormatFloat(maxMag, 'f', 1, 64),
		"limit":        strconv.Itoa(limit),
	}

	return c.GetEarthquakes(ctx, params)
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (c *USGSClient) GetCircuitBreakerStats() map[string]interface{} {
	return c.enhancedClient.GetCircuitBreakerStats()
}

// ResetCircuitBreaker resets the circuit breaker
func (c *USGSClient) ResetCircuitBreaker() {
	c.enhancedClient.ResetCircuitBreaker()
}

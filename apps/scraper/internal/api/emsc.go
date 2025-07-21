package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"quakewatch-scraper/internal/models"
)

// EMSCClient handles communication with the EMSC-CSEM API
type EMSCClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewEMSCClient creates a new EMSC API client
func NewEMSCClient(baseURL string, timeout time.Duration) *EMSCClient {
	return &EMSCClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// GetFaults fetches fault data from EMSC API
func (c *EMSCClient) GetFaults() (*models.Fault, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/gem_active_faults.geojson")
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var faults models.Fault
	if err := json.NewDecoder(resp.Body).Decode(&faults); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &faults, nil
}

// GetFaultsWithRetry fetches fault data with retry logic
func (c *EMSCClient) GetFaultsWithRetry(maxRetries int, retryDelay time.Duration) (*models.Fault, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		faults, err := c.GetFaults()
		if err == nil {
			return faults, nil
		}

		lastErr = err

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to fetch faults after %d attempts: %w", maxRetries+1, lastErr)
}

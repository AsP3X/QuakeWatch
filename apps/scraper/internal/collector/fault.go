package collector

import (
	"fmt"
	"time"

	"quakewatch-scraper/internal/api"
	"quakewatch-scraper/internal/storage"
)

// FaultCollector handles collecting fault data
type FaultCollector struct {
	emscClient *api.EMSCClient
	storage    *storage.JSONStorage
}

// NewFaultCollector creates a new fault collector
func NewFaultCollector(emscClient *api.EMSCClient, storage *storage.JSONStorage) *FaultCollector {
	return &FaultCollector{
		emscClient: emscClient,
		storage:    storage,
	}
}

// CollectFaults collects fault data from EMSC
func (c *FaultCollector) CollectFaults(filename string) error {
	fmt.Println("Collecting fault data from EMSC...")

	faults, err := c.emscClient.GetFaults()
	if err != nil {
		return fmt.Errorf("failed to fetch fault data: %w", err)
	}

	fmt.Printf("Found %d fault features\n", len(faults.Features))

	if err := c.storage.SaveFaults(faults, filename); err != nil {
		return fmt.Errorf("failed to save fault data: %w", err)
	}

	fmt.Printf("Saved fault data to %s\n", filename)
	return nil
}

// UpdateFaults updates fault data with retry logic
func (c *FaultCollector) UpdateFaults(filename string, maxRetries int, retryDelay time.Duration) error {
	fmt.Printf("Updating fault data from EMSC (max retries: %d)...\n", maxRetries)

	faults, err := c.emscClient.GetFaultsWithRetry(maxRetries, retryDelay)
	if err != nil {
		return fmt.Errorf("failed to fetch fault data with retry: %w", err)
	}

	fmt.Printf("Found %d fault features\n", len(faults.Features))

	if err := c.storage.SaveFaults(faults, filename); err != nil {
		return fmt.Errorf("failed to save fault data: %w", err)
	}

	fmt.Printf("Updated fault data saved to %s\n", filename)
	return nil
}

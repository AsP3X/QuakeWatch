package collector

import (
	"context"
	"fmt"
	"time"

	"quakewatch-scraper/internal/api"
	"quakewatch-scraper/internal/models"
	"quakewatch-scraper/internal/storage"
)

// FaultCollector handles collecting fault data
type FaultCollector struct {
	emscClient *api.EMSCClient
	storage    storage.Storage
}

// NewFaultCollector creates a new fault collector
func NewFaultCollector(emscClient *api.EMSCClient, storage storage.Storage) *FaultCollector {
	return &FaultCollector{
		emscClient: emscClient,
		storage:    storage,
	}
}

// CollectFaults collects fault data from EMSC
func (c *FaultCollector) CollectFaults(ctx context.Context) error {
	fmt.Println("Collecting fault data from EMSC...")

	faults, err := c.emscClient.GetFaults()
	if err != nil {
		return fmt.Errorf("failed to fetch fault data: %w", err)
	}

	fmt.Printf("Found %d fault features\n", len(faults.Features))

	if err := c.storage.SaveFaults(ctx, faults); err != nil {
		return fmt.Errorf("failed to save fault data: %w", err)
	}

	fmt.Println("Saved fault data")
	return nil
}

// UpdateFaults updates fault data with retry logic
func (c *FaultCollector) UpdateFaults(ctx context.Context, maxRetries int, retryDelay time.Duration) error {
	fmt.Printf("Updating fault data from EMSC (max retries: %d)...\n", maxRetries)

	faults, err := c.emscClient.GetFaultsWithRetry(maxRetries, retryDelay)
	if err != nil {
		return fmt.Errorf("failed to fetch fault data with retry: %w", err)
	}

	fmt.Printf("Found %d fault features\n", len(faults.Features))

	if err := c.storage.SaveFaults(ctx, faults); err != nil {
		return fmt.Errorf("failed to save fault data: %w", err)
	}

	fmt.Println("Updated fault data saved")
	return nil
}

// CollectFaultsData collects fault data from EMSC and returns the data without saving
func (c *FaultCollector) CollectFaultsData() (*models.Fault, error) {
	fmt.Println("Collecting fault data from EMSC...")

	faults, err := c.emscClient.GetFaults()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fault data: %w", err)
	}

	fmt.Printf("Found %d fault features\n", len(faults.Features))
	return faults, nil
}

// UpdateFaultsData updates fault data with retry logic and returns the data without saving
func (c *FaultCollector) UpdateFaultsData(maxRetries int, retryDelay time.Duration) (*models.Fault, error) {
	fmt.Printf("Updating fault data from EMSC (max retries: %d)...\n", maxRetries)

	faults, err := c.emscClient.GetFaultsWithRetry(maxRetries, retryDelay)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch fault data with retry: %w", err)
	}

	fmt.Printf("Found %d fault features\n", len(faults.Features))
	return faults, nil
}

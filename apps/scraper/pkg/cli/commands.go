package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"quakewatch-scraper/internal/api"
	"quakewatch-scraper/internal/collector"
	"quakewatch-scraper/internal/config"
	sched "quakewatch-scraper/internal/scheduler"
	"quakewatch-scraper/internal/storage"
)

// App represents the main CLI application
type App struct {
	rootCmd *cobra.Command
	cfg     *config.Config
}

// outputToStdout outputs data to stdout in JSON format
func (a *App) outputToStdout(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// NewApp creates a new CLI application
func NewApp() *App {
	app := &App{
		rootCmd: &cobra.Command{
			Use:   "quakewatch-scraper",
			Short: "QuakeWatch Data Scraper - Collect earthquake and fault data",
			Long:  `A Go application for collecting earthquake and fault data from various sources and saving to JSON files.`,
		},
	}

	// Set up the PersistentPreRunE after creating the app
	app.rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip configuration loading for version command
		if cmd.Name() == "version" {
			return nil
		}

		// Load configuration for all commands
		configPath, _ := cmd.Flags().GetString("config")

		// Only prompt for configuration if no command is given (showBanner)
		if cmd.Name() == "quakewatch-scraper" {
			// Load configuration - this will handle missing config files interactively
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return err
			}
			app.cfg = cfg
		} else {
			// For other commands, load configuration without prompting
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				// If config loading fails, use default configuration
				app.cfg = config.DefaultConfig()
			} else {
				app.cfg = cfg
			}
		}

		return nil
	}

	app.setupCommands()
	app.setupFlags()

	// Set the banner function for when no command is provided
	app.rootCmd.Run = app.showBanner

	return app
}

func (a *App) setupCommands() {
	// Add earthquake commands
	a.rootCmd.AddCommand(a.newEarthquakeCmd())

	// Add fault commands
	a.rootCmd.AddCommand(a.newFaultCmd())

	// Add interval commands
	a.rootCmd.AddCommand(a.newIntervalCmd())

	// Add utility commands
	a.rootCmd.AddCommand(a.newValidateCmd())
	a.rootCmd.AddCommand(a.newStatsCmd())
	a.rootCmd.AddCommand(a.newListCmd())
	a.rootCmd.AddCommand(a.newPurgeCmd())
	a.rootCmd.AddCommand(a.newHealthCmd())
	a.rootCmd.AddCommand(a.newVersionCmd())
	a.rootCmd.AddCommand(a.newConfigCmd())
}

func (a *App) setupFlags() {
	a.rootCmd.PersistentFlags().StringP("config", "c", "./configs/config.yaml", "Configuration file path")
	a.rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	a.rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress output")
	a.rootCmd.PersistentFlags().String("log-level", "info", "Set log level (error, warn, info, debug)")
	a.rootCmd.PersistentFlags().StringP("output-dir", "o", "./data", "Output directory for JSON files")
	a.rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without executing")
	a.rootCmd.PersistentFlags().Bool("stdout", false, "Output data to stdout instead of saving to file")
}

func (a *App) Run(args []string) error {
	// Remove the first argument (binary name) - it could be "./bin/quakewatch-scraper" or "quakewatch-scraper"
	if len(args) > 0 {
		args = args[1:]
	}

	// Set up the command
	a.rootCmd.SetArgs(args)

	// Execute the command - configuration will be loaded in PreRun
	return a.rootCmd.Execute()
}

// newEarthquakeCmd creates the earthquake command
func (a *App) newEarthquakeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "earthquakes",
		Short: "Collect earthquake data",
		Long:  `Collect earthquake data from USGS API`,
	}

	// Recent earthquakes command
	recentCmd := &cobra.Command{
		Use:   "recent",
		Short: "Collect recent earthquakes (last hour)",
		RunE:  a.runRecentEarthquakes,
	}
	recentCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	recentCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	cmd.AddCommand(recentCmd)

	// Time range command
	timeRangeCmd := &cobra.Command{
		Use:   "time-range",
		Short: "Collect earthquakes by time range",
		RunE:  a.runTimeRangeEarthquakes,
	}
	timeRangeCmd.Flags().String("start", "", "Start time (YYYY-MM-DD)")
	timeRangeCmd.Flags().String("end", "", "End time (YYYY-MM-DD)")
	timeRangeCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	timeRangeCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	if err := timeRangeCmd.MarkFlagRequired("start"); err != nil {
		panic(fmt.Sprintf("failed to mark start flag as required: %v", err))
	}
	if err := timeRangeCmd.MarkFlagRequired("end"); err != nil {
		panic(fmt.Sprintf("failed to mark end flag as required: %v", err))
	}
	cmd.AddCommand(timeRangeCmd)

	// Magnitude command
	magnitudeCmd := &cobra.Command{
		Use:   "magnitude",
		Short: "Collect earthquakes by magnitude range",
		RunE:  a.runMagnitudeEarthquakes,
	}
	magnitudeCmd.Flags().Float64("min", 0.0, "Minimum magnitude")
	magnitudeCmd.Flags().Float64("max", 10.0, "Maximum magnitude")
	magnitudeCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	magnitudeCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	if err := magnitudeCmd.MarkFlagRequired("min"); err != nil {
		panic(fmt.Sprintf("failed to mark min flag as required: %v", err))
	}
	if err := magnitudeCmd.MarkFlagRequired("max"); err != nil {
		panic(fmt.Sprintf("failed to mark max flag as required: %v", err))
	}
	cmd.AddCommand(magnitudeCmd)

	// Significant command
	significantCmd := &cobra.Command{
		Use:   "significant",
		Short: "Collect significant earthquakes (M4.5+)",
		RunE:  a.runSignificantEarthquakes,
	}
	significantCmd.Flags().String("start", "", "Start time (YYYY-MM-DD)")
	significantCmd.Flags().String("end", "", "End time (YYYY-MM-DD)")
	significantCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	significantCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	if err := significantCmd.MarkFlagRequired("start"); err != nil {
		panic(fmt.Sprintf("failed to mark start flag as required: %v", err))
	}
	if err := significantCmd.MarkFlagRequired("end"); err != nil {
		panic(fmt.Sprintf("failed to mark end flag as required: %v", err))
	}
	cmd.AddCommand(significantCmd)

	// Region command
	regionCmd := &cobra.Command{
		Use:   "region",
		Short: "Collect earthquakes by geographic region",
		RunE:  a.runRegionEarthquakes,
	}
	regionCmd.Flags().Float64("min-lat", -90.0, "Minimum latitude")
	regionCmd.Flags().Float64("max-lat", 90.0, "Maximum latitude")
	regionCmd.Flags().Float64("min-lon", -180.0, "Minimum longitude")
	regionCmd.Flags().Float64("max-lon", 180.0, "Maximum longitude")
	regionCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	regionCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	if err := regionCmd.MarkFlagRequired("min-lat"); err != nil {
		panic(fmt.Sprintf("failed to mark min-lat flag as required: %v", err))
	}
	if err := regionCmd.MarkFlagRequired("max-lat"); err != nil {
		panic(fmt.Sprintf("failed to mark max-lat flag as required: %v", err))
	}
	if err := regionCmd.MarkFlagRequired("min-lon"); err != nil {
		panic(fmt.Sprintf("failed to mark min-lon flag as required: %v", err))
	}
	if err := regionCmd.MarkFlagRequired("max-lon"); err != nil {
		panic(fmt.Sprintf("failed to mark max-lon flag as required: %v", err))
	}
	cmd.AddCommand(regionCmd)

	// Country command
	countryCmd := &cobra.Command{
		Use:   "country",
		Short: "Collect earthquakes by country",
		RunE:  a.runCountryEarthquakes,
	}
	countryCmd.Flags().String("country", "", "Country name to filter by")
	countryCmd.Flags().String("start", "", "Start time (YYYY-MM-DD)")
	countryCmd.Flags().String("end", "", "End time (YYYY-MM-DD)")
	countryCmd.Flags().Float64("min-mag", 0.0, "Minimum magnitude")
	countryCmd.Flags().Float64("max-mag", 10.0, "Maximum magnitude")
	countryCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	countryCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	if err := countryCmd.MarkFlagRequired("country"); err != nil {
		panic(fmt.Sprintf("failed to mark country flag as required: %v", err))
	}
	cmd.AddCommand(countryCmd)

	return cmd
}

// newFaultCmd creates the fault command
func (a *App) newFaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faults",
		Short: "Collect fault data",
		Long:  `Collect fault data from EMSC API`,
	}

	// Collect command
	collectCmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect fault data from EMSC",
		RunE:  a.runCollectFaults,
	}
	collectCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	cmd.AddCommand(collectCmd)

	// Update command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update fault data with retry logic",
		RunE:  a.runUpdateFaults,
	}
	updateCmd.Flags().StringP("filename", "f", "", "Custom filename (without extension)")
	updateCmd.Flags().Int("retries", 3, "Number of retry attempts")
	updateCmd.Flags().Duration("retry-delay", 5*time.Second, "Delay between retries")
	cmd.AddCommand(updateCmd)

	return cmd
}

// newValidateCmd creates the validate command
func (a *App) newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate data integrity",
		RunE:  a.runValidate,
	}
	cmd.Flags().StringP("type", "t", "all", "Data type (earthquakes, faults, all)")
	cmd.Flags().StringP("file", "f", "", "Specific file to validate")
	return cmd
}

// newStatsCmd creates the stats command
func (a *App) newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show data statistics",
		RunE:  a.runStats,
	}
	cmd.Flags().StringP("type", "t", "all", "Data type (earthquakes, faults, all)")
	cmd.Flags().StringP("file", "f", "", "Specific file to show stats for")
	return cmd
}

// newListCmd creates the list command
func (a *App) newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available data files",
		RunE:  a.runList,
	}
	cmd.Flags().StringP("type", "t", "all", "Data type (earthquakes, faults, all)")
	return cmd
}

// newPurgeCmd creates the purge command
func (a *App) newPurgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Delete all collected data files",
		Long:  `Delete all JSON data files from the storage directory. Use with caution as this action cannot be undone.`,
		RunE:  a.runPurge,
	}
	cmd.Flags().StringP("type", "t", "all", "Data type to purge (earthquakes, faults, all)")
	cmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")
	return cmd
}

// newHealthCmd creates the health command
func (a *App) newHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check system health",
		RunE:  a.runHealth,
	}
	return cmd
}

// newVersionCmd creates the version command
func (a *App) newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run:   a.runVersion,
	}
	return cmd
}

// newConfigCmd creates the configuration command
func (a *App) newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage application configuration",
		Long:  `Create or update the application configuration file through interactive prompts.`,
		RunE:  a.runConfig,
	}
	return cmd
}

// Helper methods for command execution
func (a *App) runRecentEarthquakes(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	// Use configuration values
	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	if stdout {
		earthquakes, err := collector.CollectRecentData(limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectRecent(limit, filename)
}

func (a *App) runTimeRangeEarthquakes(cmd *cobra.Command, args []string) error {
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	// Use configuration values
	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	if stdout {
		earthquakes, err := collector.CollectByTimeRangeData(startTime, endTime, limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectByTimeRange(startTime, endTime, limit, filename)
}

func (a *App) runMagnitudeEarthquakes(cmd *cobra.Command, args []string) error {
	minMag, _ := cmd.Flags().GetFloat64("min")
	maxMag, _ := cmd.Flags().GetFloat64("max")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	// Use configuration values
	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	if stdout {
		earthquakes, err := collector.CollectByMagnitudeData(minMag, maxMag, limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectByMagnitude(minMag, maxMag, limit, filename)
}

func (a *App) runSignificantEarthquakes(cmd *cobra.Command, args []string) error {
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	// Use configuration values
	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	if stdout {
		earthquakes, err := collector.CollectSignificantData(startTime, endTime, limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectSignificant(startTime, endTime, limit, filename)
}

func (a *App) runRegionEarthquakes(cmd *cobra.Command, args []string) error {
	minLat, _ := cmd.Flags().GetFloat64("min-lat")
	maxLat, _ := cmd.Flags().GetFloat64("max-lat")
	minLon, _ := cmd.Flags().GetFloat64("min-lon")
	maxLon, _ := cmd.Flags().GetFloat64("max-lon")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	// Use configuration values
	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	if stdout {
		earthquakes, err := collector.CollectByRegionData(minLat, maxLat, minLon, maxLon, limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectByRegion(minLat, maxLat, minLon, maxLon, limit, filename)
}

func (a *App) runCountryEarthquakes(cmd *cobra.Command, args []string) error {
	country, _ := cmd.Flags().GetString("country")
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	minMag, _ := cmd.Flags().GetFloat64("min-mag")
	maxMag, _ := cmd.Flags().GetFloat64("max-mag")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	// Set default time range if not provided (last 30 days)
	var startTime, endTime time.Time
	if startStr == "" || endStr == "" {
		endTime = time.Now()
		startTime = endTime.AddDate(0, 0, -30) // 30 days ago
	} else {
		var err error
		startTime, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return fmt.Errorf("invalid start time format: %w", err)
		}

		endTime, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return fmt.Errorf("invalid end time format: %w", err)
		}
	}

	// Use configuration values
	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	if stdout {
		earthquakes, err := collector.CollectByCountryData(country, startTime, endTime, minMag, maxMag, limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectByCountry(country, startTime, endTime, minMag, maxMag, limit, filename)
}

func (a *App) runCollectFaults(cmd *cobra.Command, args []string) error {
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	emscClient := api.NewEMSCClient(a.cfg.API.EMSC.BaseURL, a.cfg.API.EMSC.Timeout)
	collector := collector.NewFaultCollector(emscClient, storage)

	if stdout {
		faults, err := collector.CollectFaultsData()
		if err != nil {
			return err
		}
		return a.outputToStdout(faults)
	}

	return collector.CollectFaults(filename)
}

func (a *App) runUpdateFaults(cmd *cobra.Command, args []string) error {
	filename, _ := cmd.Flags().GetString("filename")
	retries, _ := cmd.Flags().GetInt("retries")
	retryDelay, _ := cmd.Flags().GetDuration("retry-delay")
	stdout, _ := cmd.Flags().GetBool("stdout")

	// Use configuration values if not provided
	if retries == 0 {
		retries = a.cfg.Collection.RetryAttempts
	}
	if retryDelay == 0 {
		retryDelay = a.cfg.Collection.RetryDelay
	}

	// Initialize components with configuration
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	emscClient := api.NewEMSCClient(a.cfg.API.EMSC.BaseURL, a.cfg.API.EMSC.Timeout)
	collector := collector.NewFaultCollector(emscClient, storage)

	if stdout {
		faults, err := collector.UpdateFaultsData(retries, retryDelay)
		if err != nil {
			return err
		}
		return a.outputToStdout(faults)
	}

	return collector.UpdateFaults(filename, retries, retryDelay)
}

func (a *App) runValidate(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")
	file, _ := cmd.Flags().GetString("file")

	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)

	if file != "" {
		// Validate specific file
		stats, err := storage.GetFileStats(dataType, file)
		if err != nil {
			return fmt.Errorf("failed to validate file: %w", err)
		}
		fmt.Printf("File validation successful: %+v\n", stats)
		return nil
	}

	if dataType == "all" {
		fmt.Println("Validating all data files:")

		// Validate earthquake files
		earthquakeFiles, err := storage.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("Error listing earthquake files: %v\n", err)
		} else {
			fmt.Println("Earthquakes:")
			for _, filename := range earthquakeFiles {
				stats, err := storage.GetFileStats("earthquakes", filename)
				if err != nil {
					fmt.Printf("  ✗ %s: %v\n", filename, err)
					continue
				}
				fmt.Printf("  ✓ %s: %d records\n", filename, stats["count"])
			}
		}

		// Validate fault files
		faultFiles, err := storage.ListFiles("faults")
		if err != nil {
			fmt.Printf("Error listing fault files: %v\n", err)
		} else {
			fmt.Println("Faults:")
			for _, filename := range faultFiles {
				stats, err := storage.GetFileStats("faults", filename)
				if err != nil {
					fmt.Printf("  ✗ %s: %v\n", filename, err)
					continue
				}
				fmt.Printf("  ✓ %s: %d records\n", filename, stats["count"])
			}
		}

		return nil
	}

	// Validate specific type
	files, err := storage.ListFiles(dataType)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	for _, filename := range files {
		stats, err := storage.GetFileStats(dataType, filename)
		if err != nil {
			fmt.Printf("Failed to validate %s: %v\n", filename, err)
			continue
		}
		fmt.Printf("✓ %s: %d records\n", filename, stats["count"])
	}

	return nil
}

func (a *App) runStats(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")
	file, _ := cmd.Flags().GetString("file")

	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)

	if file != "" {
		// Show stats for specific file
		stats, err := storage.GetFileStats(dataType, file)
		if err != nil {
			return fmt.Errorf("failed to get file stats: %w", err)
		}
		fmt.Printf("File Statistics:\n")
		for key, value := range stats {
			fmt.Printf("  %s: %v\n", key, value)
		}
		return nil
	}

	if dataType == "all" {
		fmt.Println("Statistics for all data:")

		// Show earthquake stats
		earthquakeFiles, err := storage.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("  Error listing earthquake files: %v\n", err)
		} else {
			fmt.Printf("  Earthquake files: %d\n", len(earthquakeFiles))
			totalEarthquakeRecords := 0
			for _, filename := range earthquakeFiles {
				stats, err := storage.GetFileStats("earthquakes", filename)
				if err != nil {
					fmt.Printf("    Failed to get stats for %s: %v\n", filename, err)
					continue
				}
				if count, ok := stats["count"].(int); ok {
					totalEarthquakeRecords += count
				}
			}
			fmt.Printf("  Total earthquake records: %d\n", totalEarthquakeRecords)
		}

		// Show fault stats
		faultFiles, err := storage.ListFiles("faults")
		if err != nil {
			fmt.Printf("  Error listing fault files: %v\n", err)
		} else {
			fmt.Printf("  Fault files: %d\n", len(faultFiles))
			totalFaultRecords := 0
			for _, filename := range faultFiles {
				stats, err := storage.GetFileStats("faults", filename)
				if err != nil {
					fmt.Printf("    Failed to get stats for %s: %v\n", filename, err)
					continue
				}
				if count, ok := stats["count"].(int); ok {
					totalFaultRecords += count
				}
			}
			fmt.Printf("  Total fault records: %d\n", totalFaultRecords)
		}

		return nil
	}

	// Show stats for specific type
	files, err := storage.ListFiles(dataType)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	fmt.Printf("Statistics for %s data:\n", dataType)
	fmt.Printf("  Total files: %d\n", len(files))

	totalRecords := 0
	for _, filename := range files {
		stats, err := storage.GetFileStats(dataType, filename)
		if err != nil {
			fmt.Printf("  Failed to get stats for %s: %v\n", filename, err)
			continue
		}
		if count, ok := stats["count"].(int); ok {
			totalRecords += count
		}
	}
	fmt.Printf("  Total records: %d\n", totalRecords)

	return nil
}

func (a *App) runList(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")

	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)

	if dataType == "all" {
		fmt.Println("Available data files:")
		fmt.Println("Earthquakes:")
		earthquakeFiles, err := storage.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("  Error listing earthquake files: %v\n", err)
		} else {
			for _, file := range earthquakeFiles {
				fmt.Printf("  %s\n", file)
			}
		}

		fmt.Println("Faults:")
		faultFiles, err := storage.ListFiles("faults")
		if err != nil {
			fmt.Printf("  Error listing fault files: %v\n", err)
		} else {
			for _, file := range faultFiles {
				fmt.Printf("  %s\n", file)
			}
		}
	} else {
		files, err := storage.ListFiles(dataType)
		if err != nil {
			return fmt.Errorf("failed to list files: %w", err)
		}

		fmt.Printf("Available %s files:\n", dataType)
		for _, file := range files {
			fmt.Printf("  %s\n", file)
		}
	}

	return nil
}

func (a *App) runPurge(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)

	if dryRun {
		fmt.Println("DRY RUN - Files that would be deleted:")

		if dataType == "all" || dataType == "earthquakes" {
			earthquakeFiles, err := storage.ListFiles("earthquakes")
			if err != nil {
				fmt.Printf("  Error listing earthquake files: %v\n", err)
			} else {
				fmt.Printf("  Earthquake files (%d):\n", len(earthquakeFiles))
				for _, filename := range earthquakeFiles {
					fmt.Printf("    %s\n", filename)
				}
			}
		}

		if dataType == "all" || dataType == "faults" {
			faultFiles, err := storage.ListFiles("faults")
			if err != nil {
				fmt.Printf("  Error listing fault files: %v\n", err)
			} else {
				fmt.Printf("  Fault files (%d):\n", len(faultFiles))
				for _, filename := range faultFiles {
					fmt.Printf("    %s\n", filename)
				}
			}
		}

		return nil
	}

	// Show what will be deleted
	fmt.Printf("About to delete %s data files:\n", dataType)

	var totalFiles int

	if dataType == "all" || dataType == "earthquakes" {
		earthquakeFiles, err := storage.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("  Error listing earthquake files: %v\n", err)
		} else {
			fmt.Printf("  Earthquake files: %d\n", len(earthquakeFiles))
			totalFiles += len(earthquakeFiles)
		}
	}

	if dataType == "all" || dataType == "faults" {
		faultFiles, err := storage.ListFiles("faults")
		if err != nil {
			fmt.Printf("  Error listing fault files: %v\n", err)
		} else {
			fmt.Printf("  Fault files: %d\n", len(faultFiles))
			totalFiles += len(faultFiles)
		}
	}

	if totalFiles == 0 {
		fmt.Println("No files to delete.")
		return nil
	}

	// Ask for confirmation unless force flag is used
	if !force {
		fmt.Printf("\nThis will permanently delete %d files. Are you sure? (y/N): ", totalFiles)

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		if response != "y" && response != "Y" && response != "yes" && response != "YES" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Perform the deletion
	if dataType == "all" {
		if err := storage.PurgeAll(); err != nil {
			return fmt.Errorf("failed to purge all files: %w", err)
		}
		fmt.Printf("Successfully deleted %d files.\n", totalFiles)
	} else {
		if err := storage.PurgeByType(dataType); err != nil {
			return fmt.Errorf("failed to purge %s files: %w", dataType, err)
		}
		fmt.Printf("Successfully deleted %s files.\n", dataType)
	}

	return nil
}

func (a *App) runHealth(cmd *cobra.Command, args []string) error {
	fmt.Println("System Health Check:")

	// Check USGS API
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, 10*time.Second)
	_, err := usgsClient.GetRecentEarthquakes(1)
	if err != nil {
		fmt.Printf("  ✗ USGS API: %v\n", err)
	} else {
		fmt.Println("  ✓ USGS API: OK")
	}

	// Check EMSC API
	emscClient := api.NewEMSCClient(a.cfg.API.EMSC.BaseURL, 10*time.Second)
	_, err = emscClient.GetFaults()
	if err != nil {
		fmt.Printf("  ✗ EMSC API: %v\n", err)
	} else {
		fmt.Println("  ✓ EMSC API: OK")
	}

	// Check storage
	storage := storage.NewJSONStorage(a.cfg.Storage.OutputDir)
	_, err = storage.ListFiles("earthquakes")
	if err != nil {
		fmt.Printf("  ✗ Storage: %v\n", err)
	} else {
		fmt.Println("  ✓ Storage: OK")
	}

	// Check database if enabled
	if a.cfg.Database.Enabled {
		if err := a.checkDatabaseHealth(); err != nil {
			fmt.Printf("  ✗ Database: %v\n", err)
		} else {
			fmt.Println("  ✓ Database: OK")
		}
	} else {
		fmt.Println("  ⚪ Database: Disabled")
	}

	return nil
}

func (a *App) runVersion(cmd *cobra.Command, args []string) {
	fmt.Println("QuakeWatch Scraper v1.0.0")
	fmt.Println("Go version: 1.21")
	fmt.Println("Build date: " + time.Now().Format("2006-01-02"))
}

func (a *App) runConfig(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")

	fmt.Println("QuakeWatch Scraper Configuration Setup")
	fmt.Println("=====================================")

	// Create configuration interactively
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}

	// Update the app's configuration
	a.cfg = cfg

	fmt.Println("\nConfiguration setup completed successfully!")
	return nil
}

// checkDatabaseHealth checks the database connectivity
func (a *App) checkDatabaseHealth() error {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		a.cfg.Database.Host,
		a.cfg.Database.Port,
		a.cfg.Database.User,
		a.cfg.Database.Password,
		a.cfg.Database.Database,
		a.cfg.Database.SSLMode,
	)

	// Try to connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Database.ConnectionTimeout)
	defer cancel()

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// showBanner displays the application banner when no command is provided
func (a *App) showBanner(cmd *cobra.Command, args []string) {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  🌋 QuakeWatch Scraper 🌋                    ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  A powerful tool for collecting earthquake and fault data    ║")
	fmt.Println("║  from various geological sources and APIs.                   ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  Version: 1.0.0                                              ║")
	fmt.Println("║  Built with Go                                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Show the help after the banner
	if err := cmd.Help(); err != nil {
		fmt.Printf("Error showing help: %v\n", err)
	}
}

// newIntervalCmd creates the interval command
func (a *App) newIntervalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interval",
		Short: "Run commands at specified intervals",
		Long:  `Execute scraping commands at regular intervals with configurable options.`,
	}

	// Add earthquake interval commands
	cmd.AddCommand(a.newIntervalEarthquakesCmd())

	// Add fault interval commands
	cmd.AddCommand(a.newIntervalFaultsCmd())

	// Add custom interval commands
	cmd.AddCommand(a.newIntervalCustomCmd())

	return cmd
}

// newIntervalEarthquakesCmd creates the interval earthquakes command
func (a *App) newIntervalEarthquakesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "earthquakes",
		Short: "Run earthquake collection at intervals",
		Long:  `Execute earthquake collection commands at specified intervals.`,
	}

	// Recent earthquakes interval command
	recentCmd := &cobra.Command{
		Use:   "recent",
		Short: "Collect recent earthquakes at intervals",
		RunE:  a.runIntervalRecentEarthquakes,
	}
	a.addIntervalFlags(recentCmd)
	recentCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	cmd.AddCommand(recentCmd)

	// Time range earthquakes interval command
	timeRangeCmd := &cobra.Command{
		Use:   "time-range",
		Short: "Collect earthquakes by time range at intervals",
		RunE:  a.runIntervalTimeRangeEarthquakes,
	}
	a.addIntervalFlags(timeRangeCmd)
	timeRangeCmd.Flags().String("start", "", "Start time (YYYY-MM-DD)")
	timeRangeCmd.Flags().String("end", "", "End time (YYYY-MM-DD)")
	timeRangeCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	if err := timeRangeCmd.MarkFlagRequired("start"); err != nil {
		panic(fmt.Sprintf("failed to mark start flag as required: %v", err))
	}
	if err := timeRangeCmd.MarkFlagRequired("end"); err != nil {
		panic(fmt.Sprintf("failed to mark end flag as required: %v", err))
	}
	cmd.AddCommand(timeRangeCmd)

	// Magnitude earthquakes interval command
	magnitudeCmd := &cobra.Command{
		Use:   "magnitude",
		Short: "Collect earthquakes by magnitude range at intervals",
		RunE:  a.runIntervalMagnitudeEarthquakes,
	}
	a.addIntervalFlags(magnitudeCmd)
	magnitudeCmd.Flags().Float64("min", 0.0, "Minimum magnitude")
	magnitudeCmd.Flags().Float64("max", 10.0, "Maximum magnitude")
	magnitudeCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	cmd.AddCommand(magnitudeCmd)

	// Significant earthquakes interval command
	significantCmd := &cobra.Command{
		Use:   "significant",
		Short: "Collect significant earthquakes at intervals",
		RunE:  a.runIntervalSignificantEarthquakes,
	}
	a.addIntervalFlags(significantCmd)
	significantCmd.Flags().String("start", "", "Start time (YYYY-MM-DD)")
	significantCmd.Flags().String("end", "", "End time (YYYY-MM-DD)")
	significantCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	cmd.AddCommand(significantCmd)

	// Region earthquakes interval command
	regionCmd := &cobra.Command{
		Use:   "region",
		Short: "Collect earthquakes by region at intervals",
		RunE:  a.runIntervalRegionEarthquakes,
	}
	a.addIntervalFlags(regionCmd)
	regionCmd.Flags().Float64("min-lat", -90.0, "Minimum latitude")
	regionCmd.Flags().Float64("max-lat", 90.0, "Maximum latitude")
	regionCmd.Flags().Float64("min-lon", -180.0, "Minimum longitude")
	regionCmd.Flags().Float64("max-lon", 180.0, "Maximum longitude")
	regionCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	cmd.AddCommand(regionCmd)

	// Country earthquakes interval command
	countryCmd := &cobra.Command{
		Use:   "country",
		Short: "Collect earthquakes by country at intervals",
		RunE:  a.runIntervalCountryEarthquakes,
	}
	a.addIntervalFlags(countryCmd)
	countryCmd.Flags().String("country", "", "Country name")
	countryCmd.Flags().IntP("limit", "l", 1000, "Limit number of records")
	if err := countryCmd.MarkFlagRequired("country"); err != nil {
		panic(fmt.Sprintf("failed to mark country flag as required: %v", err))
	}
	cmd.AddCommand(countryCmd)

	return cmd
}

// newIntervalFaultsCmd creates the interval faults command
func (a *App) newIntervalFaultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faults",
		Short: "Run fault collection at intervals",
		Long:  `Execute fault collection commands at specified intervals.`,
	}

	// Collect faults interval command
	collectCmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect fault data at intervals",
		RunE:  a.runIntervalCollectFaults,
	}
	a.addIntervalFlags(collectCmd)
	cmd.AddCommand(collectCmd)

	// Update faults interval command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update fault data at intervals",
		RunE:  a.runIntervalUpdateFaults,
	}
	a.addIntervalFlags(updateCmd)
	cmd.AddCommand(updateCmd)

	return cmd
}

// newIntervalCustomCmd creates the interval custom command
func (a *App) newIntervalCustomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom",
		Short: "Run custom command combinations at intervals",
		Long:  `Execute custom command combinations at specified intervals.`,
		RunE:  a.runIntervalCustom,
	}

	a.addIntervalFlags(cmd)
	cmd.Flags().StringSlice("commands", []string{}, "Comma-separated list of commands to execute")

	return cmd
}

// addIntervalFlags adds common interval flags to a command
func (a *App) addIntervalFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("interval", "i", "1h", "Time interval (e.g., '5m', '1h', '24h')")
	cmd.Flags().String("max-runtime", "", "Maximum total runtime (e.g., '24h', '7d')")
	cmd.Flags().Int("max-executions", 0, "Maximum number of executions")
	cmd.Flags().String("backoff", "exponential", "Backoff strategy ('none', 'linear', 'exponential')")
	cmd.Flags().String("max-backoff", "30m", "Maximum backoff duration")
	cmd.Flags().Bool("continue-on-error", true, "Continue running on individual command failures")
	cmd.Flags().Bool("skip-empty", false, "Skip execution if no new data is found")
	cmd.Flags().String("health-check-interval", "5m", "Health check interval")
	cmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
	cmd.Flags().String("pid-file", "", "PID file location")
	cmd.Flags().String("log-file", "", "Log file location for daemon mode")
}

// runIntervalRecentEarthquakes runs recent earthquakes collection at intervals
func (a *App) runIntervalRecentEarthquakes(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	// Build command arguments
	cmdArgs := []string{"earthquakes", "recent"}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
	}

	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalTimeRangeEarthquakes runs time range earthquakes collection at intervals
func (a *App) runIntervalTimeRangeEarthquakes(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	// Build command arguments
	cmdArgs := []string{"earthquakes", "time-range"}
	if start, _ := cmd.Flags().GetString("start"); start != "" {
		cmdArgs = append(cmdArgs, "--start", start)
	}
	if end, _ := cmd.Flags().GetString("end"); end != "" {
		cmdArgs = append(cmdArgs, "--end", end)
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
	}

	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalMagnitudeEarthquakes runs magnitude earthquakes collection at intervals
func (a *App) runIntervalMagnitudeEarthquakes(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	// Build command arguments
	cmdArgs := []string{"earthquakes", "magnitude"}
	if min, _ := cmd.Flags().GetFloat64("min"); min > 0 {
		cmdArgs = append(cmdArgs, "--min", fmt.Sprintf("%f", min))
	}
	if max, _ := cmd.Flags().GetFloat64("max"); max < 10.0 {
		cmdArgs = append(cmdArgs, "--max", fmt.Sprintf("%f", max))
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
	}

	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalSignificantEarthquakes runs significant earthquakes collection at intervals
func (a *App) runIntervalSignificantEarthquakes(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	// Build command arguments
	cmdArgs := []string{"earthquakes", "significant"}
	if start, _ := cmd.Flags().GetString("start"); start != "" {
		cmdArgs = append(cmdArgs, "--start", start)
	}
	if end, _ := cmd.Flags().GetString("end"); end != "" {
		cmdArgs = append(cmdArgs, "--end", end)
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
	}

	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalRegionEarthquakes runs region earthquakes collection at intervals
func (a *App) runIntervalRegionEarthquakes(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	// Build command arguments
	cmdArgs := []string{"earthquakes", "region"}
	if minLat, _ := cmd.Flags().GetFloat64("min-lat"); minLat > -90.0 {
		cmdArgs = append(cmdArgs, "--min-lat", fmt.Sprintf("%f", minLat))
	}
	if maxLat, _ := cmd.Flags().GetFloat64("max-lat"); maxLat < 90.0 {
		cmdArgs = append(cmdArgs, "--max-lat", fmt.Sprintf("%f", maxLat))
	}
	if minLon, _ := cmd.Flags().GetFloat64("min-lon"); minLon > -180.0 {
		cmdArgs = append(cmdArgs, "--min-lon", fmt.Sprintf("%f", minLon))
	}
	if maxLon, _ := cmd.Flags().GetFloat64("max-lon"); maxLon < 180.0 {
		cmdArgs = append(cmdArgs, "--max-lon", fmt.Sprintf("%f", maxLon))
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
	}

	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalCountryEarthquakes runs country earthquakes collection at intervals
func (a *App) runIntervalCountryEarthquakes(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	// Build command arguments
	cmdArgs := []string{"earthquakes", "country"}
	if country, _ := cmd.Flags().GetString("country"); country != "" {
		cmdArgs = append(cmdArgs, "--country", country)
	}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
	}

	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalCollectFaults runs fault collection at intervals
func (a *App) runIntervalCollectFaults(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)
	cmdArgs := []string{"faults", "collect"}
	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalUpdateFaults runs fault updates at intervals
func (a *App) runIntervalUpdateFaults(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)
	cmdArgs := []string{"faults", "update"}
	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalCustom runs custom command combinations at intervals
func (a *App) runIntervalCustom(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	commands, _ := cmd.Flags().GetStringSlice("commands")
	if len(commands) == 0 {
		return fmt.Errorf("no commands specified")
	}

	// For custom commands, we'll execute them sequentially
	// This is a simplified implementation - in a full implementation,
	// you might want to support parallel execution or more complex workflows
	for _, command := range commands {
		cmdArgs := strings.Split(command, " ")
		if err := a.runIntervalCommand(cmd, intervalConfig, cmdArgs); err != nil {
			return fmt.Errorf("custom command failed: %w", err)
		}
	}

	return nil
}

// buildIntervalConfig builds the interval configuration from command flags
func (a *App) buildIntervalConfig(cmd *cobra.Command) *config.IntervalConfig {
	intervalStr, _ := cmd.Flags().GetString("interval")
	interval, _ := time.ParseDuration(intervalStr)
	if interval == 0 {
		interval = a.cfg.Interval.DefaultInterval
	}

	maxRuntimeStr, _ := cmd.Flags().GetString("max-runtime")
	maxRuntime, _ := time.ParseDuration(maxRuntimeStr)

	maxExecutions, _ := cmd.Flags().GetInt("max-executions")
	if maxExecutions == 0 {
		maxExecutions = a.cfg.Interval.MaxExecutions
	}

	backoffStrategy, _ := cmd.Flags().GetString("backoff")
	maxBackoffStr, _ := cmd.Flags().GetString("max-backoff")
	maxBackoff, _ := time.ParseDuration(maxBackoffStr)
	if maxBackoff == 0 {
		maxBackoff = a.cfg.Interval.MaxBackoff
	}

	continueOnError, _ := cmd.Flags().GetBool("continue-on-error")
	skipEmpty, _ := cmd.Flags().GetBool("skip-empty")

	healthCheckIntervalStr, _ := cmd.Flags().GetString("health-check-interval")
	healthCheckInterval, _ := time.ParseDuration(healthCheckIntervalStr)
	if healthCheckInterval == 0 {
		healthCheckInterval = a.cfg.Interval.HealthCheckInterval
	}

	daemonMode, _ := cmd.Flags().GetBool("daemon")
	pidFile, _ := cmd.Flags().GetString("pid-file")
	if pidFile == "" {
		pidFile = a.cfg.Interval.PIDFile
	}

	logFile, _ := cmd.Flags().GetString("log-file")
	if logFile == "" {
		logFile = a.cfg.Interval.LogFile
	}

	return &config.IntervalConfig{
		DefaultInterval:     interval,
		MaxRuntime:          maxRuntime,
		MaxExecutions:       maxExecutions,
		BackoffStrategy:     backoffStrategy,
		MaxBackoff:          maxBackoff,
		ContinueOnError:     continueOnError,
		SkipEmpty:           skipEmpty,
		HealthCheckInterval: healthCheckInterval,
		DaemonMode:          daemonMode,
		PIDFile:             pidFile,
		LogFile:             logFile,
	}
}

// runIntervalCommand runs a command at intervals using the scheduler
func (a *App) runIntervalCommand(cmd *cobra.Command, intervalConfig *config.IntervalConfig, cmdArgs []string) error {
	// Create logger
	logger := log.New(os.Stdout, "[INTERVAL] ", log.LstdFlags)

	// Create internal command executor function
	internalExecutor := func(ctx context.Context, args []string) error {
		// Create a new command with the arguments
		execCmd := exec.CommandContext(ctx, os.Args[0], args...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin
		return execCmd.Run()
	}

	// Create scheduler with internal executor
	scheduler := sched.NewIntervalScheduler(intervalConfig, logger)

	// Replace the executor with our internal one
	executor := sched.NewCommandExecutorWithFunction(logger, internalExecutor)
	scheduler.SetExecutor(executor)

	// Set up backoff strategy
	switch intervalConfig.BackoffStrategy {
	case "none":
		executor.SetBackoffStrategy(&sched.NoBackoff{})
	case "linear":
		executor.SetBackoffStrategy(sched.NewLinearBackoff(5 * time.Second))
	case "exponential":
		executor.SetBackoffStrategy(sched.NewExponentialBackoff(5*time.Second, intervalConfig.MaxBackoff))
	default:
		executor.SetBackoffStrategy(sched.NewExponentialBackoff(5*time.Second, intervalConfig.MaxBackoff))
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Printf("Received shutdown signal, stopping scheduler...")
		cancel()
		scheduler.Stop()
	}()

	// Start the scheduler
	if intervalConfig.DaemonMode {
		logger.Printf("Starting interval scheduler in daemon mode")
		return scheduler.StartDaemon(ctx, "quakewatch-scraper", cmdArgs)
	} else {
		logger.Printf("Starting interval scheduler")
		return scheduler.Start(ctx, "quakewatch-scraper", cmdArgs)
	}
}

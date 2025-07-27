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

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"quakewatch-scraper/internal/api"
	"quakewatch-scraper/internal/collector"
	"quakewatch-scraper/internal/config"
	sched "quakewatch-scraper/internal/scheduler"
	"quakewatch-scraper/internal/storage"
	"quakewatch-scraper/internal/utils"
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

// getOutputDir returns the output directory, respecting both configuration and command flags
func (a *App) getOutputDir(cmd *cobra.Command) string {
	// Check for custom output directory flag
	outputDir, _ := cmd.Flags().GetString("output-dir")
	if outputDir != "" {
		return outputDir
	}
	// Use configured output directory
	return a.getOutputDir(cmd)
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
		// Skip configuration loading for version and config commands
		if cmd.Name() == "version" || cmd.Name() == "config" {
			return nil
		}

		// Load configuration for all commands
		configPath, _ := cmd.Flags().GetString("config")

		// Load configuration for all commands
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			// If config loading fails, use default configuration
			app.cfg = config.DefaultConfig()
		} else {
			app.cfg = cfg
		}

		return nil
	}

	app.setupCommands()
	app.setupFlags()

	// Set the banner function for when no command is provided
	app.rootCmd.Run = app.showBanner

	// Create and set up the help command with flags
	helpCmd := &cobra.Command{
		Use:   "help",
		Short: "Show comprehensive help and examples",
		Long:  `Display comprehensive help information with examples, organized by category.`,
		Run:   app.runHelp,
	}

	// Add flags to the help command - using different names to avoid conflicts
	helpCmd.Flags().String("help-category", "all", "Help category (all, earthquakes, faults, interval, db, utils, examples)")
	helpCmd.Flags().Bool("help-examples", false, "Show usage examples")
	helpCmd.Flags().Bool("help-quick", false, "Show quick reference")

	// Override the default help command with our custom version
	app.rootCmd.SetHelpCommand(helpCmd)

	// Set a custom help template for better formatting
	app.rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`)

	return app
}

func (a *App) setupCommands() {
	// Add earthquake commands
	a.rootCmd.AddCommand(a.newEarthquakeCmd())

	// Add fault commands
	a.rootCmd.AddCommand(a.newFaultCmd())

	// Add interval commands
	a.rootCmd.AddCommand(a.newIntervalCmd())

	// Add database commands
	a.rootCmd.AddCommand(a.newDatabaseCmd())

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
	recentCmd.Flags().Bool("smart", false, "Use smart collection to avoid duplicates")
	recentCmd.Flags().String("storage", "json", "Storage backend (json, postgresql)")
	recentCmd.Flags().Int("hours-back", 1, "Number of hours to look back")
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
	countryCmd.Flags().String("start", "", "Start time (YYYY-MM-DD) (default: 30 days ago)")
	countryCmd.Flags().String("end", "", "End time (YYYY-MM-DD) (default: today)")
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
	smart, _ := cmd.Flags().GetBool("smart")

	// Initialize storage - for now just use JSON storage
	outputDir := a.getOutputDir(cmd)
	jsonStorage := storage.NewJSONStorage(outputDir)

	// Initialize API client with enhanced features
	metrics := utils.NewCollectionMetrics()
	logger, err := utils.NewStructuredLogger(a.cfg.Logging.Level, a.cfg.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout, metrics, logger)
	collector := collector.NewEarthquakeCollector(usgsClient, jsonStorage)

	ctx := context.Background()

	if smart {
		// Use smart collection to avoid duplicates
		if err := collector.CollectRecentEarthquakesSmart(ctx, jsonStorage); err != nil {
			return fmt.Errorf("smart collection failed: %w", err)
		}
	} else {
		// Use regular collection
		if err := collector.CollectRecent(ctx, limit, filename); err != nil {
			return fmt.Errorf("collection failed: %w", err)
		}
	}

	return nil
}

func (a *App) runTimeRangeEarthquakes(cmd *cobra.Command, args []string) error {
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}

	// Initialize storage
	outputDir := a.getOutputDir(cmd)
	jsonStorage := storage.NewJSONStorage(outputDir)

	// Initialize API client with enhanced features
	metrics := utils.NewCollectionMetrics()
	logger, err := utils.NewStructuredLogger(a.cfg.Logging.Level, a.cfg.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout, metrics, logger)
	collector := collector.NewEarthquakeCollector(usgsClient, jsonStorage)

	ctx := context.Background()
	if err := collector.CollectByTimeRange(ctx, startTime, endTime, limit, filename); err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}

	return nil
}

func (a *App) runMagnitudeEarthquakes(cmd *cobra.Command, args []string) error {
	minMag, _ := cmd.Flags().GetFloat64("min")
	maxMag, _ := cmd.Flags().GetFloat64("max")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	// Initialize storage
	outputDir := a.getOutputDir(cmd)
	jsonStorage := storage.NewJSONStorage(outputDir)

	// Initialize API client with enhanced features
	metrics := utils.NewCollectionMetrics()
	logger, err := utils.NewStructuredLogger(a.cfg.Logging.Level, a.cfg.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout, metrics, logger)
	collector := collector.NewEarthquakeCollector(usgsClient, jsonStorage)

	ctx := context.Background()
	if err := collector.CollectByMagnitude(ctx, minMag, maxMag, limit, filename); err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}

	return nil
}

func (a *App) runSignificantEarthquakes(cmd *cobra.Command, args []string) error {
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}

	// Initialize storage
	outputDir := a.getOutputDir(cmd)
	jsonStorage := storage.NewJSONStorage(outputDir)

	// Initialize API client with enhanced features
	metrics := utils.NewCollectionMetrics()
	logger, err := utils.NewStructuredLogger(a.cfg.Logging.Level, a.cfg.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout, metrics, logger)
	collector := collector.NewEarthquakeCollector(usgsClient, jsonStorage)

	ctx := context.Background()
	if err := collector.CollectSignificant(ctx, startTime, endTime, limit, filename); err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}

	return nil
}

func (a *App) runRegionEarthquakes(cmd *cobra.Command, args []string) error {
	minLat, _ := cmd.Flags().GetFloat64("min-lat")
	maxLat, _ := cmd.Flags().GetFloat64("max-lat")
	minLon, _ := cmd.Flags().GetFloat64("min-lon")
	maxLon, _ := cmd.Flags().GetFloat64("max-lon")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")
	stdout, _ := cmd.Flags().GetBool("stdout")

	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize storage
	outputDir := a.getOutputDir(cmd)
	jsonStorage := storage.NewJSONStorage(outputDir)

	// Initialize API client with enhanced features
	metrics := utils.NewCollectionMetrics()
	logger, err := utils.NewStructuredLogger(a.cfg.Logging.Level, a.cfg.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout, metrics, logger)
	collector := collector.NewEarthquakeCollector(usgsClient, jsonStorage)
	ctx := context.Background()

	if stdout {
		earthquakes, err := collector.CollectByRegionData(ctx, minLat, maxLat, minLon, maxLon, limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectByRegion(ctx, minLat, maxLat, minLon, maxLon, limit, filename)
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

	// Set default values if start/end are not provided
	var startTime, endTime time.Time
	var err error

	if startStr == "" {
		// Default to 30 days ago
		startTime = time.Now().AddDate(0, 0, -30)
	} else {
		startTime, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return fmt.Errorf("invalid start time format: %w", err)
		}
	}

	if endStr == "" {
		// Default to today
		endTime = time.Now()
	} else {
		endTime, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return fmt.Errorf("invalid end time format: %w", err)
		}
	}

	if limit == 0 {
		limit = a.cfg.Collection.DefaultLimit
	}
	if limit > a.cfg.Collection.MaxLimit {
		limit = a.cfg.Collection.MaxLimit
	}

	// Initialize storage
	outputDir := a.getOutputDir(cmd)
	jsonStorage := storage.NewJSONStorage(outputDir)

	// Initialize API client with enhanced features
	metrics := utils.NewCollectionMetrics()
	logger, err := utils.NewStructuredLogger(a.cfg.Logging.Level, a.cfg.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, a.cfg.API.USGS.Timeout, metrics, logger)
	collector := collector.NewEarthquakeCollector(usgsClient, jsonStorage)
	ctx := context.Background()

	if stdout {
		earthquakes, err := collector.CollectByCountryData(ctx, country, startTime, endTime, minMag, maxMag, limit)
		if err != nil {
			return err
		}
		return a.outputToStdout(earthquakes)
	}

	return collector.CollectByCountry(ctx, country, startTime, endTime, minMag, maxMag, limit, filename)
}

func (a *App) runCollectFaults(cmd *cobra.Command, args []string) error {
	storageType, _ := cmd.Flags().GetString("storage")
	var store storage.Storage
	if storageType == "postgresql" {
		dbConfig := config.NewDatabaseConfig()
		pgStore, err := storage.NewPostgreSQLStorage(dbConfig)
		if err != nil {
			return fmt.Errorf("failed to initialize PostgreSQL storage: %w", err)
		}
		store = pgStore
		defer store.Close()
	} else {
		jsonStore := storage.NewJSONStorage(a.getOutputDir(cmd))
		store = jsonStore
	}

	emscClient := api.NewEMSCClient(a.cfg.API.EMSC.BaseURL, a.cfg.API.EMSC.Timeout)
	collector := collector.NewFaultCollector(emscClient, store)
	ctx := context.Background()

	return collector.CollectFaults(ctx)
}

func (a *App) runUpdateFaults(cmd *cobra.Command, args []string) error {
	storageType, _ := cmd.Flags().GetString("storage")
	maxRetries, _ := cmd.Flags().GetInt("retries")
	retryDelay, _ := cmd.Flags().GetDuration("retry-delay")
	var store storage.Storage
	if storageType == "postgresql" {
		dbConfig := config.NewDatabaseConfig()
		pgStore, err := storage.NewPostgreSQLStorage(dbConfig)
		if err != nil {
			return fmt.Errorf("failed to initialize PostgreSQL storage: %w", err)
		}
		store = pgStore
		defer store.Close()
	} else {
		jsonStore := storage.NewJSONStorage(a.getOutputDir(cmd))
		store = jsonStore
	}

	emscClient := api.NewEMSCClient(a.cfg.API.EMSC.BaseURL, a.cfg.API.EMSC.Timeout)
	collector := collector.NewFaultCollector(emscClient, store)
	ctx := context.Background()

	return collector.UpdateFaults(ctx, maxRetries, retryDelay)
}

func (a *App) runValidate(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")
	file, _ := cmd.Flags().GetString("file")

	store := storage.NewJSONStorage(a.getOutputDir(cmd))
	ctx := context.Background()

	if file != "" {
		// Validate specific file
		stats, err := store.GetFileStats(ctx, dataType)
		if err != nil {
			return fmt.Errorf("failed to validate file: %w", err)
		}
		fmt.Printf("Stats for %s: %+v\n", file, stats)
		return nil
	}

	if dataType == "all" {
		fmt.Println("Validating all data files:")

		// Validate earthquake files
		earthquakeFiles, err := store.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("Error listing earthquake files: %v\n", err)
		} else {
			fmt.Println("Earthquakes:")
			for _, filename := range earthquakeFiles {
				stats, err := store.GetFileStats(ctx, "earthquakes")
				if err != nil {
					fmt.Printf("  ✗ %s: %v\n", filename, err)
				} else {
					fmt.Printf("  ✓ %s: %+v\n", filename, stats)
				}
			}
		}

		// Validate fault files
		faultFiles, err := store.ListFiles("faults")
		if err != nil {
			fmt.Printf("Error listing fault files: %v\n", err)
		} else {
			fmt.Println("Faults:")
			for _, filename := range faultFiles {
				stats, err := store.GetFileStats(ctx, "faults")
				if err != nil {
					fmt.Printf("  ✗ %s: %v\n", filename, err)
				} else {
					fmt.Printf("  ✓ %s: %+v\n", filename, stats)
				}
			}
		}

		return nil
	}

	// Validate specific type
	files, err := store.ListFiles(dataType)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	for _, filename := range files {
		stats, err := store.GetFileStats(ctx, dataType)
		if err != nil {
			fmt.Printf("Failed to validate %s: %v\n", filename, err)
		} else {
			fmt.Printf("Validated %s: %+v\n", filename, stats)
		}
	}

	return nil
}

func (a *App) runStats(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")
	file, _ := cmd.Flags().GetString("file")

	store := storage.NewJSONStorage(a.getOutputDir(cmd))
	ctx := context.Background()

	if file != "" {
		// Show stats for specific file
		stats, err := store.GetFileStats(ctx, dataType)
		if err != nil {
			return fmt.Errorf("failed to get file stats: %w", err)
		}
		fmt.Printf("Stats for %s: %+v\n", file, stats)
		return nil
	}

	if dataType == "all" {
		fmt.Println("Statistics for all data:")

		// Show earthquake stats
		earthquakeFiles, err := store.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("  Error listing earthquake files: %v\n", err)
		} else {
			fmt.Printf("  Earthquake files: %d\n", len(earthquakeFiles))
			totalEarthquakeRecords := 0
			for _, filename := range earthquakeFiles {
				stats, err := store.GetFileStats(ctx, "earthquakes")
				if err != nil {
					fmt.Printf("    Failed to get stats for %s: %v\n", filename, err)
				} else {
					fmt.Printf("    %s: %+v\n", filename, stats)
					if count, ok := stats["count"].(int); ok {
						totalEarthquakeRecords += count
					}
				}
			}
			fmt.Printf("  Total earthquake records: %d\n", totalEarthquakeRecords)
		}

		// Show fault stats
		faultFiles, err := store.ListFiles("faults")
		if err != nil {
			fmt.Printf("  Error listing fault files: %v\n", err)
		} else {
			fmt.Printf("  Fault files: %d\n", len(faultFiles))
			totalFaultRecords := 0
			for _, filename := range faultFiles {
				stats, err := store.GetFileStats(ctx, "faults")
				if err != nil {
					fmt.Printf("    Failed to get stats for %s: %v\n", filename, err)
				} else {
					fmt.Printf("    %s: %+v\n", filename, stats)
					if count, ok := stats["count"].(int); ok {
						totalFaultRecords += count
					}
				}
			}
			fmt.Printf("  Total fault records: %d\n", totalFaultRecords)
		}

		return nil
	}

	// Show stats for specific type
	files, err := store.ListFiles(dataType)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	totalRecords := 0
	for _, filename := range files {
		stats, err := store.GetFileStats(ctx, dataType)
		if err != nil {
			fmt.Printf("  Failed to get stats for %s: %v\n", filename, err)
		} else {
			fmt.Printf("  %s: %+v\n", filename, stats)
			if count, ok := stats["count"].(int); ok {
				totalRecords += count
			}
		}
	}
	fmt.Printf("Total records for %s: %d\n", dataType, totalRecords)

	return nil
}

func (a *App) runList(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")

	store := storage.NewJSONStorage(a.getOutputDir(cmd))

	if dataType == "all" {
		fmt.Println("Available data files:")
		fmt.Println("Earthquakes:")
		earthquakeFiles, err := store.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("  Error listing earthquake files: %v\n", err)
		} else {
			for _, file := range earthquakeFiles {
				fmt.Printf("  %s\n", file)
			}
		}

		fmt.Println("Faults:")
		faultFiles, err := store.ListFiles("faults")
		if err != nil {
			fmt.Printf("  Error listing fault files: %v\n", err)
		} else {
			for _, file := range faultFiles {
				fmt.Printf("  %s\n", file)
			}
		}
	} else {
		files, err := store.ListFiles(dataType)
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

	store := storage.NewJSONStorage(a.getOutputDir(cmd))
	ctx := context.Background()

	if dryRun {
		fmt.Println("DRY RUN - Files that would be deleted:")

		if dataType == "all" || dataType == "earthquakes" {
			earthquakeFiles, err := store.ListFiles("earthquakes")
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
			faultFiles, err := store.ListFiles("faults")
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
		earthquakeFiles, err := store.ListFiles("earthquakes")
		if err != nil {
			fmt.Printf("  Error listing earthquake files: %v\n", err)
		} else {
			fmt.Printf("  Earthquake files: %d\n", len(earthquakeFiles))
			totalFiles += len(earthquakeFiles)
		}
	}

	if dataType == "all" || dataType == "faults" {
		faultFiles, err := store.ListFiles("faults")
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
		if err := store.PurgeAll(ctx); err != nil {
			return fmt.Errorf("failed to purge all files: %w", err)
		}
		fmt.Printf("Successfully deleted %d files.\n", totalFiles)
	} else {
		if err := store.PurgeByType(ctx, dataType); err != nil {
			return fmt.Errorf("failed to purge %s files: %w", dataType, err)
		}
		fmt.Printf("Successfully deleted %s files.\n", dataType)
	}

	return nil
}

func (a *App) runHealth(cmd *cobra.Command, args []string) error {
	fmt.Println("System Health Check:")

	// Initialize enhanced features
	metrics := utils.NewCollectionMetrics()
	logger, err := utils.NewStructuredLogger(a.cfg.Logging.Level, a.cfg.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Check USGS API
	usgsClient := api.NewUSGSClient(a.cfg.API.USGS.BaseURL, 10*time.Second, metrics, logger)
	ctx := context.Background()
	_, err = usgsClient.GetRecentEarthquakes(ctx, 1)
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
	store := storage.NewJSONStorage(a.getOutputDir(cmd))
	_, err = store.ListFiles("earthquakes")
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
	fmt.Println("QuakeWatch Scraper v1.2.2")
	fmt.Println("Go version: 1.24")
	fmt.Println("Build date: " + time.Now().Format("2006-01-02"))
}

func (a *App) runConfig(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")

	fmt.Println("QuakeWatch Scraper Configuration Setup")
	fmt.Println("=====================================")

	// Force interactive configuration creation
	cfg, err := config.CreateInteractiveConfig(configPath)
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
	fmt.Println("║  Version: 1.2.2                                              ║")
	fmt.Println("║  Built with Go                                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Show a more user-friendly help summary
	fmt.Println("🚀 QUICK START")
	fmt.Println("==============")
	fmt.Println("  • Collect recent earthquakes: quakewatch-scraper earthquakes recent")
	fmt.Println("  • Collect fault data: quakewatch-scraper faults collect")
	fmt.Println("  • Check system health: quakewatch-scraper health")
	fmt.Println()

	fmt.Println("📚 COMMAND CATEGORIES")
	fmt.Println("====================")
	fmt.Println("  🌍 earthquakes    - Collect earthquake data from USGS")
	fmt.Println("  🏔️  faults        - Collect fault data from EMSC")
	fmt.Println("  ⏰ interval       - Run commands at regular intervals")
	fmt.Println("  🗄️  db            - Manage database operations")
	fmt.Println("  🛠️  utilities     - Data management and system tools")
	fmt.Println()

	fmt.Println("🔍 GETTING HELP")
	fmt.Println("===============")
	fmt.Println("  • Main help: quakewatch-scraper help")
	fmt.Println("  • Category help: quakewatch-scraper help --help-category earthquakes")
	fmt.Println("  • Quick reference: quakewatch-scraper help --help-quick")
	fmt.Println("  • Examples: quakewatch-scraper help --help-examples")
	fmt.Println("  • Command help: quakewatch-scraper [command] --help")
	fmt.Println()

	fmt.Println("💡 For detailed help and examples, run: quakewatch-scraper help")
	fmt.Println()
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
	cmd.Flags().String("storage", "json", "Storage backend (json, postgresql)")
	cmd.Flags().Bool("smart", false, "Use smart collection to avoid duplicates")
	cmd.Flags().Int("hours-back", 1, "Number of hours to look back")
}

// runIntervalRecentEarthquakes runs recent earthquakes collection at intervals
func (a *App) runIntervalRecentEarthquakes(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)

	// Build command arguments
	cmdArgs := []string{"earthquakes", "recent"}
	if limit, _ := cmd.Flags().GetInt("limit"); limit > 0 {
		cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", limit))
	}
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
	}
	if smart, _ := cmd.Flags().GetBool("smart"); smart {
		cmdArgs = append(cmdArgs, "--smart")
	}
	if hoursBack, _ := cmd.Flags().GetInt("hours-back"); hoursBack > 0 {
		cmdArgs = append(cmdArgs, "--hours-back", fmt.Sprintf("%d", hoursBack))
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
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
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
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
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
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
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
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
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
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
	}

	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalCollectFaults runs fault collection at intervals
func (a *App) runIntervalCollectFaults(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)
	cmdArgs := []string{"faults", "collect"}
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
	}
	return a.runIntervalCommand(cmd, intervalConfig, cmdArgs)
}

// runIntervalUpdateFaults runs fault updates at intervals
func (a *App) runIntervalUpdateFaults(cmd *cobra.Command, args []string) error {
	intervalConfig := a.buildIntervalConfig(cmd)
	cmdArgs := []string{"faults", "update"}
	if storage, _ := cmd.Flags().GetString("storage"); storage != "" {
		cmdArgs = append(cmdArgs, "--storage", storage)
	}
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

// newDatabaseCmd creates the database command
func (a *App) newDatabaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
		Long:  `Manage database initialization, migrations, and status`,
	}

	// Database init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize database and run migrations",
		Long:  `Initialize the database by creating the database if it doesn't exist and running all migrations`,
		RunE:  a.runDatabaseInit,
	}
	initCmd.Flags().Bool("force", false, "Force re-initialization even if database exists")
	cmd.AddCommand(initCmd)

	// Database migrate command
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
		Long:  `Manage database migrations`,
	}

	// Migrate up command
	migrateUpCmd := &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		Long:  `Apply all pending database migrations`,
		RunE:  a.runDatabaseMigrateUp,
	}
	migrateCmd.AddCommand(migrateUpCmd)

	// Migrate down command
	migrateDownCmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback all migrations",
		Long:  `Rollback all database migrations`,
		RunE:  a.runDatabaseMigrateDown,
	}
	migrateCmd.AddCommand(migrateDownCmd)

	// Migrate to version command
	migrateToCmd := &cobra.Command{
		Use:   "to",
		Short: "Migrate to specific version",
		Long:  `Migrate database to a specific version`,
		RunE:  a.runDatabaseMigrateTo,
	}
	migrateToCmd.Flags().Uint("version", 0, "Target migration version")
	if err := migrateToCmd.MarkFlagRequired("version"); err != nil {
		panic(fmt.Sprintf("failed to mark version flag as required: %v", err))
	}
	migrateCmd.AddCommand(migrateToCmd)

	// Force version command
	forceCmd := &cobra.Command{
		Use:   "force",
		Short: "Force migration version",
		Long:  `Force the database migration version (use with caution)`,
		RunE:  a.runDatabaseForceVersion,
	}
	forceCmd.Flags().Uint("version", 0, "Target migration version")
	if err := forceCmd.MarkFlagRequired("version"); err != nil {
		panic(fmt.Sprintf("failed to mark version flag as required: %v", err))
	}
	migrateCmd.AddCommand(forceCmd)

	cmd.AddCommand(migrateCmd)

	// Database status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show database status and migration information",
		Long:  `Display current database status, connection info, and migration version`,
		RunE:  a.runDatabaseStatus,
	}
	cmd.AddCommand(statusCmd)

	return cmd
}

// runDatabaseInit initializes the database
func (a *App) runDatabaseInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing database...")

	// Check if database configuration is available
	if a.cfg == nil {
		return fmt.Errorf("configuration not found. Please check your config file")
	}

	force, _ := cmd.Flags().GetBool("force")

	// Create migration manager
	migrationManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer migrationManager.Close()

	// Check if database exists and is accessible
	fmt.Println("Checking database connection...")
	if err := migrationManager.TestConnection(); err != nil {
		if force {
			fmt.Println("Database connection failed, but continuing with force flag...")
		} else {
			return fmt.Errorf("database connection failed: %w", err)
		}
	} else {
		fmt.Println("Database connection successful")
	}

	// Run migrations
	fmt.Println("Running database migrations...")
	if err := migrationManager.MigrateUp(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create a new migration manager to get version (since the previous one closed the connection)
	versionManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create version manager: %w", err)
	}
	defer versionManager.Close()

	// Get current version
	version, dirty, err := versionManager.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	fmt.Printf("Database initialized successfully!\n")
	fmt.Printf("Current migration version: %d\n", version)
	if dirty {
		fmt.Println("Warning: Database is in dirty state")
	}

	return nil
}

// runDatabaseMigrateUp runs all pending migrations
func (a *App) runDatabaseMigrateUp(cmd *cobra.Command, args []string) error {
	fmt.Println("Running database migrations...")

	// Check if database configuration is available
	if a.cfg == nil {
		return fmt.Errorf("configuration not found. Please check your config file")
	}

	// Create migration manager
	migrationManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer migrationManager.Close()

	// Run migrations
	if err := migrationManager.MigrateUp(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create a new migration manager to get version (since the previous one closed the connection)
	versionManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create version manager: %w", err)
	}
	defer versionManager.Close()

	// Get current version
	version, dirty, err := versionManager.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	fmt.Printf("Migrations completed successfully!\n")
	fmt.Printf("Current migration version: %d\n", version)
	if dirty {
		fmt.Println("Warning: Database is in dirty state")
	}

	return nil
}

// runDatabaseMigrateDown rolls back all migrations
func (a *App) runDatabaseMigrateDown(cmd *cobra.Command, args []string) error {
	fmt.Println("Rolling back all database migrations...")

	// Check if database configuration is available
	if a.cfg == nil {
		return fmt.Errorf("configuration not found. Please check your config file")
	}

	// Create migration manager
	migrationManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer migrationManager.Close()

	// Run migrations down
	if err := migrationManager.MigrateDown(); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	fmt.Println("All migrations rolled back successfully!")
	return nil
}

// runDatabaseMigrateTo migrates to a specific version
func (a *App) runDatabaseMigrateTo(cmd *cobra.Command, args []string) error {
	version, _ := cmd.Flags().GetUint("version")
	fmt.Printf("Migrating database to version %d...\n", version)

	// Check if database configuration is available
	if a.cfg == nil {
		return fmt.Errorf("configuration not found. Please check your config file")
	}

	// Create migration manager
	migrationManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer migrationManager.Close()

	// Migrate to specific version
	if err := migrationManager.MigrateToVersion(version); err != nil {
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	fmt.Printf("Successfully migrated to version %d!\n", version)
	return nil
}

// runDatabaseForceVersion forces the migration version
func (a *App) runDatabaseForceVersion(cmd *cobra.Command, args []string) error {
	version, _ := cmd.Flags().GetUint("version")
	fmt.Printf("Forcing database migration version to %d...\n", version)

	// Check if database configuration is available
	if a.cfg == nil {
		return fmt.Errorf("configuration not found. Please check your config file")
	}

	// Create migration manager
	migrationManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer migrationManager.Close()

	// Force version
	if err := migrationManager.ForceVersion(version); err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}

	fmt.Printf("Successfully forced migration version to %d!\n", version)
	return nil
}

// runDatabaseStatus shows database status
func (a *App) runDatabaseStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("Database Status")
	fmt.Println("===============")

	// Check if database configuration is available
	if a.cfg == nil {
		fmt.Println("❌ Configuration not found")
		return fmt.Errorf("configuration not found. Please check your config file")
	}

	// Display configuration
	fmt.Printf("Host: %s:%d\n", a.cfg.Database.Host, a.cfg.Database.Port)
	fmt.Printf("Database: %s\n", a.cfg.Database.Database)
	fmt.Printf("User: %s\n", a.cfg.Database.User)
	fmt.Printf("SSL Mode: %s\n", a.cfg.Database.SSLMode)

	// Test connection
	fmt.Println("\nConnection Test:")
	migrationManager, err := storage.NewMigrationManager(&a.cfg.Database)
	if err != nil {
		fmt.Printf("❌ Failed to create migration manager: %v\n", err)
		return nil
	}
	defer migrationManager.Close()

	if err := migrationManager.TestConnection(); err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		return nil
	}
	fmt.Println("✅ Database connection successful")

	// Get migration status
	fmt.Println("\nMigration Status:")
	version, dirty, err := migrationManager.GetVersionWithoutClose()
	if err != nil {
		fmt.Printf("❌ Failed to get migration version: %v\n", err)
		return nil
	}

	fmt.Printf("Current Version: %d\n", version)
	if dirty {
		fmt.Println("⚠️  Database is in dirty state")
	} else {
		fmt.Println("✅ Database is clean")
	}

	// Check if tables exist using a separate connection
	fmt.Println("\nTable Status:")
	tables := []string{"earthquakes", "faults", "collection_logs", "collection_metadata"}

	// Create a separate database connection for table checks
	db, err := sqlx.Connect("postgres", a.cfg.Database.GetDSN())
	if err != nil {
		fmt.Printf("❌ Failed to create database connection for table checks: %v\n", err)
	} else {
		defer db.Close()

		for _, table := range tables {
			var exists bool
			query := `SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)`
			err := db.Get(&exists, query, table)
			if err != nil {
				fmt.Printf("❌ %s: Error checking table - %v\n", table, err)
			} else if exists {
				fmt.Printf("✅ %s: Table exists\n", table)
			} else {
				fmt.Printf("❌ %s: Table does not exist\n", table)
			}
		}
	}

	return nil
}

// getOutputDir returns the output directory, respecting both configuration and command flags

// newHelpCmd creates a custom help command with better organization and examples
func (a *App) newHelpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help",
		Short: "Show comprehensive help and examples",
		Long:  `Display comprehensive help information with examples, organized by category.`,
		Run:   a.runHelp,
	}

	cmd.Flags().String("category", "all", "Help category (all, earthquakes, faults, interval, db, utils, examples)")
	cmd.Flags().Bool("examples", false, "Show usage examples")
	cmd.Flags().Bool("quick", false, "Show quick reference")

	return cmd
}

// runHelp displays comprehensive help information
func (a *App) runHelp(cmd *cobra.Command, args []string) {
	category, _ := cmd.Flags().GetString("help-category")
	showExamples, _ := cmd.Flags().GetBool("help-examples")
	quickRef, _ := cmd.Flags().GetBool("help-quick")

	if quickRef {
		a.showQuickReference()
		return
	}

	if showExamples {
		a.showExamples()
		return
	}

	switch category {
	case "earthquakes":
		a.showEarthquakeHelp()
	case "faults":
		a.showFaultHelp()
	case "interval":
		a.showIntervalHelp()
	case "db":
		a.showDatabaseHelp()
	case "utils":
		a.showUtilityHelp()
	case "examples":
		a.showExamples()
	default:
		a.showComprehensiveHelp()
	}
}

// showComprehensiveHelp displays the main help screen
func (a *App) showComprehensiveHelp() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  🌋 QuakeWatch Scraper Help 🌋              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("📋 QUICK START")
	fmt.Println("==============")
	fmt.Println("  • Collect recent earthquakes: quakewatch-scraper earthquakes recent")
	fmt.Println("  • Collect fault data: quakewatch-scraper faults collect")
	fmt.Println("  • Check system health: quakewatch-scraper health")
	fmt.Println("  • View examples: quakewatch-scraper help --examples")
	fmt.Println()

	fmt.Println("📚 COMMAND CATEGORIES")
	fmt.Println("====================")
	fmt.Println("  🌍 Earthquakes    - Collect earthquake data from USGS")
	fmt.Println("  🏔️  Faults        - Collect fault data from EMSC")
	fmt.Println("  ⏰ Interval       - Run commands at regular intervals")
	fmt.Println("  🗄️  Database       - Manage database operations")
	fmt.Println("  🛠️  Utilities      - Data management and system tools")
	fmt.Println()

	fmt.Println("🔍 GETTING HELP")
	fmt.Println("===============")
	fmt.Println("  • Main help: quakewatch-scraper help")
	fmt.Println("  • Category help: quakewatch-scraper help --help-category earthquakes")
	fmt.Println("  • Quick reference: quakewatch-scraper help --help-quick")
	fmt.Println("  • Examples: quakewatch-scraper help --help-examples")
	fmt.Println("  • Command help: quakewatch-scraper [command] --help")
	fmt.Println()

	fmt.Println("⚙️  GLOBAL OPTIONS")
	fmt.Println("=================")
	fmt.Println("  -c, --config <file>     Configuration file path")
	fmt.Println("  -v, --verbose           Enable verbose logging")
	fmt.Println("  -q, --quiet             Suppress output")
	fmt.Println("  -o, --output-dir <dir>  Output directory for files")
	fmt.Println("  --dry-run               Show what would be done")
	fmt.Println("  --stdout                Output to stdout instead of file")
	fmt.Println()

	fmt.Println("📖 For detailed information about each category, use:")
	fmt.Println("   quakewatch-scraper help --help-category <category>")
	fmt.Println()
	fmt.Println("💡 For practical examples, use:")
	fmt.Println("   quakewatch-scraper help --help-examples")
}

// showQuickReference displays a quick reference guide
func (a *App) showQuickReference() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                🚀 Quick Reference Guide 🚀                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("🌍 EARTHQUAKE COMMANDS")
	fmt.Println("=====================")
	fmt.Println("  recent                    # Last hour earthquakes")
	fmt.Println("  time-range --start --end  # Time range collection")
	fmt.Println("  magnitude --min --max     # Magnitude range")
	fmt.Println("  significant --start --end # Significant earthquakes (M4.5+)")
	fmt.Println("  region --min-lat --max-lat --min-lon --max-lon")
	fmt.Println("  country --country         # Country-specific")
	fmt.Println()

	fmt.Println("🏔️  FAULT COMMANDS")
	fmt.Println("==================")
	fmt.Println("  collect                   # Collect fault data")
	fmt.Println("  update                    # Update with retry logic")
	fmt.Println()

	fmt.Println("⏰ INTERVAL COMMANDS")
	fmt.Println("===================")
	fmt.Println("  earthquakes recent --interval 1h")
	fmt.Println("  faults collect --interval 6h")
	fmt.Println("  custom --commands 'cmd1,cmd2' --interval 30m")
	fmt.Println()

	fmt.Println("🗄️  DATABASE COMMANDS")
	fmt.Println("====================")
	fmt.Println("  init                      # Initialize database")
	fmt.Println("  migrate up                # Run migrations")
	fmt.Println("  migrate down              # Rollback migrations")
	fmt.Println("  status                    # Check status")
	fmt.Println()

	fmt.Println("🛠️  UTILITY COMMANDS")
	fmt.Println("====================")
	fmt.Println("  health                    # System health check")
	fmt.Println("  stats                     # Data statistics")
	fmt.Println("  validate                  # Data validation")
	fmt.Println("  list                      # List data files")
	fmt.Println("  purge                     # Delete data files")
	fmt.Println("  version                   # Version information")
	fmt.Println()

	fmt.Println("💡 COMMON FLAGS")
	fmt.Println("===============")
	fmt.Println("  --limit <number>          # Limit records")
	fmt.Println("  --filename <name>         # Custom filename")
	fmt.Println("  --storage <backend>       # Storage backend")
	fmt.Println("  --smart                   # Smart collection")
	fmt.Println()
}

// showExamples displays practical usage examples
func (a *App) showExamples() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    📝 Usage Examples 📝                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("🌍 EARTHQUAKE EXAMPLES")
	fmt.Println("======================")
	fmt.Println("  # Collect recent earthquakes (last hour)")
	fmt.Println("  quakewatch-scraper earthquakes recent")
	fmt.Println()
	fmt.Println("  # Collect earthquakes from last 24 hours")
	fmt.Println("  quakewatch-scraper earthquakes recent --hours-back 24")
	fmt.Println()
	fmt.Println("  # Collect earthquakes by time range")
	fmt.Println("  quakewatch-scraper earthquakes time-range \\")
	fmt.Println("    --start '2024-01-01' \\")
	fmt.Println("    --end '2024-01-31' \\")
	fmt.Println("    --limit 500")
	fmt.Println()
	fmt.Println("  # Collect significant earthquakes")
	fmt.Println("  quakewatch-scraper earthquakes significant \\")
	fmt.Println("    --start '2024-01-01' \\")
	fmt.Println("    --end '2024-01-31'")
	fmt.Println()
	fmt.Println("  # Collect earthquakes by magnitude")
	fmt.Println("  quakewatch-scraper earthquakes magnitude \\")
	fmt.Println("    --min 4.5 --max 10.0")
	fmt.Println()
	fmt.Println("  # Collect earthquakes in California region")
	fmt.Println("  quakewatch-scraper earthquakes region \\")
	fmt.Println("    --min-lat 32.0 --max-lat 42.0 \\")
	fmt.Println("    --min-lon -125.0 --max-lon -114.0")
	fmt.Println()

	fmt.Println("🏔️  FAULT EXAMPLES")
	fmt.Println("==================")
	fmt.Println("  # Collect fault data")
	fmt.Println("  quakewatch-scraper faults collect")
	fmt.Println()
	fmt.Println("  # Update fault data with retry")
	fmt.Println("  quakewatch-scraper faults update --retries 5")
	fmt.Println()

	fmt.Println("⏰ INTERVAL EXAMPLES")
	fmt.Println("===================")
	fmt.Println("  # Collect earthquakes every hour")
	fmt.Println("  quakewatch-scraper interval earthquakes recent \\")
	fmt.Println("    --interval 1h --max-runtime 24h")
	fmt.Println()
	fmt.Println("  # Collect faults every 6 hours")
	fmt.Println("  quakewatch-scraper interval faults collect \\")
	fmt.Println("    --interval 6h --daemon")
	fmt.Println()
	fmt.Println("  # Custom interval with multiple commands")
	fmt.Println("  quakewatch-scraper interval custom \\")
	fmt.Println("    --commands 'earthquakes recent,faults collect' \\")
	fmt.Println("    --interval 30m")
	fmt.Println()

	fmt.Println("🗄️  DATABASE EXAMPLES")
	fmt.Println("====================")
	fmt.Println("  # Initialize database")
	fmt.Println("  quakewatch-scraper db init")
	fmt.Println()
	fmt.Println("  # Run migrations")
	fmt.Println("  quakewatch-scraper db migrate up")
	fmt.Println()
	fmt.Println("  # Check database status")
	fmt.Println("  quakewatch-scraper db status")
	fmt.Println()

	fmt.Println("🛠️  UTILITY EXAMPLES")
	fmt.Println("====================")
	fmt.Println("  # Check system health")
	fmt.Println("  quakewatch-scraper health")
	fmt.Println()
	fmt.Println("  # View data statistics")
	fmt.Println("  quakewatch-scraper stats --type earthquakes")
	fmt.Println()
	fmt.Println("  # Validate collected data")
	fmt.Println("  quakewatch-scraper validate --type all")
	fmt.Println()
	fmt.Println("  # List all data files")
	fmt.Println("  quakewatch-scraper list --type all")
	fmt.Println()
	fmt.Println("  # Purge old data (dry run first)")
	fmt.Println("  quakewatch-scraper purge --dry-run")
	fmt.Println("  quakewatch-scraper purge --force")
	fmt.Println()

	fmt.Println("⚙️  CONFIGURATION EXAMPLES")
	fmt.Println("=========================")
	fmt.Println("  # Use custom config file")
	fmt.Println("  quakewatch-scraper earthquakes recent -c ./my-config.yaml")
	fmt.Println()
	fmt.Println("  # Enable verbose logging")
	fmt.Println("  quakewatch-scraper earthquakes recent -v")
	fmt.Println()
	fmt.Println("  # Output to custom directory")
	fmt.Println("  quakewatch-scraper earthquakes recent -o ./my-data")
	fmt.Println()
	fmt.Println("  # Output to stdout")
	fmt.Println("  quakewatch-scraper earthquakes recent --stdout")
	fmt.Println()
}

// showEarthquakeHelp displays earthquake-specific help
func (a *App) showEarthquakeHelp() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🌍 Earthquake Commands 🌍                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("OVERVIEW")
	fmt.Println("=========")
	fmt.Println("Collect earthquake data from the USGS FDSNWS API. All commands")
	fmt.Println("support filtering, limiting, and custom output options.")
	fmt.Println()

	fmt.Println("COMMANDS")
	fmt.Println("=========")
	fmt.Println("  recent                    Collect recent earthquakes (last hour)")
	fmt.Println("  time-range                Collect earthquakes by time range")
	fmt.Println("  magnitude                 Collect earthquakes by magnitude range")
	fmt.Println("  significant               Collect significant earthquakes (M4.5+)")
	fmt.Println("  region                    Collect earthquakes by geographic region")
	fmt.Println("  country                   Collect earthquakes by country")
	fmt.Println()

	fmt.Println("COMMON OPTIONS")
	fmt.Println("==============")
	fmt.Println("  --limit <number>          Maximum number of records (default: 1000)")
	fmt.Println("  --filename <name>         Custom filename (without extension)")
	fmt.Println("  --storage <backend>       Storage backend (json, postgresql)")
	fmt.Println("  --smart                   Use smart collection to avoid duplicates")
	fmt.Println("  --hours-back <number>     Hours to look back (recent command)")
	fmt.Println()

	fmt.Println("EXAMPLES")
	fmt.Println("=========")
	fmt.Println("  # Basic recent collection")
	fmt.Println("  quakewatch-scraper earthquakes recent")
	fmt.Println()
	fmt.Println("  # Time range with limit")
	fmt.Println("  quakewatch-scraper earthquakes time-range \\")
	fmt.Println("    --start '2024-01-01' --end '2024-01-31' --limit 500")
	fmt.Println()
	fmt.Println("  # Magnitude filtering")
	fmt.Println("  quakewatch-scraper earthquakes magnitude --min 4.5 --max 10.0")
	fmt.Println()
	fmt.Println("  # Geographic region")
	fmt.Println("  quakewatch-scraper earthquakes region \\")
	fmt.Println("    --min-lat 32.0 --max-lat 42.0 \\")
	fmt.Println("    --min-lon -125.0 --max-lon -114.0")
	fmt.Println()
}

// showFaultHelp displays fault-specific help
func (a *App) showFaultHelp() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🏔️  Fault Commands 🏔️                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("OVERVIEW")
	fmt.Println("=========")
	fmt.Println("Collect fault data from the EMSC-CSEM API. Fault data includes")
	fmt.Println("information about geological faults and their characteristics.")
	fmt.Println()

	fmt.Println("COMMANDS")
	fmt.Println("=========")
	fmt.Println("  collect                   Collect fault data from EMSC")
	fmt.Println("  update                    Update fault data with retry logic")
	fmt.Println()

	fmt.Println("COMMON OPTIONS")
	fmt.Println("==============")
	fmt.Println("  --filename <name>         Custom filename (without extension)")
	fmt.Println("  --retries <number>        Number of retry attempts (update)")
	fmt.Println("  --retry-delay <duration>  Delay between retries (update)")
	fmt.Println()

	fmt.Println("EXAMPLES")
	fmt.Println("=========")
	fmt.Println("  # Basic fault collection")
	fmt.Println("  quakewatch-scraper faults collect")
	fmt.Println()
	fmt.Println("  # Update with retry")
	fmt.Println("  quakewatch-scraper faults update --retries 5")
	fmt.Println()
}

// showIntervalHelp displays interval-specific help
func (a *App) showIntervalHelp() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ⏰ Interval Commands ⏰                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("OVERVIEW")
	fmt.Println("=========")
	fmt.Println("Run data collection commands at regular intervals. Supports")
	fmt.Println("various scheduling options, error handling, and daemon mode.")
	fmt.Println()

	fmt.Println("COMMANDS")
	fmt.Println("=========")
	fmt.Println("  earthquakes recent        Run recent earthquake collection")
	fmt.Println("  earthquakes time-range    Run time range earthquake collection")
	fmt.Println("  earthquakes magnitude     Run magnitude earthquake collection")
	fmt.Println("  earthquakes significant   Run significant earthquake collection")
	fmt.Println("  earthquakes region        Run region earthquake collection")
	fmt.Println("  earthquakes country       Run country earthquake collection")
	fmt.Println("  faults collect            Run fault collection")
	fmt.Println("  faults update             Run fault update")
	fmt.Println("  custom                    Run custom command combinations")
	fmt.Println()

	fmt.Println("INTERVAL OPTIONS")
	fmt.Println("================")
	fmt.Println("  --interval <duration>     Time interval (e.g., '5m', '1h', '24h')")
	fmt.Println("  --max-runtime <duration>  Maximum total runtime")
	fmt.Println("  --max-executions <number> Maximum number of executions")
	fmt.Println("  --backoff <strategy>      Backoff strategy (none, linear, exponential)")
	fmt.Println("  --max-backoff <duration>  Maximum backoff duration")
	fmt.Println("  --continue-on-error       Continue on individual failures")
	fmt.Println("  --skip-empty              Skip if no new data found")
	fmt.Println("  --daemon                  Run in background mode")
	fmt.Println("  --pid-file <path>         PID file location")
	fmt.Println("  --log-file <path>         Log file for daemon mode")
	fmt.Println()

	fmt.Println("EXAMPLES")
	fmt.Println("=========")
	fmt.Println("  # Collect earthquakes every hour")
	fmt.Println("  quakewatch-scraper interval earthquakes recent --interval 1h")
	fmt.Println()
	fmt.Println("  # Run as daemon with max runtime")
	fmt.Println("  quakewatch-scraper interval earthquakes recent \\")
	fmt.Println("    --interval 30m --max-runtime 24h --daemon")
	fmt.Println()
	fmt.Println("  # Custom commands every 15 minutes")
	fmt.Println("  quakewatch-scraper interval custom \\")
	fmt.Println("    --commands 'earthquakes recent,faults collect' \\")
	fmt.Println("    --interval 15m")
	fmt.Println()
}

// showDatabaseHelp displays database-specific help
func (a *App) showDatabaseHelp() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🗄️  Database Commands 🗄️                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("OVERVIEW")
	fmt.Println("=========")
	fmt.Println("Manage PostgreSQL database operations including initialization,")
	fmt.Println("migrations, and status monitoring.")
	fmt.Println()

	fmt.Println("COMMANDS")
	fmt.Println("=========")
	fmt.Println("  init                      Initialize database and run migrations")
	fmt.Println("  migrate up                Run all pending migrations")
	fmt.Println("  migrate down              Rollback all migrations")
	fmt.Println("  migrate to                Migrate to specific version")
	fmt.Println("  migrate force             Force migration version")
	fmt.Println("  status                    Show database status and migration info")
	fmt.Println()

	fmt.Println("COMMON OPTIONS")
	fmt.Println("==============")
	fmt.Println("  --force                   Force operations (init)")
	fmt.Println("  --version <number>        Target migration version")
	fmt.Println()

	fmt.Println("EXAMPLES")
	fmt.Println("=========")
	fmt.Println("  # Initialize database")
	fmt.Println("  quakewatch-scraper db init")
	fmt.Println()
	fmt.Println("  # Run migrations")
	fmt.Println("  quakewatch-scraper db migrate up")
	fmt.Println()
	fmt.Println("  # Check status")
	fmt.Println("  quakewatch-scraper db status")
	fmt.Println()
}

// showUtilityHelp displays utility-specific help
func (a *App) showUtilityHelp() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🛠️  Utility Commands 🛠️                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("OVERVIEW")
	fmt.Println("=========")
	fmt.Println("Utility commands for data management, system monitoring, and")
	fmt.Println("application maintenance.")
	fmt.Println()

	fmt.Println("COMMANDS")
	fmt.Println("=========")
	fmt.Println("  health                    Check system and API health")
	fmt.Println("  stats                     Show data collection statistics")
	fmt.Println("  validate                  Validate data integrity")
	fmt.Println("  list                      List available data files")
	fmt.Println("  purge                     Delete collected data files")
	fmt.Println("  version                   Show version information")
	fmt.Println("  config                    Manage application configuration")
	fmt.Println()

	fmt.Println("COMMON OPTIONS")
	fmt.Println("==============")
	fmt.Println("  --type <type>             Data type (earthquakes, faults, all)")
	fmt.Println("  --file <path>             Specific file to process")
	fmt.Println("  --force                   Force operations without confirmation")
	fmt.Println("  --dry-run                 Show what would be done")
	fmt.Println()

	fmt.Println("EXAMPLES")
	fmt.Println("=========")
	fmt.Println("  # Check system health")
	fmt.Println("  quakewatch-scraper health")
	fmt.Println()
	fmt.Println("  # View statistics")
	fmt.Println("  quakewatch-scraper stats --type earthquakes")
	fmt.Println()
	fmt.Println("  # Validate data")
	fmt.Println("  quakewatch-scraper validate --type all")
	fmt.Println()
	fmt.Println("  # List files")
	fmt.Println("  quakewatch-scraper list --type all")
	fmt.Println()
	fmt.Println("  # Purge with confirmation")
	fmt.Println("  quakewatch-scraper purge --dry-run")
	fmt.Println("  quakewatch-scraper purge --force")
	fmt.Println()
}

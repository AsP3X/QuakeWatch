package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"quakewatch-scraper/internal/api"
	"quakewatch-scraper/internal/collector"
	"quakewatch-scraper/internal/storage"
	"quakewatch-scraper/internal/utils"
)

// App represents the main CLI application
type App struct {
	rootCmd *cobra.Command
	logger  *utils.Logger
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

	// Add utility commands
	a.rootCmd.AddCommand(a.newValidateCmd())
	a.rootCmd.AddCommand(a.newStatsCmd())
	a.rootCmd.AddCommand(a.newListCmd())
	a.rootCmd.AddCommand(a.newPurgeCmd())
	a.rootCmd.AddCommand(a.newHealthCmd())
	a.rootCmd.AddCommand(a.newVersionCmd())
}

func (a *App) setupFlags() {
	a.rootCmd.PersistentFlags().StringP("config", "c", "./configs/config.yaml", "Configuration file path")
	a.rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	a.rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress output")
	a.rootCmd.PersistentFlags().String("log-level", "info", "Set log level (error, warn, info, debug)")
	a.rootCmd.PersistentFlags().StringP("output-dir", "o", "./data", "Output directory for JSON files")
	a.rootCmd.PersistentFlags().Bool("dry-run", false, "Show what would be done without executing")
}

func (a *App) Run(args []string) error {
	// Remove the first argument (binary name) - it could be "./bin/quakewatch-scraper" or "quakewatch-scraper"
	if len(args) > 0 {
		args = args[1:]
	}
	a.rootCmd.SetArgs(args)
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
	timeRangeCmd.MarkFlagRequired("start")
	timeRangeCmd.MarkFlagRequired("end")
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
	magnitudeCmd.MarkFlagRequired("min")
	magnitudeCmd.MarkFlagRequired("max")
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
	significantCmd.MarkFlagRequired("start")
	significantCmd.MarkFlagRequired("end")
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
	regionCmd.MarkFlagRequired("min-lat")
	regionCmd.MarkFlagRequired("max-lat")
	regionCmd.MarkFlagRequired("min-lon")
	regionCmd.MarkFlagRequired("max-lon")
	cmd.AddCommand(regionCmd)

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

// Helper methods for command execution
func (a *App) runRecentEarthquakes(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	// Initialize components
	storage := storage.NewJSONStorage("./data")
	usgsClient := api.NewUSGSClient("https://earthquake.usgs.gov/fdsnws/event/1", 30*time.Second)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	return collector.CollectRecent(limit, filename)
}

func (a *App) runTimeRangeEarthquakes(cmd *cobra.Command, args []string) error {
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	// Initialize components
	storage := storage.NewJSONStorage("./data")
	usgsClient := api.NewUSGSClient("https://earthquake.usgs.gov/fdsnws/event/1", 30*time.Second)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	return collector.CollectByTimeRange(startTime, endTime, limit, filename)
}

func (a *App) runMagnitudeEarthquakes(cmd *cobra.Command, args []string) error {
	minMag, _ := cmd.Flags().GetFloat64("min")
	maxMag, _ := cmd.Flags().GetFloat64("max")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	// Initialize components
	storage := storage.NewJSONStorage("./data")
	usgsClient := api.NewUSGSClient("https://earthquake.usgs.gov/fdsnws/event/1", 30*time.Second)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	return collector.CollectByMagnitude(minMag, maxMag, limit, filename)
}

func (a *App) runSignificantEarthquakes(cmd *cobra.Command, args []string) error {
	startStr, _ := cmd.Flags().GetString("start")
	endStr, _ := cmd.Flags().GetString("end")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	startTime, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return fmt.Errorf("invalid start time format: %w", err)
	}

	endTime, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return fmt.Errorf("invalid end time format: %w", err)
	}

	// Initialize components
	storage := storage.NewJSONStorage("./data")
	usgsClient := api.NewUSGSClient("https://earthquake.usgs.gov/fdsnws/event/1", 30*time.Second)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	return collector.CollectSignificant(startTime, endTime, limit, filename)
}

func (a *App) runRegionEarthquakes(cmd *cobra.Command, args []string) error {
	minLat, _ := cmd.Flags().GetFloat64("min-lat")
	maxLat, _ := cmd.Flags().GetFloat64("max-lat")
	minLon, _ := cmd.Flags().GetFloat64("min-lon")
	maxLon, _ := cmd.Flags().GetFloat64("max-lon")
	limit, _ := cmd.Flags().GetInt("limit")
	filename, _ := cmd.Flags().GetString("filename")

	// Initialize components
	storage := storage.NewJSONStorage("./data")
	usgsClient := api.NewUSGSClient("https://earthquake.usgs.gov/fdsnws/event/1", 30*time.Second)
	collector := collector.NewEarthquakeCollector(usgsClient, storage)

	return collector.CollectByRegion(minLat, maxLat, minLon, maxLon, limit, filename)
}

func (a *App) runCollectFaults(cmd *cobra.Command, args []string) error {
	filename, _ := cmd.Flags().GetString("filename")

	// Initialize components
	storage := storage.NewJSONStorage("./data")
	emscClient := api.NewEMSCClient("https://www.emsc-csem.org/javascript", 30*time.Second)
	collector := collector.NewFaultCollector(emscClient, storage)

	return collector.CollectFaults(filename)
}

func (a *App) runUpdateFaults(cmd *cobra.Command, args []string) error {
	filename, _ := cmd.Flags().GetString("filename")
	retries, _ := cmd.Flags().GetInt("retries")
	retryDelay, _ := cmd.Flags().GetDuration("retry-delay")

	// Initialize components
	storage := storage.NewJSONStorage("./data")
	emscClient := api.NewEMSCClient("https://www.emsc-csem.org/javascript", 30*time.Second)
	collector := collector.NewFaultCollector(emscClient, storage)

	return collector.UpdateFaults(filename, retries, retryDelay)
}

func (a *App) runValidate(cmd *cobra.Command, args []string) error {
	dataType, _ := cmd.Flags().GetString("type")
	file, _ := cmd.Flags().GetString("file")

	storage := storage.NewJSONStorage("./data")

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

	storage := storage.NewJSONStorage("./data")

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

	storage := storage.NewJSONStorage("./data")

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

	storage := storage.NewJSONStorage("./data")

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
		fmt.Scanln(&response)

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
	usgsClient := api.NewUSGSClient("https://earthquake.usgs.gov/fdsnws/event/1", 10*time.Second)
	_, err := usgsClient.GetRecentEarthquakes(1)
	if err != nil {
		fmt.Printf("  ✗ USGS API: %v\n", err)
	} else {
		fmt.Println("  ✓ USGS API: OK")
	}

	// Check EMSC API
	emscClient := api.NewEMSCClient("https://www.emsc-csem.org/javascript", 10*time.Second)
	_, err = emscClient.GetFaults()
	if err != nil {
		fmt.Printf("  ✗ EMSC API: %v\n", err)
	} else {
		fmt.Println("  ✓ EMSC API: OK")
	}

	// Check storage
	storage := storage.NewJSONStorage("./data")
	_, err = storage.ListFiles("earthquakes")
	if err != nil {
		fmt.Printf("  ✗ Storage: %v\n", err)
	} else {
		fmt.Println("  ✓ Storage: OK")
	}

	return nil
}

func (a *App) runVersion(cmd *cobra.Command, args []string) {
	fmt.Println("QuakeWatch Scraper v1.0.0")
	fmt.Println("Go version: 1.21")
	fmt.Println("Build date: " + time.Now().Format("2006-01-02"))
}

// showBanner displays the application banner when no command is provided
func (a *App) showBanner(cmd *cobra.Command, args []string) {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    🌋 QuakeWatch Scraper 🌋                    ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  A powerful tool for collecting earthquake and fault data    ║")
	fmt.Println("║  from various geological sources and APIs.                   ║")
	fmt.Println("║                                                              ║")
	fmt.Println("║  Version: 1.0.0                                              ║")
	fmt.Println("║  Built with Go                                               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Show the help after the banner
	cmd.Help()
}

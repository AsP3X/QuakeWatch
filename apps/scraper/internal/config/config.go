package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"quakewatch-scraper/internal/utils"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	API        APIConfig        `mapstructure:"api"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Collection CollectionConfig `mapstructure:"collection"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Interval   IntervalConfig   `mapstructure:"interval"`
}

// APIConfig contains API-related configuration
type APIConfig struct {
	USGS USGSConfig `mapstructure:"usgs"`
	EMSC EMSCConfig `mapstructure:"emsc"`
}

// USGSConfig contains USGS API configuration
type USGSConfig struct {
	BaseURL   string        `mapstructure:"base_url"`
	Timeout   time.Duration `mapstructure:"timeout"`
	RateLimit int           `mapstructure:"rate_limit"`
}

// EMSCConfig contains EMSC API configuration
type EMSCConfig struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// StorageConfig contains storage-related configuration
type StorageConfig struct {
	OutputDir      string `mapstructure:"output_dir"`
	EarthquakesDir string `mapstructure:"earthquakes_dir"`
	FaultsDir      string `mapstructure:"faults_dir"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// CollectionConfig contains data collection configuration
type CollectionConfig struct {
	DefaultLimit  int           `mapstructure:"default_limit"`
	MaxLimit      int           `mapstructure:"max_limit"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
}

// IntervalConfig contains interval scraping configuration
type IntervalConfig struct {
	DefaultInterval     time.Duration `mapstructure:"default_interval"`
	MaxRuntime          time.Duration `mapstructure:"max_runtime"`
	MaxExecutions       int           `mapstructure:"max_executions"`
	BackoffStrategy     string        `mapstructure:"backoff_strategy"`
	MaxBackoff          time.Duration `mapstructure:"max_backoff"`
	ContinueOnError     bool          `mapstructure:"continue_on_error"`
	SkipEmpty           bool          `mapstructure:"skip_empty"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	DaemonMode          bool          `mapstructure:"daemon_mode"`
	PIDFile             string        `mapstructure:"pid_file"`
	LogFile             string        `mapstructure:"log_file"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	pathManager := utils.NewPathManager()

	return &Config{
		API: APIConfig{
			USGS: USGSConfig{
				BaseURL:   "https://earthquake.usgs.gov/fdsnws/event/1",
				Timeout:   30 * time.Second,
				RateLimit: 60,
			},
			EMSC: EMSCConfig{
				BaseURL: "https://www.emsc-csem.org/javascript",
				Timeout: 30 * time.Second,
			},
		},
		Storage: StorageConfig{
			OutputDir:      pathManager.GetDefaultDataDir(),
			EarthquakesDir: "earthquakes",
			FaultsDir:      "faults",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Collection: CollectionConfig{
			DefaultLimit:  1000,
			MaxLimit:      10000,
			RetryAttempts: 3,
			RetryDelay:    5 * time.Second,
		},
		Database: DatabaseConfig{
			Enabled:           false,
			Type:              "postgres",
			Host:              "localhost",
			Port:              5432,
			User:              "postgres",
			Password:          "",
			Database:          "quakewatch",
			SSLMode:           "disable",
			MaxOpenConns:      25,
			MaxIdleConns:      5,
			ConnMaxLifetime:   5 * time.Minute,
			ConnMaxIdleTime:   5 * time.Minute,
			MaxConnections:    10,
			ConnectionTimeout: 30 * time.Second,
		},
		Interval: IntervalConfig{
			DefaultInterval:     1 * time.Hour,
			MaxRuntime:          24 * time.Hour,
			MaxExecutions:       1000,
			BackoffStrategy:     "exponential",
			MaxBackoff:          30 * time.Minute,
			ContinueOnError:     true,
			SkipEmpty:           false,
			HealthCheckInterval: 5 * time.Minute,
			DaemonMode:          false,
			PIDFile:             pathManager.GetDefaultPIDFile(),
			LogFile:             pathManager.GetDefaultLogFile(),
		},
	}
}

// LoadConfig loads configuration from file or creates default if not exists
func LoadConfig(configPath string) (*Config, error) {
	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Add multiple search paths for config file
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")

		// Add executable directory path
		if execDir, err := getExecutableDir(); err == nil {
			viper.AddConfigPath(execDir)
			viper.AddConfigPath(filepath.Join(execDir, "configs"))
		}

		// Add platform-specific paths
		pathManager := utils.NewPathManager()
		viper.AddConfigPath(pathManager.GetDefaultConfigDir())
	}

	// Try to read the config file
	if err := viper.ReadInConfig(); err != nil {
		// Check if it's a config file not found error
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, check if user wants to create one
			return handleMissingConfig(configPath)
		}
		// Check if the error message indicates a missing file or directory
		errMsg := err.Error()
		if contains(errMsg, "no such file or directory") ||
			contains(errMsg, "The system cannot find the path specified") ||
			contains(errMsg, "cannot find the file specified") {
			return handleMissingConfig(configPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// handleMissingConfig handles the case when no config file is found
func handleMissingConfig(configPath string) (*Config, error) {
	fmt.Println("No configuration file found.")
	fmt.Print("Would you like to create a configuration file? (y/N): ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return nil, fmt.Errorf("failed to read user input: %w", err)
	}

	if response == "y" || response == "Y" || response == "yes" || response == "YES" {
		return CreateInteractiveConfig(configPath)
	}

	// User chose not to create config, use defaults
	fmt.Println("Using default configuration with platform-appropriate paths.")

	// Create default config directories if they don't exist
	config := DefaultConfig()

	// Ensure data directory exists
	if err := os.MkdirAll(config.Storage.OutputDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create data directory %s: %v\n", config.Storage.OutputDir, err)
	}

	// Ensure config directory exists - try executable directory first
	configDir := "./configs"
	if configPath != "" {
		configDir = filepath.Dir(configPath)
	} else {
		// Try to create config in executable directory
		if execDir, err := getExecutableDir(); err == nil {
			configDir = filepath.Join(execDir, "configs")
		}
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create config directory %s: %v\n", configDir, err)
	}

	return config, nil
}

// CreateInteractiveConfig creates a configuration file through user interaction
func CreateInteractiveConfig(configPath string) (*Config, error) {
	fmt.Println("\n=== QuakeWatch Scraper Configuration Setup ===")

	config := DefaultConfig()

	// API Configuration
	fmt.Println("\n--- API Configuration ---")

	// USGS API
	fmt.Printf("USGS API Base URL (default: %s): ", config.API.USGS.BaseURL)
	usgsURL := readInput()
	if usgsURL != "" {
		config.API.USGS.BaseURL = usgsURL
	}

	fmt.Printf("USGS API Timeout in seconds (default: %.0f): ", config.API.USGS.Timeout.Seconds())
	usgsTimeoutStr := readInput()
	if usgsTimeoutStr != "" {
		if usgsTimeout, err := strconv.Atoi(usgsTimeoutStr); err == nil && usgsTimeout > 0 {
			config.API.USGS.Timeout = time.Duration(usgsTimeout) * time.Second
		}
	}

	fmt.Printf("USGS API Rate Limit (default: %d): ", config.API.USGS.RateLimit)
	usgsRateLimitStr := readInput()
	if usgsRateLimitStr != "" {
		if usgsRateLimit, err := strconv.Atoi(usgsRateLimitStr); err == nil && usgsRateLimit > 0 {
			config.API.USGS.RateLimit = usgsRateLimit
		}
	}

	// EMSC API
	fmt.Printf("EMSC API Base URL (default: %s): ", config.API.EMSC.BaseURL)
	var emscURL string
	fmt.Scanln(&emscURL)
	if emscURL != "" {
		config.API.EMSC.BaseURL = emscURL
	}

	fmt.Printf("EMSC API Timeout in seconds (default: %.0f): ", config.API.EMSC.Timeout.Seconds())
	var emscTimeout int
	fmt.Scanln(&emscTimeout)
	if emscTimeout > 0 {
		config.API.EMSC.Timeout = time.Duration(emscTimeout) * time.Second
	}

	// Storage Configuration
	fmt.Println("\n--- Storage Configuration ---")

	fmt.Printf("Output Directory (default: %s): ", config.Storage.OutputDir)
	var outputDir string
	fmt.Scanln(&outputDir)
	if outputDir != "" {
		config.Storage.OutputDir = outputDir
	}

	fmt.Printf("Earthquakes Directory (default: %s): ", config.Storage.EarthquakesDir)
	var earthquakesDir string
	fmt.Scanln(&earthquakesDir)
	if earthquakesDir != "" {
		config.Storage.EarthquakesDir = earthquakesDir
	}

	fmt.Printf("Faults Directory (default: %s): ", config.Storage.FaultsDir)
	var faultsDir string
	fmt.Scanln(&faultsDir)
	if faultsDir != "" {
		config.Storage.FaultsDir = faultsDir
	}

	// Logging Configuration
	fmt.Println("\n--- Logging Configuration ---")

	fmt.Printf("Log Level (debug/info/warn/error, default: %s): ", config.Logging.Level)
	var logLevel string
	fmt.Scanln(&logLevel)
	if logLevel != "" {
		config.Logging.Level = logLevel
	}

	fmt.Printf("Log Format (json/text, default: %s): ", config.Logging.Format)
	var logFormat string
	fmt.Scanln(&logFormat)
	if logFormat != "" {
		config.Logging.Format = logFormat
	}

	// Collection Configuration
	fmt.Println("\n--- Collection Configuration ---")

	fmt.Printf("Default Limit (default: %d): ", config.Collection.DefaultLimit)
	var defaultLimit int
	fmt.Scanln(&defaultLimit)
	if defaultLimit > 0 {
		config.Collection.DefaultLimit = defaultLimit
	}

	fmt.Printf("Max Limit (default: %d): ", config.Collection.MaxLimit)
	var maxLimit int
	fmt.Scanln(&maxLimit)
	if maxLimit > 0 {
		config.Collection.MaxLimit = maxLimit
	}

	fmt.Printf("Retry Attempts (default: %d): ", config.Collection.RetryAttempts)
	var retryAttempts int
	fmt.Scanln(&retryAttempts)
	if retryAttempts > 0 {
		config.Collection.RetryAttempts = retryAttempts
	}

	fmt.Printf("Retry Delay in seconds (default: %.0f): ", config.Collection.RetryDelay.Seconds())
	var retryDelay int
	fmt.Scanln(&retryDelay)
	if retryDelay > 0 {
		config.Collection.RetryDelay = time.Duration(retryDelay) * time.Second
	}

	// Database Configuration
	fmt.Println("\n--- Database Configuration ---")

	fmt.Printf("Database Enabled (default: %t): ", config.Database.Enabled)
	var dbEnabled bool
	fmt.Scanln(&dbEnabled)
	if dbEnabled != config.Database.Enabled {
		config.Database.Enabled = dbEnabled
	}

	if config.Database.Enabled {
		fmt.Printf("Database Type (postgres/sqlite, default: %s): ", config.Database.Type)
		var dbType string
		fmt.Scanln(&dbType)
		if dbType != "" {
			config.Database.Type = dbType
		}

		fmt.Printf("Database Host (default: %s): ", config.Database.Host)
		var dbHost string
		fmt.Scanln(&dbHost)
		if dbHost != "" {
			config.Database.Host = dbHost
		}

		fmt.Printf("Database Port (default: %d): ", config.Database.Port)
		var dbPort int
		fmt.Scanln(&dbPort)
		if dbPort > 0 {
			config.Database.Port = dbPort
		}

		fmt.Printf("Database Username (default: %s): ", config.Database.User)
		var dbUsername string
		fmt.Scanln(&dbUsername)
		if dbUsername != "" {
			config.Database.User = dbUsername
		}

		fmt.Printf("Database Password (default: %s): ", config.Database.Password)
		var dbPassword string
		fmt.Scanln(&dbPassword)
		if dbPassword != "" {
			config.Database.Password = dbPassword
		}

		fmt.Printf("Database Name (default: %s): ", config.Database.Database)
		var dbName string
		fmt.Scanln(&dbName)
		if dbName != "" {
			config.Database.Database = dbName
		}

		fmt.Printf("Database SSL Mode (disable/require/verify-ca/verify-full, default: %s): ", config.Database.SSLMode)
		var dbSSLMode string
		fmt.Scanln(&dbSSLMode)
		if dbSSLMode != "" {
			config.Database.SSLMode = dbSSLMode
		}

		fmt.Printf("Database Max Connections (default: %d): ", config.Database.MaxConnections)
		var dbMaxConnections int
		fmt.Scanln(&dbMaxConnections)
		if dbMaxConnections > 0 {
			config.Database.MaxConnections = dbMaxConnections
		}

		fmt.Printf("Database Connection Timeout in seconds (default: %.0f): ", config.Database.ConnectionTimeout.Seconds())
		var dbTimeout int
		fmt.Scanln(&dbTimeout)
		if dbTimeout > 0 {
			config.Database.ConnectionTimeout = time.Duration(dbTimeout) * time.Second
		}
	}

	// Save the configuration
	if err := SaveConfig(config, configPath); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\nConfiguration saved to: %s\n", getConfigPath(configPath))
	return config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, configPath string) error {
	// Set up viper with the config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	// Set the configuration values
	viper.Set("api.usgs.base_url", config.API.USGS.BaseURL)
	viper.Set("api.usgs.timeout", config.API.USGS.Timeout)
	viper.Set("api.usgs.rate_limit", config.API.USGS.RateLimit)
	viper.Set("api.emsc.base_url", config.API.EMSC.BaseURL)
	viper.Set("api.emsc.timeout", config.API.EMSC.Timeout)

	viper.Set("storage.output_dir", config.Storage.OutputDir)
	viper.Set("storage.earthquakes_dir", config.Storage.EarthquakesDir)
	viper.Set("storage.faults_dir", config.Storage.FaultsDir)

	viper.Set("logging.level", config.Logging.Level)
	viper.Set("logging.format", config.Logging.Format)
	viper.Set("logging.output", config.Logging.Output)

	viper.Set("collection.default_limit", config.Collection.DefaultLimit)
	viper.Set("collection.max_limit", config.Collection.MaxLimit)
	viper.Set("collection.retry_attempts", config.Collection.RetryAttempts)
	viper.Set("collection.retry_delay", config.Collection.RetryDelay)

	viper.Set("database.enabled", config.Database.Enabled)
	viper.Set("database.type", config.Database.Type)
	viper.Set("database.host", config.Database.Host)
	viper.Set("database.port", config.Database.Port)
	viper.Set("database.username", config.Database.User)
	viper.Set("database.password", config.Database.Password)
	viper.Set("database.database", config.Database.Database)
	viper.Set("database.ssl_mode", config.Database.SSLMode)
	viper.Set("database.max_connections", config.Database.MaxConnections)
	viper.Set("database.connection_timeout", config.Database.ConnectionTimeout)

	// Ensure the directory exists
	configDir := filepath.Dir(getConfigPath(configPath))
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the config file
	if err := viper.WriteConfigAs(getConfigPath(configPath)); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath returns the full path to the config file
func getConfigPath(configPath string) string {
	if configPath != "" {
		return configPath
	}
	return "./configs/config.yaml"
}

// getExecutableDir returns the directory containing the current executable
func getExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(execPath), nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// readInput reads a line of input and returns it, handling empty input gracefully
func readInput() string {
	var input string
	fmt.Scanln(&input)
	// Trim whitespace and check for common "no" responses
	input = strings.TrimSpace(input)
	if input == "n" || input == "N" || input == "no" || input == "NO" {
		return ""
	}
	return input
}

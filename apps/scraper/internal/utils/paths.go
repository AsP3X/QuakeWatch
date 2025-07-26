package utils

import (
	"os"
	"path/filepath"
)

// PathManager handles platform-specific path operations
type PathManager struct {
	platform *Platform
}

// NewPathManager creates a new path manager
func NewPathManager() *PathManager {
	return &PathManager{
		platform: GetPlatform(),
	}
}

// GetDefaultPIDFile returns the default PID file path for the platform
func (p *PathManager) GetDefaultPIDFile() string {
	if p.platform.IsWindows {
		return filepath.Join(os.Getenv("TEMP"), "quakewatch-scraper.pid")
	}
	return "/var/run/quakewatch-scraper.pid"
}

// GetDefaultLogFile returns the default log file path for the platform
func (p *PathManager) GetDefaultLogFile() string {
	if p.platform.IsWindows {
		return filepath.Join(os.Getenv("TEMP"), "quakewatch-scraper.log")
	}
	return "/var/log/quakewatch-scraper.log"
}

// GetDefaultDataDir returns the default data directory for the platform
func (p *PathManager) GetDefaultDataDir() string {
	if p.platform.IsWindows {
		// Use %APPDATA% on Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			// Fallback to current directory
			return "./data"
		}
		return filepath.Join(appData, "QuakeWatch", "data")
	}
	return "./data"
}

// GetDefaultConfigDir returns the default configuration directory for the platform
func (p *PathManager) GetDefaultConfigDir() string {
	if p.platform.IsWindows {
		// Use %APPDATA% on Windows
		appData := os.Getenv("APPDATA")
		if appData == "" {
			// Fallback to current directory
			return "./configs"
		}
		return filepath.Join(appData, "QuakeWatch", "configs")
	}
	return "./configs"
}

// EnsureDirectoryExists creates directory with appropriate permissions
func (p *PathManager) EnsureDirectoryExists(path string) error {
	if err := os.MkdirAll(path, p.getDefaultDirPerms()); err != nil {
		return err
	}
	return nil
}

// getDefaultDirPerms returns platform-appropriate directory permissions
func (p *PathManager) getDefaultDirPerms() os.FileMode {
	if p.platform.IsWindows {
		return 0755 // Windows doesn't use Unix permissions, but Go handles this
	}
	return 0755
}

// NormalizePath converts path separators to platform-appropriate ones
func (p *PathManager) NormalizePath(path string) string {
	return filepath.Clean(path)
}

// IsAbsolutePath checks if a path is absolute for the current platform
func (p *PathManager) IsAbsolutePath(path string) bool {
	return filepath.IsAbs(path)
}

// JoinPaths joins path components using platform-appropriate separators
func (p *PathManager) JoinPaths(elem ...string) string {
	return filepath.Join(elem...)
}

// GetExecutablePath returns the path to the current executable
func (p *PathManager) GetExecutablePath() (string, error) {
	return os.Executable()
}

// GetExecutableDir returns the directory containing the current executable
func (p *PathManager) GetExecutableDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(execPath), nil
}

package utils

import (
	"runtime"
	"strings"
)

// Platform represents the current operating system platform
type Platform struct {
	OS        string
	Arch      string
	IsUnix    bool
	IsWindows bool
}

// GetPlatform returns the current platform information
func GetPlatform() *Platform {
	os := runtime.GOOS
	arch := runtime.GOARCH

	return &Platform{
		OS:        os,
		Arch:      arch,
		IsUnix:    os == "linux" || os == "darwin" || os == "freebsd" || os == "openbsd",
		IsWindows: os == "windows",
	}
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsUnix returns true if running on Unix-like system
func IsUnix() bool {
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin" ||
		runtime.GOOS == "freebsd" || runtime.GOOS == "openbsd"
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsDarwin returns true if running on macOS
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// GetOSName returns a human-readable OS name
func GetOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	case "freebsd":
		return "FreeBSD"
	case "openbsd":
		return "OpenBSD"
	default:
		return strings.Title(runtime.GOOS)
	}
}

// GetArchName returns a human-readable architecture name
func GetArchName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "ARM64"
	case "386":
		return "x86"
	case "arm":
		return "ARM"
	default:
		return runtime.GOARCH
	}
}

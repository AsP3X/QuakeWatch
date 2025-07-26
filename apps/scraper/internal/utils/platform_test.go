package utils

import (
	"testing"
)

func TestPlatformDetection(t *testing.T) {
	platform := GetPlatform()

	if platform.OS == "" {
		t.Error("Platform OS should not be empty")
	}

	if platform.Arch == "" {
		t.Error("Platform Arch should not be empty")
	}

	// Test platform-specific flags
	if platform.IsWindows && platform.IsUnix {
		t.Error("Platform cannot be both Windows and Unix")
	}
}

func TestIsWindows(t *testing.T) {
	// This test will pass on Windows and be skipped on other platforms
	if !IsWindows() {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	if !IsWindows() {
		t.Error("IsWindows() should return true on Windows")
	}
}

func TestIsUnix(t *testing.T) {
	// This test will pass on Unix systems and be skipped on Windows
	if IsWindows() {
		t.Skip("Skipping Unix-specific test on Windows platform")
	}

	if !IsUnix() {
		t.Error("IsUnix() should return true on Unix systems")
	}
}

func TestGetOSName(t *testing.T) {
	osName := GetOSName()
	if osName == "" {
		t.Error("OS name should not be empty")
	}

	// Test that we get expected names for known platforms
	platform := GetPlatform()
	switch platform.OS {
	case "windows":
		if osName != "Windows" {
			t.Errorf("Expected 'Windows', got '%s'", osName)
		}
	case "linux":
		if osName != "Linux" {
			t.Errorf("Expected 'Linux', got '%s'", osName)
		}
	case "darwin":
		if osName != "macOS" {
			t.Errorf("Expected 'macOS', got '%s'", osName)
		}
	}
}

func TestGetArchName(t *testing.T) {
	archName := GetArchName()
	if archName == "" {
		t.Error("Architecture name should not be empty")
	}

	// Test that we get expected names for known architectures
	platform := GetPlatform()
	switch platform.Arch {
	case "amd64":
		if archName != "x86_64" {
			t.Errorf("Expected 'x86_64', got '%s'", archName)
		}
	case "arm64":
		if archName != "ARM64" {
			t.Errorf("Expected 'ARM64', got '%s'", archName)
		}
	case "386":
		if archName != "x86" {
			t.Errorf("Expected 'x86', got '%s'", archName)
		}
	}
}

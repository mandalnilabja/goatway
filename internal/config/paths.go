package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DataDir returns the path to the Goatway data directory.
// Priority:
// 1. GOATWAY_DATA_DIR environment variable
// 2. XDG_DATA_HOME/goatway (Linux)
// 3. ~/.goatway (Linux/macOS) or %APPDATA%\goatway (Windows)
func DataDir() string {
	if dir := os.Getenv("GOATWAY_DATA_DIR"); dir != "" {
		return dir
	}

	// On Linux, respect XDG_DATA_HOME
	if runtime.GOOS == "linux" {
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, "goatway")
		}
	}

	// Windows uses APPDATA
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "goatway")
		}
	}

	// Default to ~/.goatway
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home unavailable
		return ".goatway"
	}
	return filepath.Join(home, ".goatway")
}

// DBPath returns the path to the SQLite database file.
func DBPath() string {
	return filepath.Join(DataDir(), "goatway.db")
}

// ConfigPath returns the path to the optional YAML config file.
func ConfigPath() string {
	return filepath.Join(DataDir(), "config.yaml")
}

// LogDir returns the path to the log directory.
func LogDir() string {
	return filepath.Join(DataDir(), "logs")
}

// EnsureDataDir creates the data directory if it doesn't exist.
func EnsureDataDir() error {
	dir := DataDir()
	return os.MkdirAll(dir, 0700)
}

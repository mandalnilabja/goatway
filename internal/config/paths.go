package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DataDir returns the path to the Goatway data directory.
// - Windows: %APPDATA%\goatway
// - Other OS: ~/.goatway
func DataDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "goatway")
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".goatway"
	}
	return filepath.Join(home, ".goatway")
}

// DBPath returns the path to the SQLite database file.
func DBPath() string {
	return filepath.Join(DataDir(), "goatway.db")
}

// EnsureDataDir creates the data directory if it doesn't exist.
func EnsureDataDir() error {
	return os.MkdirAll(DataDir(), 0700)
}

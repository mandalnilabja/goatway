package config

import "os"

// Config holds application configuration loaded from environment and file.
// Priority: CLI flags → Env vars → config.toml → defaults
type Config struct {
	// ServerPort is the address to bind the server to (e.g., ":8080")
	ServerPort string

	// EnableWebUI enables the web dashboard at /web
	EnableWebUI bool

	// Default routing for unaliased models
	Default *DefaultRoute

	// Models contains model alias mappings
	Models []ModelAlias
}

// Load reads configuration from file and environment variables.
// Environment variables override file config values.
func Load() *Config {
	fileConfig, _ := LoadFile() // Ignore error, use defaults

	return &Config{
		ServerPort:  getEnvOrFile("SERVER_PORT", fileConfig.ServerPort, ":8080"),
		EnableWebUI: getEnvBoolOrFile("ENABLE_WEB_UI", fileConfig.EnableWebUI, true),
		Default:     fileConfig.Default,
		Models:      fileConfig.Models,
	}
}

// getEnvOrFile returns env value, file value, or default (in priority order)
func getEnvOrFile(key, fileValue, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if fileValue != "" {
		return fileValue
	}
	return defaultValue
}

// getEnvBoolOrFile returns env bool, file bool, or default (in priority order)
func getEnvBoolOrFile(key string, fileValue *bool, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	if fileValue != nil {
		return *fileValue
	}
	return defaultValue
}

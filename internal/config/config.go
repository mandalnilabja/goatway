package config

import "os"

// Config holds application configuration loaded from environment variables.
// All LLM provider credentials are stored in the database, not environment variables.
type Config struct {
	// ServerAddr is the address to bind the server to (e.g., ":8080")
	ServerAddr string

	// EnableWebUI enables the web dashboard at /web
	EnableWebUI bool
}

// Load reads configuration from environment variables with sensible defaults.
// Only SERVER_ADDR and ENABLE_WEB_UI are supported as environment variables.
// All other configuration (credentials, API keys) must be set via the admin API.
func Load() *Config {
	return &Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		EnableWebUI: getEnvBool("ENABLE_WEB_UI", true),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

package config

import (
	"os"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	// Server settings
	ServerAddr string
	LogLevel   string
	LogFormat  string // "json" or "text"

	// Provider settings (legacy - prefer credentials from storage)
	Provider         string
	OpenRouterAPIKey string
	OpenAIAPIKey     string
	OpenAIOrg        string
	AzureAPIKey      string
	AzureEndpoint    string
	AnthropicAPIKey  string

	// Storage settings
	DataDir       string // Override for data directory
	EncryptionKey string // Optional encryption key for API keys at rest

	// Security settings
	AdminPassword string // Optional password for admin API access

	// Feature flags
	EnableWebUI bool
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		// Server settings
		ServerAddr: getEnv("SERVER_ADDR", ":8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
		LogFormat:  getEnv("LOG_FORMAT", "text"),

		// Provider settings
		Provider:         getEnv("LLM_PROVIDER", "openrouter"),
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		OpenAIOrg:        getEnv("OPENAI_ORG", ""),
		AzureAPIKey:      getEnv("AZURE_API_KEY", ""),
		AzureEndpoint:    getEnv("AZURE_ENDPOINT", ""),
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),

		// Storage settings
		DataDir:       getEnv("GOATWAY_DATA_DIR", ""),
		EncryptionKey: getEnv("GOATWAY_ENCRYPTION_KEY", ""),

		// Security settings
		AdminPassword: getEnv("GOATWAY_ADMIN_PASSWORD", ""),

		// Feature flags
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

package config

import (
	"os"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	ServerAddr       string
	Provider         string
	OpenRouterAPIKey string
	OpenAIAPIKey     string
	OpenAIOrg        string
	AzureAPIKey      string
	AzureEndpoint    string
	AnthropicAPIKey  string
	LogLevel         string
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		ServerAddr:       getEnv("SERVER_ADDR", ":8080"),
		Provider:         getEnv("LLM_PROVIDER", "openrouter"),
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		OpenAIOrg:        getEnv("OPENAI_ORG", ""),
		AzureAPIKey:      getEnv("AZURE_API_KEY", ""),
		AzureEndpoint:    getEnv("AZURE_ENDPOINT", ""),
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

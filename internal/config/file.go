package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// FileConfig represents the TOML configuration file structure.
type FileConfig struct {
	ServerPort  string        `toml:"server_port"`
	EnableWebUI *bool         `toml:"enable_web_ui"`
	Default     *DefaultRoute `toml:"default"`
	Models      []ModelAlias  `toml:"models"`
}

// DefaultRoute defines the fallback provider and model for unknown slugs.
type DefaultRoute struct {
	Provider       string `toml:"provider"`
	Model          string `toml:"model"`
	CredentialName string `toml:"credential_name"`
}

// ModelAlias maps a short slug to a provider and model combination.
type ModelAlias struct {
	Slug           string `toml:"slug"`
	Provider       string `toml:"provider"`
	Model          string `toml:"model"`
	CredentialName string `toml:"credential_name"`
}

// ConfigPath returns the path to the config file (~/.goatway/config.toml).
func ConfigPath() string {
	return filepath.Join(DataDir(), "config.toml")
}

// LoadFile loads configuration from the TOML file.
// Returns an empty FileConfig if the file doesn't exist.
func LoadFile() (*FileConfig, error) {
	cfg := &FileConfig{}

	path := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// EnsureConfigFile creates a default config file with commented examples if none exists.
func EnsureConfigFile() error {
	path := ConfigPath()

	// If config already exists, do nothing
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// Ensure directory exists
	if err := EnsureDataDir(); err != nil {
		return err
	}

	defaultConfig := `# Goatway Configuration
# server_port = ":8080"
# enable_web_ui = true

# Optional default routing for unaliased models
# [default]
# provider = "openrouter"
# credential_name = "my-openrouter-key"  # Name of credential to use

# Model aliases - map short names to provider/model combinations
# [[models]]
# slug = "gpt4"
# provider = "openrouter"
# model = "openai/gpt-4o"
# credential_name = "my-openrouter-key"  # Required: name of credential to use

# [[models]]
# slug = "claude"
# provider = "openrouter"
# model = "anthropic/claude-3.5-sonnet"
# credential_name = "my-openrouter-key"

# Azure AI Foundry example
# [[models]]
# slug = "deepseek-r1"
# provider = "azurefoundry"
# model = "DeepSeek-R1"
# credential_name = "my-azure-foundry-key"
`

	return os.WriteFile(path, []byte(defaultConfig), 0644)
}

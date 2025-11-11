package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

const (
	// ConfigFileName is the name of the project-level configuration file.
	ConfigFileName = ".tuna.toml"

	// GlobalConfigPath is the path to the user-level configuration file.
	GlobalConfigPath = ".config/tuna.toml"
)

// Environment variable names for backward compatibility.
const (
	EnvAPIToken = "LLM_API_TOKEN"
	EnvBaseURL  = "LLM_BASE_URL"
)

var (
	// ErrNoConfig is returned when no configuration is found.
	ErrNoConfig = errors.New("no configuration found")

	// ErrInvalidConfig is returned when configuration validation fails.
	ErrInvalidConfig = errors.New("invalid configuration")
)

// LoadResult contains the loaded configuration and metadata about the source.
type LoadResult struct {
	Config     *Config
	Source     string // Path to the config file or "environment" for env vars
	Deprecated bool   // True if using deprecated environment variables
}

// Load loads configuration with priority:
// 1. .tuna.toml in current/parent directories
// 2. ~/.config/tuna.toml
// 3. Fallback to env variables (backward compatibility).
func Load() (*LoadResult, error) {
	// Try to find project-level config
	projectPath, err := findConfigFile()
	if err == nil {
		cfg, err := LoadFromFile(projectPath)
		if err != nil {
			return nil, err
		}
		return &LoadResult{
			Config: cfg,
			Source: projectPath,
		}, nil
	}

	// Try user-level config
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, GlobalConfigPath)
		if _, err := os.Stat(globalPath); err == nil {
			cfg, err := LoadFromFile(globalPath)
			if err != nil {
				return nil, err
			}
			return &LoadResult{
				Config: cfg,
				Source: globalPath,
			}, nil
		}
	}

	// Fallback to environment variables (backward compatibility)
	cfg, err := loadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoConfig, err)
	}

	return &LoadResult{
		Config:     cfg,
		Source:     "environment",
		Deprecated: true,
	}, nil
}

// LoadFromFile loads configuration from a specific file.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w in %s:\n%v", ErrInvalidConfig, path, err)
	}

	return &cfg, nil
}

// findConfigFile searches for .tuna.toml up the directory tree.
func findConfigFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("config file %s not found in %s or parent directories", ConfigFileName, cwd)
}

// loadFromEnv creates a configuration from environment variables for backward compatibility.
func loadFromEnv() (*Config, error) {
	token := os.Getenv(EnvAPIToken)
	if token == "" {
		return nil, fmt.Errorf("missing %s environment variable and no config file found\n\nCreate a config file (.tuna.toml) or set environment variables:\n  export %s=your-api-token\n  export %s=https://api.example.com/v1", EnvAPIToken, EnvAPIToken, EnvBaseURL)
	}

	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("missing %s environment variable and no config file found\n\nCreate a config file (.tuna.toml) or set environment variables:\n  export %s=your-api-token\n  export %s=https://api.example.com/v1", EnvBaseURL, EnvAPIToken, EnvBaseURL)
	}

	// Create an implicit "default" provider from environment variables
	return &Config{
		DefaultProvider: "default",
		Providers: []Provider{
			{
				Name:        "default",
				BaseURL:     baseURL,
				APITokenEnv: EnvAPIToken,
				// No rate limit for backward compatibility
			},
		},
	}, nil
}

// FindConfigFile returns the path to the configuration file that would be loaded.
// Returns empty string if no config file exists (only env vars would be used).
func FindConfigFile() (string, error) {
	// Try project-level config
	projectPath, err := findConfigFile()
	if err == nil {
		return projectPath, nil
	}

	// Try user-level config
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, GlobalConfigPath)
		if _, err := os.Stat(globalPath); err == nil {
			return globalPath, nil
		}
	}

	return "", ErrNoConfig
}

// DeprecationWarning returns a warning message about deprecated configuration.
func DeprecationWarning() string {
	return fmt.Sprintf(`Warning: Using deprecated environment variables (%s, %s).

Consider creating a configuration file for better flexibility:

  # .tuna.toml
  default_provider = "default"

  [[providers]]
  name = "default"
  base_url = "$%s"
  api_token_env = "%s"

See documentation for more examples.
`, EnvAPIToken, EnvBaseURL, EnvBaseURL, EnvAPIToken)
}

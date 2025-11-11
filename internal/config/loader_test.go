package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// evalSymlinks resolves symlinks for path comparison (handles macOS /var -> /private/var).
func evalSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}
	return resolved
}

func TestLoadFromFile(t *testing.T) {
	t.Run("loads valid config", func(t *testing.T) {
		content := `
default_provider = "openrouter"

[aliases]
sonnet = "claude-sonnet-4-20250514"
gpt4 = "gpt-4o"

[[providers]]
name = "openrouter"
base_url = "https://openrouter.ai/api/v1"
api_token_env = "OPENROUTER_API_KEY"
rate_limit = "10rpm"
models = ["anthropic/claude-sonnet-4", "openai/gpt-4o"]

[[providers]]
name = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_token_env = "ANTHROPIC_API_KEY"
models = ["claude-sonnet-4-20250514"]
`
		tmpFile := createTempConfig(t, content)

		cfg, err := LoadFromFile(tmpFile)
		require.NoError(t, err)

		assert.Equal(t, "openrouter", cfg.DefaultProvider)
		assert.Equal(t, "claude-sonnet-4-20250514", cfg.Aliases["sonnet"])
		assert.Equal(t, "gpt-4o", cfg.Aliases["gpt4"])
		assert.Len(t, cfg.Providers, 2)

		assert.Equal(t, "openrouter", cfg.Providers[0].Name)
		assert.Equal(t, "https://openrouter.ai/api/v1", cfg.Providers[0].BaseURL)
		assert.Equal(t, "OPENROUTER_API_KEY", cfg.Providers[0].APITokenEnv)
		assert.Equal(t, "10rpm", cfg.Providers[0].RateLimit)
		assert.Equal(t, []string{"anthropic/claude-sonnet-4", "openai/gpt-4o"}, cfg.Providers[0].Models)

		assert.Equal(t, "anthropic", cfg.Providers[1].Name)
		assert.Empty(t, cfg.Providers[1].RateLimit)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := LoadFromFile("/non/existent/path.toml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		content := `
default_provider = "openrouter
this is not valid toml
`
		tmpFile := createTempConfig(t, content)

		_, err := LoadFromFile(tmpFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("returns error for invalid config", func(t *testing.T) {
		content := `
# Missing default_provider
[[providers]]
name = "test"
base_url = "https://test.com"
api_token_env = "TEST_KEY"
`
		tmpFile := createTempConfig(t, content)

		_, err := LoadFromFile(tmpFile)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConfig)
	})

	t.Run("loads minimal valid config", func(t *testing.T) {
		content := `
default_provider = "test"

[[providers]]
name = "test"
base_url = "https://test.com/v1"
api_token_env = "TEST_KEY"
`
		tmpFile := createTempConfig(t, content)

		cfg, err := LoadFromFile(tmpFile)
		require.NoError(t, err)
		assert.Equal(t, "test", cfg.DefaultProvider)
		assert.Len(t, cfg.Providers, 1)
	})
}

func TestLoad(t *testing.T) {
	t.Run("loads from project directory", func(t *testing.T) {
		// Create a temp directory with a config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ConfigFileName)
		content := `
default_provider = "test"

[[providers]]
name = "test"
base_url = "https://test.com/v1"
api_token_env = "TEST_KEY"
`
		err := os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		// Change to the temp directory
		origDir, _ := os.Getwd()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		result, err := Load()
		require.NoError(t, err)
		assert.Equal(t, evalSymlinks(t, configPath), evalSymlinks(t, result.Source))
		assert.False(t, result.Deprecated)
		assert.Equal(t, "test", result.Config.DefaultProvider)
	})

	t.Run("loads from parent directory", func(t *testing.T) {
		// Create a temp directory structure: parent/child
		parentDir := t.TempDir()
		childDir := filepath.Join(parentDir, "child")
		err := os.Mkdir(childDir, 0755)
		require.NoError(t, err)

		// Put config in parent
		configPath := filepath.Join(parentDir, ConfigFileName)
		content := `
default_provider = "parent"

[[providers]]
name = "parent"
base_url = "https://parent.com/v1"
api_token_env = "PARENT_KEY"
`
		err = os.WriteFile(configPath, []byte(content), 0644)
		require.NoError(t, err)

		// Change to child directory
		origDir, _ := os.Getwd()
		err = os.Chdir(childDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		result, err := Load()
		require.NoError(t, err)
		assert.Equal(t, evalSymlinks(t, configPath), evalSymlinks(t, result.Source))
		assert.Equal(t, "parent", result.Config.DefaultProvider)
	})

	t.Run("project config takes precedence over parent", func(t *testing.T) {
		// Create a temp directory structure: parent/child
		parentDir := t.TempDir()
		childDir := filepath.Join(parentDir, "child")
		err := os.Mkdir(childDir, 0755)
		require.NoError(t, err)

		// Put config in parent
		parentConfigPath := filepath.Join(parentDir, ConfigFileName)
		parentContent := `
default_provider = "parent"

[[providers]]
name = "parent"
base_url = "https://parent.com/v1"
api_token_env = "PARENT_KEY"
`
		err = os.WriteFile(parentConfigPath, []byte(parentContent), 0644)
		require.NoError(t, err)

		// Put config in child (should take precedence)
		childConfigPath := filepath.Join(childDir, ConfigFileName)
		childContent := `
default_provider = "child"

[[providers]]
name = "child"
base_url = "https://child.com/v1"
api_token_env = "CHILD_KEY"
`
		err = os.WriteFile(childConfigPath, []byte(childContent), 0644)
		require.NoError(t, err)

		// Change to child directory
		origDir, _ := os.Getwd()
		err = os.Chdir(childDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		result, err := Load()
		require.NoError(t, err)
		assert.Equal(t, evalSymlinks(t, childConfigPath), evalSymlinks(t, result.Source))
		assert.Equal(t, "child", result.Config.DefaultProvider)
	})

	t.Run("falls back to env vars", func(t *testing.T) {
		// Create a temp directory without any config
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		err := os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		// Set env vars
		t.Setenv(EnvAPIToken, "test-token")
		t.Setenv(EnvBaseURL, "https://test.com/v1")

		// Ensure no global config exists
		home := t.TempDir()
		t.Setenv("HOME", home)

		result, err := Load()
		require.NoError(t, err)
		assert.Equal(t, "environment", result.Source)
		assert.True(t, result.Deprecated)
		assert.Equal(t, "default", result.Config.DefaultProvider)
		require.Len(t, result.Config.Providers, 1)
		assert.Equal(t, "https://test.com/v1", result.Config.Providers[0].BaseURL)
	})

	t.Run("returns error when no config and no env vars", func(t *testing.T) {
		// Create a temp directory without any config
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		err := os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		// Ensure no env vars
		os.Unsetenv(EnvAPIToken)
		os.Unsetenv(EnvBaseURL)

		// Ensure no global config
		home := t.TempDir()
		t.Setenv("HOME", home)

		_, err = Load()
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoConfig)
	})
}

func TestFindConfigFile(t *testing.T) {
	t.Run("finds config in current directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ConfigFileName)
		err := os.WriteFile(configPath, []byte(""), 0644)
		require.NoError(t, err)

		origDir, _ := os.Getwd()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		path, err := FindConfigFile()
		require.NoError(t, err)
		assert.Equal(t, evalSymlinks(t, configPath), evalSymlinks(t, path))
	})

	t.Run("returns error when no config found", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		err := os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		// Ensure no global config
		home := t.TempDir()
		t.Setenv("HOME", home)

		_, err = FindConfigFile()
		require.ErrorIs(t, err, ErrNoConfig)
	})
}

func TestDeprecationWarning(t *testing.T) {
	warning := DeprecationWarning()
	assert.Contains(t, warning, EnvAPIToken)
	assert.Contains(t, warning, EnvBaseURL)
	assert.Contains(t, warning, "deprecated")
	assert.Contains(t, warning, ".tuna.toml")
}

func TestLoadFromEnv(t *testing.T) {
	t.Run("creates config from env vars", func(t *testing.T) {
		t.Setenv(EnvAPIToken, "test-token")
		t.Setenv(EnvBaseURL, "https://api.test.com/v1")

		cfg, err := loadFromEnv()
		require.NoError(t, err)

		assert.Equal(t, "default", cfg.DefaultProvider)
		require.Len(t, cfg.Providers, 1)
		assert.Equal(t, "default", cfg.Providers[0].Name)
		assert.Equal(t, "https://api.test.com/v1", cfg.Providers[0].BaseURL)
		assert.Equal(t, EnvAPIToken, cfg.Providers[0].APITokenEnv)
		assert.Empty(t, cfg.Providers[0].RateLimit)
	})

	t.Run("returns error for missing token", func(t *testing.T) {
		os.Unsetenv(EnvAPIToken)
		t.Setenv(EnvBaseURL, "https://api.test.com/v1")

		_, err := loadFromEnv()
		require.Error(t, err)
		assert.Contains(t, err.Error(), EnvAPIToken)
	})

	t.Run("returns error for missing base URL", func(t *testing.T) {
		t.Setenv(EnvAPIToken, "test-token")
		os.Unsetenv(EnvBaseURL)

		_, err := loadFromEnv()
		require.Error(t, err)
		assert.Contains(t, err.Error(), EnvBaseURL)
	})
}

func TestLoadGlobalConfig(t *testing.T) {
	t.Run("loads from global config path", func(t *testing.T) {
		// Create a temp directory to serve as HOME
		home := t.TempDir()
		t.Setenv("HOME", home)

		// Create .config directory
		configDir := filepath.Join(home, ".config")
		err := os.Mkdir(configDir, 0755)
		require.NoError(t, err)

		// Create global config
		globalConfigPath := filepath.Join(home, GlobalConfigPath)
		content := `
default_provider = "global"

[[providers]]
name = "global"
base_url = "https://global.com/v1"
api_token_env = "GLOBAL_KEY"
`
		err = os.WriteFile(globalConfigPath, []byte(content), 0644)
		require.NoError(t, err)

		// Create a working directory without a local config
		workDir := t.TempDir()
		origDir, _ := os.Getwd()
		err = os.Chdir(workDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		result, err := Load()
		require.NoError(t, err)
		assert.Equal(t, globalConfigPath, result.Source)
		assert.Equal(t, "global", result.Config.DefaultProvider)
	})

	t.Run("project config takes precedence over global", func(t *testing.T) {
		// Create a temp directory to serve as HOME
		home := t.TempDir()
		t.Setenv("HOME", home)

		// Create global config
		configDir := filepath.Join(home, ".config")
		err := os.Mkdir(configDir, 0755)
		require.NoError(t, err)
		globalConfigPath := filepath.Join(home, GlobalConfigPath)
		globalContent := `
default_provider = "global"

[[providers]]
name = "global"
base_url = "https://global.com/v1"
api_token_env = "GLOBAL_KEY"
`
		err = os.WriteFile(globalConfigPath, []byte(globalContent), 0644)
		require.NoError(t, err)

		// Create a working directory with a local config
		workDir := t.TempDir()
		localConfigPath := filepath.Join(workDir, ConfigFileName)
		localContent := `
default_provider = "local"

[[providers]]
name = "local"
base_url = "https://local.com/v1"
api_token_env = "LOCAL_KEY"
`
		err = os.WriteFile(localConfigPath, []byte(localContent), 0644)
		require.NoError(t, err)

		origDir, _ := os.Getwd()
		err = os.Chdir(workDir)
		require.NoError(t, err)
		defer os.Chdir(origDir)

		result, err := Load()
		require.NoError(t, err)
		assert.Equal(t, evalSymlinks(t, localConfigPath), evalSymlinks(t, result.Source))
		assert.Equal(t, "local", result.Config.DefaultProvider)
	})
}

// createTempConfig is a helper that creates a temporary config file with the given content.
func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.toml")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)
	return tmpFile
}

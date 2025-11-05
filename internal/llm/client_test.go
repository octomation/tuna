package llm

import (
	"os"
	"testing"
)

func TestConfigFromEnv(t *testing.T) {
	t.Run("returns config when vars set", func(t *testing.T) {
		t.Setenv(EnvAPIToken, "test-token")
		t.Setenv(EnvBaseURL, "https://api.example.com")

		cfg, err := ConfigFromEnv()
		if err != nil {
			t.Fatalf("ConfigFromEnv() error = %v", err)
		}

		if cfg.APIToken != "test-token" {
			t.Errorf("APIToken = %q, want %q", cfg.APIToken, "test-token")
		}
		if cfg.BaseURL != "https://api.example.com" {
			t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://api.example.com")
		}
	})

	t.Run("returns error for missing token", func(t *testing.T) {
		os.Unsetenv(EnvAPIToken)
		t.Setenv(EnvBaseURL, "https://api.example.com")

		_, err := ConfigFromEnv()
		if err == nil {
			t.Error("Expected error for missing API token")
		}
	})

	t.Run("returns error for missing base URL", func(t *testing.T) {
		t.Setenv(EnvAPIToken, "test-token")
		os.Unsetenv(EnvBaseURL)

		_, err := ConfigFromEnv()
		if err == nil {
			t.Error("Expected error for missing base URL")
		}
	})
}

func TestNewClient(t *testing.T) {
	cfg := &Config{
		APIToken: "test-token",
		BaseURL:  "https://api.example.com/v1",
	}

	client := NewClient(cfg)
	if client == nil {
		t.Error("NewClient() returned nil")
	}
	if client.client == nil {
		t.Error("NewClient().client is nil")
	}
}

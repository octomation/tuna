package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRateLimit(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue int
		wantUnit  time.Duration
		wantNil   bool
		wantErr   bool
	}{
		{
			name:      "10 requests per minute",
			input:     "10rpm",
			wantValue: 10,
			wantUnit:  time.Minute,
		},
		{
			name:      "5 requests per second",
			input:     "5rps",
			wantValue: 5,
			wantUnit:  time.Second,
		},
		{
			name:      "100 requests per hour",
			input:     "100rph",
			wantValue: 100,
			wantUnit:  time.Hour,
		},
		{
			name:      "1 request per minute",
			input:     "1rpm",
			wantValue: 1,
			wantUnit:  time.Minute,
		},
		{
			name:    "empty string returns nil",
			input:   "",
			wantNil: true,
		},
		{
			name:    "invalid format - no number",
			input:   "rpm",
			wantErr: true,
		},
		{
			name:    "invalid format - no unit",
			input:   "10",
			wantErr: true,
		},
		{
			name:    "invalid format - unknown unit",
			input:   "10rpd",
			wantErr: true,
		},
		{
			name:    "invalid format - spaces",
			input:   "10 rpm",
			wantErr: true,
		},
		{
			name:    "invalid format - negative",
			input:   "-10rpm",
			wantErr: true,
		},
		{
			name:    "zero value",
			input:   "0rpm",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRateLimit(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tt.wantValue, got.Value)
			assert.Equal(t, tt.wantUnit, got.Unit)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openrouter",
			Aliases: map[string]string{
				"sonnet": "claude-sonnet-4-20250514",
				"gpt4":   "gpt-4o",
			},
			Providers: []Provider{
				{
					Name:        "openrouter",
					BaseURL:     "https://openrouter.ai/api/v1",
					APITokenEnv: "OPENROUTER_API_KEY",
					RateLimit:   "10rpm",
					Models:      []string{"anthropic/claude-sonnet-4", "openai/gpt-4o"},
				},
				{
					Name:        "anthropic",
					BaseURL:     "https://api.anthropic.com/v1",
					APITokenEnv: "ANTHROPIC_API_KEY",
					RateLimit:   "60rpm",
					Models:      []string{"claude-sonnet-4-20250514"},
				},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid configuration without rate limit", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Providers: []Provider{
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
					Models:      []string{"gpt-4o"},
				},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing default_provider", func(t *testing.T) {
		cfg := &Config{
			Providers: []Provider{
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_provider is required")
	})

	t.Run("no providers", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Providers:       []Provider{},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one provider is required")
	})

	t.Run("duplicate provider names", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Providers: []Provider{
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
				},
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY_2",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate provider name")
	})

	t.Run("default provider not in list", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "unknown",
			Providers: []Provider{
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "default_provider \"unknown\" not found")
	})

	t.Run("provider missing name", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Providers: []Provider{
				{
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("provider missing base_url", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Providers: []Provider{
				{
					Name:        "openai",
					APITokenEnv: "OPENAI_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url is required")
	})

	t.Run("provider missing api_token_env", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Providers: []Provider{
				{
					Name:    "openai",
					BaseURL: "https://api.openai.com/v1",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "api_token_env is required")
	})

	t.Run("provider invalid rate limit", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Providers: []Provider{
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
					RateLimit:   "invalid",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid rate limit format")
	})

	t.Run("alias with empty key", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Aliases: map[string]string{
				"": "gpt-4o",
			},
			Providers: []Provider{
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "alias key cannot be empty")
	})

	t.Run("alias with empty model", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "openai",
			Aliases: map[string]string{
				"gpt4": "",
			},
			Providers: []Provider{
				{
					Name:        "openai",
					BaseURL:     "https://api.openai.com/v1",
					APITokenEnv: "OPENAI_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model name cannot be empty")
	})

	t.Run("multiple errors collected", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "",
			Providers:       []Provider{},
		}

		err := cfg.Validate()
		require.Error(t, err)
		// Should contain both errors
		assert.Contains(t, err.Error(), "default_provider is required")
		assert.Contains(t, err.Error(), "at least one provider is required")
	})
}

func TestConfig_ValidateProviders(t *testing.T) {
	t.Run("valid minimal configuration", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "test",
			Providers: []Provider{
				{
					Name:        "test",
					BaseURL:     "https://api.test.com/v1",
					APITokenEnv: "TEST_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid configuration with empty aliases", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "test",
			Aliases:         map[string]string{},
			Providers: []Provider{
				{
					Name:        "test",
					BaseURL:     "https://api.test.com/v1",
					APITokenEnv: "TEST_API_KEY",
				},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid configuration with empty models", func(t *testing.T) {
		cfg := &Config{
			DefaultProvider: "test",
			Providers: []Provider{
				{
					Name:        "test",
					BaseURL:     "https://api.test.com/v1",
					APITokenEnv: "TEST_API_KEY",
					Models:      []string{},
				},
			},
		}

		err := cfg.Validate()
		assert.NoError(t, err)
	})
}

func TestRateLimitRegex(t *testing.T) {
	// Test that the regex is properly anchored
	tests := []struct {
		input   string
		matches bool
	}{
		{"10rpm", true},
		{"5rps", true},
		{"100rph", true},
		{"10rpm extra", false},
		{"prefix 10rpm", false},
		{"10RPM", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := rateLimitRegex.MatchString(tt.input)
			if got != tt.matches {
				t.Errorf("rateLimitRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.matches)
			}
		})
	}
}

func TestParseRateLimitEdgeCases(t *testing.T) {
	t.Run("large value", func(t *testing.T) {
		got, err := ParseRateLimit("999999rpm")
		require.NoError(t, err)
		assert.Equal(t, 999999, got.Value)
		assert.Equal(t, time.Minute, got.Unit)
	})

	t.Run("value with leading zeros", func(t *testing.T) {
		got, err := ParseRateLimit("007rpm")
		require.NoError(t, err)
		assert.Equal(t, 7, got.Value)
	})
}

func TestConfig_ValidateMultipleErrors(t *testing.T) {
	cfg := &Config{
		DefaultProvider: "unknown",
		Aliases: map[string]string{
			"":    "model1",
			"m2":  "",
			"gpt": "gpt-4o",
		},
		Providers: []Provider{
			{
				Name: "provider1",
				// missing base_url and api_token_env
			},
			{
				Name:        "provider1", // duplicate
				BaseURL:     "https://test.com",
				APITokenEnv: "TEST_KEY",
				RateLimit:   "invalid",
			},
		},
	}

	err := cfg.Validate()
	require.Error(t, err)

	errStr := err.Error()
	// Count how many distinct errors we got
	errorCount := strings.Count(errStr, "\n") + 1

	// We expect multiple errors to be collected:
	// - alias key cannot be empty
	// - alias model name cannot be empty
	// - provider missing base_url
	// - provider missing api_token_env
	// - duplicate provider name
	// - invalid rate limit format
	// - default provider not found
	assert.GreaterOrEqual(t, errorCount, 5, "expected at least 5 errors, got: %s", errStr)
}

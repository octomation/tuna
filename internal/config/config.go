package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

// Config represents the root tuna configuration.
type Config struct {
	DefaultProvider string            `toml:"default_provider"`
	Aliases         map[string]string `toml:"aliases"`
	Providers       []Provider        `toml:"providers"`
}

// Provider describes a single LLM provider configuration.
type Provider struct {
	Name        string   `toml:"name"`
	BaseURL     string   `toml:"base_url"`
	APIToken    string   `toml:"api_token"`     // Direct token value
	APITokenEnv string   `toml:"api_token_env"` // Environment variable reference
	RateLimit   string   `toml:"rate_limit"`
	Models      []string `toml:"models"`
}

// ResolveAPIToken returns the API token using priority:
// 1. Direct api_token value
// 2. Value from api_token_env environment variable
// Returns error if no token is available.
func (p *Provider) ResolveAPIToken() (string, error) {
	if p.APIToken != "" {
		return p.APIToken, nil
	}
	if p.APITokenEnv != "" {
		if token := os.Getenv(p.APITokenEnv); token != "" {
			return token, nil
		}
		return "", fmt.Errorf("environment variable %q is not set", p.APITokenEnv)
	}
	return "", errors.New("neither api_token nor api_token_env is specified")
}

// RateLimit represents a parsed rate limit value.
type RateLimit struct {
	Value int           // Number of requests
	Unit  time.Duration // Per unit of time (time.Second, time.Minute, time.Hour)
}

// rateLimitRegex matches rate limit strings like "10rpm", "5rps", "100rph".
var rateLimitRegex = regexp.MustCompile(`^(\d+)(rps|rpm|rph)$`)

// ParseRateLimit parses rate limit string like "10rpm", "5rps", "100rph".
// Supported units: rps (per second), rpm (per minute), rph (per hour).
// Returns nil if empty string (unlimited).
func ParseRateLimit(s string) (*RateLimit, error) {
	if s == "" {
		return nil, nil
	}

	matches := rateLimitRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid rate limit format %q: expected format like '10rpm', '5rps', or '100rph'", s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit value: %w", err)
	}

	if value <= 0 {
		return nil, fmt.Errorf("rate limit value must be positive, got %d", value)
	}

	var unit time.Duration
	switch matches[2] {
	case "rps":
		unit = time.Second
	case "rpm":
		unit = time.Minute
	case "rph":
		unit = time.Hour
	default:
		return nil, fmt.Errorf("unknown rate limit unit %q", matches[2])
	}

	return &RateLimit{
		Value: value,
		Unit:  unit,
	}, nil
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	var errs []error

	if c.DefaultProvider == "" {
		errs = append(errs, errors.New("default_provider is required"))
	}

	if len(c.Providers) == 0 {
		errs = append(errs, errors.New("at least one provider is required"))
	}

	// Check for duplicate provider names
	providerNames := make(map[string]bool)
	defaultProviderFound := false

	for i, p := range c.Providers {
		if p.Name == "" {
			errs = append(errs, fmt.Errorf("provider[%d]: name is required", i))
			continue
		}

		if providerNames[p.Name] {
			errs = append(errs, fmt.Errorf("provider[%d]: duplicate provider name %q", i, p.Name))
		}
		providerNames[p.Name] = true

		if p.Name == c.DefaultProvider {
			defaultProviderFound = true
		}

		if p.BaseURL == "" {
			errs = append(errs, fmt.Errorf("provider[%d] %q: base_url is required", i, p.Name))
		}

		if p.APIToken == "" && p.APITokenEnv == "" {
			errs = append(errs, fmt.Errorf("provider[%d] %q: either api_token or api_token_env is required", i, p.Name))
		}

		if p.RateLimit != "" {
			if _, err := ParseRateLimit(p.RateLimit); err != nil {
				errs = append(errs, fmt.Errorf("provider[%d] %q: %w", i, p.Name, err))
			}
		}
	}

	if c.DefaultProvider != "" && len(c.Providers) > 0 && !defaultProviderFound {
		errs = append(errs, fmt.Errorf("default_provider %q not found in providers list", c.DefaultProvider))
	}

	// Validate aliases reference valid model names (optional: just check format)
	for alias, model := range c.Aliases {
		if alias == "" {
			errs = append(errs, errors.New("alias key cannot be empty"))
		}
		if model == "" {
			errs = append(errs, fmt.Errorf("alias %q: model name cannot be empty", alias))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

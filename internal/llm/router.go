package llm

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	"go.octolab.org/toolset/tuna/internal/config"
)

// Router routes requests to appropriate providers based on model name.
type Router struct {
	providers       map[string]*Client       // name -> client
	providerURLs    map[string]string        // name -> base URL
	rateLimiters    map[string]*rate.Limiter // name -> rate limiter
	aliases         map[string]string        // alias -> full model name
	modelMapping    map[string]string        // model -> provider name
	defaultProvider string
}

// Compile-time interface implementation check.
var _ ChatClient = (*Router)(nil)

// NewRouter creates a router from configuration.
func NewRouter(cfg *config.Config) (*Router, error) {
	r := &Router{
		providers:       make(map[string]*Client),
		providerURLs:    make(map[string]string),
		rateLimiters:    make(map[string]*rate.Limiter),
		aliases:         cfg.Aliases,
		modelMapping:    make(map[string]string),
		defaultProvider: cfg.DefaultProvider,
	}

	if r.aliases == nil {
		r.aliases = make(map[string]string)
	}

	// Create clients and rate limiters for each provider
	for _, p := range cfg.Providers {
		// Resolve API token (direct value or from environment)
		token, err := p.ResolveAPIToken()
		if err != nil {
			return nil, fmt.Errorf("provider %q: %w", p.Name, err)
		}

		// Create client
		client := NewClient(&Config{
			APIToken: token,
			BaseURL:  p.BaseURL,
		})
		r.providers[p.Name] = client
		r.providerURLs[p.Name] = p.BaseURL

		// Create rate limiter if configured
		if p.RateLimit != "" {
			rl, err := config.ParseRateLimit(p.RateLimit)
			if err != nil {
				return nil, fmt.Errorf("provider %q: %w", p.Name, err)
			}
			if rl != nil {
				// rate.Every returns the duration between events
				// For "10rpm", we want 1 request every 6 seconds
				limiter := rate.NewLimiter(rate.Every(rl.Unit/time.Duration(rl.Value)), 1)
				r.rateLimiters[p.Name] = limiter
			}
		}

		// Build model to provider mapping
		for _, model := range p.Models {
			r.modelMapping[model] = p.Name
		}
	}

	return r, nil
}

// Chat sends a request to the appropriate provider.
func (r *Router) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Resolve alias to full model name
	resolvedModel := r.resolveAlias(req.Model)

	// Find the provider for this model
	providerName := r.resolveProvider(resolvedModel)

	client, ok := r.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %q not found for model %q", providerName, req.Model)
	}

	providerURL := r.providerURLs[providerName]

	// Wait for rate limiter if configured
	if limiter, ok := r.rateLimiters[providerName]; ok {
		if err := limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit wait cancelled: %w", err)
		}
	}

	// Update request with resolved model name
	req.Model = resolvedModel

	// Time the actual API request (excluding rate limit wait)
	start := time.Now()
	resp, err := client.Chat(ctx, req)
	duration := time.Since(start)

	if err != nil {
		return nil, err
	}

	// Add provider URL and timing to response
	resp.ProviderURL = providerURL
	resp.Duration = duration

	return resp, nil
}

// resolveAlias resolves an alias to the full model name.
func (r *Router) resolveAlias(model string) string {
	if fullName, ok := r.aliases[model]; ok {
		return fullName
	}
	return model
}

// resolveProvider determines the provider for a model.
func (r *Router) resolveProvider(model string) string {
	if provider, ok := r.modelMapping[model]; ok {
		return provider
	}
	return r.defaultProvider
}

// ResolveModel returns full model name and provider name for a given model or alias.
// This is useful for CLI commands like "tuna config resolve <model>".
func (r *Router) ResolveModel(model string) (fullName, provider string) {
	fullName = r.resolveAlias(model)
	provider = r.resolveProvider(fullName)
	return fullName, provider
}

// Providers returns the list of provider names.
func (r *Router) Providers() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Aliases returns a copy of the aliases map.
func (r *Router) Aliases() map[string]string {
	result := make(map[string]string, len(r.aliases))
	for k, v := range r.aliases {
		result[k] = v
	}
	return result
}

// DefaultProvider returns the name of the default provider.
func (r *Router) DefaultProvider() string {
	return r.defaultProvider
}

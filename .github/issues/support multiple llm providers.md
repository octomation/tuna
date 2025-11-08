---
issue: null
status: open
type: task
labels:
  - "effort: medium"
  - "impact: high"
  - "scope: code"
  - "type: feature"
assignees:
  - kamilsk
milestone: null
projects: []
relationships:
  parent: null
  blocked_by: []
  blocks: []
---

# Support Multiple LLM Providers

## Summary

Implement support for multiple LLM providers (OpenRouter, Anthropic, OpenAI, etc.) with a unified client that routes requests to the appropriate provider based on model name.

## Motivation

Currently, tuna supports only a single LLM provider configured via environment variables (`LLM_API_TOKEN`, `LLM_BASE_URL`). This limits the ability to:

- Compare responses from models hosted on different providers in a single execution
- Use specialized providers for specific models (e.g., Anthropic API for Claude, OpenAI API for GPT)
- Leverage cost-effective aggregators like OpenRouter alongside direct provider access

## Current State

```go
// internal/llm/client.go
type Config struct {
    APIToken string
    BaseURL  string
}

func ConfigFromEnv() (*Config, error) {
    token := os.Getenv("LLM_API_TOKEN")
    baseURL := os.Getenv("LLM_BASE_URL")
    // ...
}
```

**Limitations:**
- Single provider per execution
- Environment variables don't scale for multiple providers
- No model-to-provider mapping

## Proposed Solution

### 1. Configuration File

Replace environment variables with a TOML configuration file (`~/.config/tuna/config.toml` or `.tuna.toml` in project root):

```toml
# Default provider used when model is not found in any provider's model list
default_provider = "openrouter"

# Model aliases for convenience (short name -> full model name)
[aliases]
sonnet = "claude-sonnet-4-20250514"
haiku = "claude-haiku-3-5-20241022"
gpt4 = "gpt-4o"
gpt4-mini = "gpt-4o-mini"
llama = "meta-llama/llama-3.3-70b-instruct"

[[providers]]
name = "openrouter"
base_url = "https://openrouter.ai/api/v1"
api_token_env = "OPENROUTER_API_KEY"  # Reference to env variable
rate_limit = "10rpm"  # 10 requests per minute
models = [
    "anthropic/claude-sonnet-4",
    "openai/gpt-4o",
    "google/gemini-2.0-flash",
    "meta-llama/llama-3.3-70b-instruct",
]

[[providers]]
name = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_token_env = "ANTHROPIC_API_KEY"
rate_limit = "60rpm"  # Anthropic tier-based limit
models = [
    "claude-sonnet-4-20250514",
    "claude-haiku-3-5-20241022",
]

[[providers]]
name = "openai"
base_url = "https://api.openai.com/v1"
api_token_env = "OPENAI_API_KEY"
rate_limit = "500rpm"  # OpenAI tier-based limit
models = [
    "gpt-4o",
    "gpt-4o-mini",
    "o1",
]
```

### 2. Configuration Structure

```go
// internal/config/config.go

type Config struct {
    DefaultProvider string            `toml:"default_provider"`
    Aliases         map[string]string `toml:"aliases"`  // short name -> full model name
    Providers       []Provider        `toml:"providers"`
}

type Provider struct {
    Name        string   `toml:"name"`
    BaseURL     string   `toml:"base_url"`
    APITokenEnv string   `toml:"api_token_env"`
    RateLimit   string   `toml:"rate_limit"` // e.g., "10rpm", "5rps" (empty = unlimited)
    Models      []string `toml:"models"`
}

// RateLimit represents a parsed rate limit value.
type RateLimit struct {
    Value int           // Number of requests
    Unit  time.Duration // Per unit of time (time.Second, time.Minute, time.Hour)
}

// ParseRateLimit parses rate limit string like "10rpm", "5rps", "100rph".
// Supported units: rps (per second), rpm (per minute), rph (per hour).
func ParseRateLimit(s string) (*RateLimit, error) {
    // Parse format: "<number><unit>" (e.g., "10rpm")
    // Returns nil if empty string (unlimited)
}
```

### 3. Router Client

Create a unified client that routes requests to the appropriate provider:

```go
// internal/llm/router.go

type Router struct {
    providers       map[string]*Client        // name -> client
    rateLimiters    map[string]*rate.Limiter  // name -> rate limiter
    aliases         map[string]string         // alias -> full model name
    modelMapping    map[string]string         // model -> provider name
    defaultProvider string
}

func NewRouter(cfg *config.Config) (*Router, error) {
    // Build provider clients, rate limiters, and model mapping
    // For each provider with rate_limit set:
    //   limit := config.ParseRateLimit(provider.RateLimit)
    //   rate.NewLimiter(rate.Every(limit.Unit/limit.Value), 1)
}

func (r *Router) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    provider := r.resolveProvider(req.Model)

    // Wait for rate limiter if configured
    if limiter, ok := r.rateLimiters[provider]; ok {
        if err := limiter.Wait(ctx); err != nil {
            return nil, fmt.Errorf("rate limit wait cancelled: %w", err)
        }
    }

    client := r.providers[provider]
    return client.Chat(ctx, req)
}

func (r *Router) resolveProvider(model string) string {
    // First, resolve alias to full model name
    resolvedModel := r.resolveAlias(model)

    if provider, ok := r.modelMapping[resolvedModel]; ok {
        return provider
    }
    return r.defaultProvider
}

func (r *Router) resolveAlias(model string) string {
    if fullName, ok := r.aliases[model]; ok {
        return fullName
    }
    return model  // Return as-is if not an alias
}
```

### 4. Configuration Loading Priority

1. Project-level: `.tuna.toml` in current directory or parent directories
2. User-level: `~/.config/tuna/config.toml`
3. Fallback: Environment variables (backward compatibility)

### 5. CLI Changes

Add `config` subcommand for configuration management:

```bash
# Show current configuration
tuna config show

# Validate configuration
tuna config validate

# Show which provider will be used for a model
tuna config resolve <model>
```

## Implementation Plan

### Phase 1: Configuration Infrastructure

| File                             | Action | Description                              |
|----------------------------------|--------|------------------------------------------|
| `internal/config/config.go`      | Create | Configuration structures                 |
| `internal/config/loader.go`      | Create | TOML loading with priority resolution    |
| `internal/config/loader_test.go` | Create | Configuration loading tests              |

### Phase 2: Router Client

| File                           | Action | Description                                |
|--------------------------------|--------|--------------------------------------------|
| `internal/llm/router.go`       | Create | Multi-provider router implementation       |
| `internal/llm/router_test.go`  | Create | Router tests with mocked providers         |
| `internal/llm/client.go`       | Modify | Extract interface, keep as single provider |

### Phase 3: Integration

| File                           | Action | Description                                |
|--------------------------------|--------|--------------------------------------------|
| `internal/command/exec.go`     | Modify | Use Router instead of single Client        |
| `internal/exec/executor.go`    | Modify | Accept Router interface                    |
| `internal/command/config.go`   | Create | Config subcommand implementation           |

### Phase 4: Documentation & Migration

| File                           | Action | Description                                |
|--------------------------------|--------|--------------------------------------------|
| `CLAUDE.md`                    | Modify | Update configuration docs                  |
| `README.md`                    | Modify | Add multi-provider setup guide             |

## Interface Design

```go
// internal/llm/interface.go

// ChatClient defines the interface for LLM chat operations.
type ChatClient interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// Ensure both Client and Router implement ChatClient
var _ ChatClient = (*Client)(nil)
var _ ChatClient = (*Router)(nil)
```

## Error Handling

1. **Missing provider for model**: Use default provider with warning log
2. **Missing API token**: Clear error message with env variable name
3. **Invalid configuration**: Validation errors on load with line numbers
4. **Provider unavailable**: Retry with exponential backoff (existing behavior)
5. **Rate limit exceeded**: Block and wait until the next request slot is available (using `golang.org/x/time/rate`)

## Backward Compatibility

- If no config file exists and `LLM_API_TOKEN`/`LLM_BASE_URL` are set, create an implicit "default" provider
- Deprecation warning when using environment variables only
- Migration guide in documentation

## Example Usage

```bash
# Plan with aliases - much shorter than full model names
tuna plan MyAssistant --models "sonnet,gpt4,llama"

# Or mix aliases with full names
tuna plan MyAssistant \
    --models "sonnet,gpt-4o-mini,anthropic/claude-sonnet-4"

# Execute - router resolves aliases and routes to correct providers
tuna exec abc123

# Check alias and provider resolution
tuna config resolve sonnet
# Output: claude-sonnet-4-20250514 -> anthropic

tuna config resolve gpt4
# Output: gpt-4o -> openai

tuna config resolve unknown-model
# Output: unknown-model -> openrouter (default provider)
```

## Testing Strategy

1. **Unit tests**: Config loading, router logic, provider resolution
2. **Integration tests**: Mock HTTP servers for each provider type
3. **E2E tests**: Real API calls with test credentials (optional, CI secrets)

## Acceptance Criteria

- [ ] Configuration file is loaded from `.tuna.toml` or `~/.config/tuna/config.toml`
- [ ] Multiple providers can be defined with their own credentials
- [ ] Models are automatically routed to the correct provider
- [ ] Default provider is used for unknown models
- [ ] Backward compatibility with environment variables
- [ ] `tuna config show/validate/resolve` commands work correctly
- [ ] Rate limiting works per provider (configurable via `rate_limit`)
- [ ] Model aliases resolve to full model names before provider lookup
- [ ] Existing tests pass, new tests cover multi-provider scenarios

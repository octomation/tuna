---
issue: null
status: open
type: task
labels:
  - "effort: low"
  - "impact: low"
  - "scope: code"
  - "type: improvement"
assignees:
  - kamilsk
milestone: null
projects: []
relationships:
  parent: null
  blocked_by: []
  blocks: []
---

# Support Direct API Token in Configuration

## Summary

Add support for specifying API tokens directly in the configuration file, with a fallback to environment variable references for backward compatibility.

## Motivation

Currently, the only way to provide an API token for a provider is via `api_token_env`, which references an environment variable:

```toml
[[providers]]
name = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_token_env = "ANTHROPIC_API_KEY"  # Must be set in environment
```

This approach has limitations:
- Requires managing environment variables separately from the config
- Makes it harder to use different tokens for different projects
- Adds friction for quick local testing or one-off usage

## Proposed Solution

Add a new `api_token` field to the provider configuration that allows specifying the token directly:

```toml
[[providers]]
name = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_token = "sk-ant-..."  # Direct token value
```

### Token Resolution Priority

When resolving the API token for a provider:

1. **`api_token`** — use the value directly if present
2. **`api_token_env`** — read from the specified environment variable
3. **Error** — if neither is provided or the env variable is empty

This allows flexibility: use direct tokens for local development and env variables for production/CI.

### Configuration Examples

**Direct token (simple local setup):**
```toml
[[providers]]
name = "openrouter"
base_url = "https://openrouter.ai/api/v1"
api_token = "sk-or-v1-..."
models = ["anthropic/claude-sonnet-4"]
```

**Environment variable (production/CI):**
```toml
[[providers]]
name = "openrouter"
base_url = "https://openrouter.ai/api/v1"
api_token_env = "OPENROUTER_API_KEY"
models = ["anthropic/claude-sonnet-4"]
```

**Both specified (direct takes precedence):**
```toml
[[providers]]
name = "openrouter"
base_url = "https://openrouter.ai/api/v1"
api_token = "sk-or-v1-..."      # Used when present
api_token_env = "OPENROUTER_API_KEY"  # Fallback if api_token is empty
models = ["anthropic/claude-sonnet-4"]
```

## Implementation Plan

| File                              | Action | Description                                      |
|-----------------------------------|--------|--------------------------------------------------|
| `internal/config/config.go`       | Modify | Add `APIToken` field to `Provider` struct        |
| `internal/config/config.go`       | Modify | Update `Validate()` to accept either field       |
| `internal/config/config.go`       | Add    | Add `ResolveAPIToken()` method to `Provider`     |
| `internal/config/config_test.go`  | Modify | Add tests for new token resolution logic         |
| `internal/llm/router.go`          | Modify | Use `ResolveAPIToken()` instead of direct env    |
| `CLAUDE.md`                       | Modify | Document the new `api_token` field               |

### Code Changes

**Provider struct:**
```go
type Provider struct {
    Name        string   `toml:"name"`
    BaseURL     string   `toml:"base_url"`
    APIToken    string   `toml:"api_token"`      // Direct token value (new)
    APITokenEnv string   `toml:"api_token_env"`  // Env variable reference
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
```

**Validation update:**
```go
// In Validate(), change:
// Before:
if p.APITokenEnv == "" {
    errs = append(errs, fmt.Errorf("provider[%d] %q: api_token_env is required", i, p.Name))
}

// After:
if p.APIToken == "" && p.APITokenEnv == "" {
    errs = append(errs, fmt.Errorf("provider[%d] %q: either api_token or api_token_env is required", i, p.Name))
}
```

## Security Considerations

- Direct tokens in config files should be used carefully
- The config file (`.tuna.toml`) should be in `.gitignore` if it contains secrets
- For shared projects, prefer `api_token_env` to avoid accidental token commits
- Consider adding a warning when `tuna config show` displays a direct token

## Acceptance Criteria

- [ ] `api_token` field is recognized in provider configuration
- [ ] Direct token takes precedence over environment variable
- [ ] Validation passes when either `api_token` or `api_token_env` is provided
- [ ] Validation fails when neither is provided
- [ ] `tuna config show` works correctly with both token types
- [ ] `tuna config validate` validates the new field properly
- [ ] Documentation is updated with examples
- [ ] Existing tests pass, new tests cover token resolution

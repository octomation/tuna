# Tuna - Prompt Engineering Automation Tool

## Project Overview

Tuna is a CLI utility written in Go that automates the routine of testing and comparing LLM prompts across multiple models. It's designed for teams working on goal-setting assistants who need to iterate on system prompts efficiently.

## Core Concepts

### Assistant

A directory structure for a specific goal-setting assistant:
```
AssistantID/
├── Input/          # User query files (.txt or .md)
│   ├── query_001.txt
│   ├── query_002.txt
│   └── ...
├── Output/         # Generated responses (created by tuna, .txt or .md)
│   └── {plan_id}/
│       ├── {model_hash}/
│       │   ├── query_001_response.md
│       │   └── ...
│       └── plan.toml
└── System prompt   # Each file handles a specific aspect of the Assistant' system prompt (.txt or .md)
    ├── fragment_001.txt
    ├── fragment_002.txt
    └── ...
```

### Plan

A TOML configuration file that defines an execution plan:
- Plan ID (UUID)
- Assistant ID (folder name)
- Compiled system prompt
- List of input files
- Target models list
- Execution parameters (temperature, max_tokens, etc.)

## CLI Interface

```bash
# Initialize project structure
tuna init <AssistantID>

# Create an execution plan
tuna plan <AssistantID> [flags]
  --models, -m      Comma-separated list of models (default: claude-sonnet-4-20250514)
  --temperature     Temperature setting (default: 0.7)
  --max-tokens      Max tokens for response (default: 4096)

# Execute a plan
tuna exec <PlanID> [flags]
  --parallel, -p    Number of parallel requests (default: 1)
  --dry-run         Show what would be executed without making API calls
  --continue        Continue from last checkpoint if interrupted
```

## Configuration

Tuna supports multiple LLM providers via a TOML configuration file. Configuration is loaded from (in order of priority):

1. `.tuna.toml` in current directory or parent directories
2. `~/.config/tuna.toml`
3. Environment variables (deprecated): `LLM_API_TOKEN`, `LLM_BASE_URL`

### Configuration File Format

```toml
# Default provider used when model is not found in any provider's model list
default_provider = "openrouter"

# Model aliases for convenience (short name -> full model name)
[aliases]
sonnet = "claude-sonnet-4-20250514"
haiku = "claude-haiku-3-5-20241022"
gpt4 = "gpt-4o"

[[providers]]
name = "openrouter"
base_url = "https://openrouter.ai/api/v1"
api_token_env = "OPENROUTER_API_KEY"  # Reference to env variable
rate_limit = "10rpm"                   # 10 requests per minute
models = [
    "anthropic/claude-sonnet-4",
    "openai/gpt-4o",
    "meta-llama/llama-3.3-70b-instruct",
]

[[providers]]
name = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_token_env = "ANTHROPIC_API_KEY"
rate_limit = "60rpm"
models = [
    "claude-sonnet-4-20250514",
    "claude-haiku-3-5-20241022",
]

[[providers]]
name = "openai"
base_url = "https://api.openai.com/v1"
api_token_env = "OPENAI_API_KEY"
models = ["gpt-4o", "gpt-4o-mini"]
```

### Rate Limiting

Rate limits are specified in the format `<value><unit>`:
- `rps` - requests per second (e.g., `5rps`)
- `rpm` - requests per minute (e.g., `60rpm`)
- `rph` - requests per hour (e.g., `100rph`)

If no rate limit is specified, requests are unlimited.

### Configuration Commands

```bash
# Show current configuration
tuna config show

# Validate configuration file
tuna config validate

# Show which provider will be used for a model
tuna config resolve <model>
```

### Example Usage with Aliases

```bash
# Plan with aliases - much shorter than full model names
tuna plan MyAssistant --models "sonnet,gpt4"

# Mix aliases with full names
tuna plan MyAssistant --models "sonnet,gpt-4o-mini"

# Check alias resolution
tuna config resolve sonnet
# Output: sonnet -> claude-sonnet-4-20250514 -> anthropic
```

## Project Structure

```
tuna/
├── main.go              # Application entry point
├── go.mod               # Go module dependencies
├── go.sum               # Go module checksums
├── Makefile             # Build automation
├── Taskfile             # Task runner configuration
├── .tuna.toml           # Project-level configuration (optional)
├── internal/            # Internal packages (not exported)
│   ├── command/         # CLI command implementations
│   │   ├── root.go      # Root command and CLI setup
│   │   ├── init.go      # Init command
│   │   ├── plan.go      # Plan command
│   │   ├── exec.go      # Exec command
│   │   └── config.go    # Config command (show, validate, resolve)
│   ├── config/          # Configuration management
│   │   ├── config.go    # Config structures and validation
│   │   ├── loader.go    # TOML loading with priority resolution
│   │   └── features.go  # Feature flags and settings
│   ├── llm/             # LLM client implementations
│   │   ├── interface.go # ChatClient interface
│   │   ├── client.go    # Single provider client
│   │   └── router.go    # Multi-provider router
│   ├── exec/            # Plan execution
│   │   ├── executor.go  # Execution logic
│   │   ├── writer.go    # Response file writer
│   │   └── hash.go      # Model hash generation
│   ├── plan/            # Plan management
│   │   ├── plan.go      # Plan structures
│   │   └── loader.go    # Plan loading
│   └── assistant/       # Assistant management
│       ├── init.go      # Assistant initialization
│       ├── files.go     # File operations
│       └── prompt.go    # System prompt compilation
└── tools/               # Go tool dependencies
    ├── go.mod
    └── tools.go         # Tool imports for go generate
```

## Technical Requirements

### Dependencies

- **TUI Framework**: Charm libraries
  - `github.com/charmbracelet/bubbletea` - TUI framework
  - `github.com/charmbracelet/bubbles` - TUI components
  - `github.com/charmbracelet/lipgloss` - Styling
  - `github.com/charmbracelet/glamour` - Markdown rendering
- **CLI**: `github.com/spf13/cobra` - Command structure
- **Config**: `github.com/spf13/viper` - Configuration management
- **Logging**: `github.com/charmbracelet/log` - Styled logging
- **HTTP**: Standard library + `golang.org/x/time/rate` for rate limiting

## Error Handling

- All API errors should be retried with exponential backoff
- Failed requests should be logged with full context
- Execution should be resumable from checkpoints
- User should see clear, actionable error messages in TUI

## Testing Strategy

- Unit tests for concatenation logic
- Unit tests for plan generation
- Integration tests with mock LLM server
- E2E tests for CLI commands

## Notes for Development

- Start with a minimal working version (MVP): `plan` + `exec` with simple output
- Add TUI progressively after core logic works
- Use interfaces for LLM clients to enable easy testing and model additions
- Keep the codebase idiomatic Go (no over-engineering)
- Document public APIs with Go doc comments

## Language Guidelines

- **Code and files**: Always write in English (code, comments, documentation, commit messages, file names)
- **Communication**: Respond in the same language the user writes in

## Markdown Formatting

### Tables

Always align Markdown tables for readability in both raw and rendered views:

- Align all `|` characters vertically
- Pad cells with spaces so columns have consistent width
- Use the longest cell content as reference for column width

**Good example:**
```markdown
| File                             | Action |
|----------------------------------|--------|
| `go.mod`                         | Modify |
| `internal/plan/loader.go`        | Create |
| `internal/exec/executor_test.go` | Create |
```

**Bad example:**
```markdown
| File | Action |
|---|---|
| `go.mod` | Modify |
| `internal/plan/loader.go` | Create |
| `internal/exec/executor_test.go`| Create |
```

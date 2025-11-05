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

Environment variables for LLM integration:
- `LLM_API_TOKEN` - API token for authentication (required for `exec`)
- `LLM_BASE_URL` - Base URL for OpenAI-compatible API (required for `exec`)

## Project Structure

```
tuna/
├── main.go              # Application entry point
├── go.mod               # Go module dependencies
├── go.sum               # Go module checksums
├── Makefile             # Build automation
├── Taskfile             # Task runner configuration
├── internal/            # Internal packages (not exported)
│   ├── command/         # CLI command implementations
│   │   ├── demo/        # Demo commands (stdout, stderr, panic)
│   │   ├── root.go      # Root command and CLI setup
│   │   └── root_test.go # Root command tests
│   └── config/          # Configuration management
│       └── features.go  # Feature flags and settings
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

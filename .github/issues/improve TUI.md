---
issue: null
status: open
type: task
labels:
  - "effort: medium"
  - "impact: high"
  - "scope: code"
  - "scope: deps"
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

# Implement interactive TUI using Charm libraries

## Summary

Replace plain text output in all CLI commands with interactive TUI components built on Charm libraries. This will provide consistent visual feedback, progress tracking, and improved UX.

## Current state

All commands (`init`, `plan`, `exec`) use basic `cmd.Printf` for output. No Charm libraries are currently used.

## Requirements

### Dependencies

Add Charm libraries to `go.mod`:

| Package                                | Purpose                              |
|----------------------------------------|--------------------------------------|
| `github.com/charmbracelet/bubbletea`   | TUI framework                        |
| `github.com/charmbracelet/bubbles`     | Reusable TUI components              |
| `github.com/charmbracelet/lipgloss`    | Terminal styling                     |
| `github.com/charmbracelet/log`         | Styled logging (replace fatih/color) |

### Commands

#### `tuna init`

- Animated spinner while creating directories
- Styled list of created/skipped items (color-coded)
- Success/info message with lipgloss styling

#### `tuna plan`

- Spinner during plan generation
- Styled summary box with plan details (ID, models, queries count)
- Warning styling for edge cases (no queries found)

#### `tuna exec`

- Real-time progress bar showing overall completion
- Per-model/per-query status table
- Live token counter
- Elapsed time display
- Error highlighting
- Final summary with styled statistics

### Architecture

Create `internal/tui/` package:

```
internal/tui/
├── styles.go      # Shared lipgloss styles
├── spinner.go     # Reusable spinner component
├── progress.go    # Progress bar component
├── table.go       # Status table component
└── exec/
    └── model.go   # bubbletea model for exec command
```

### Fallback

Preserve non-interactive mode for:
- Piped output (`!isatty`)
- CI environments (`CI=true`)
- Explicit flag (`--no-tui`)

## Acceptance criteria

- [ ] Charm libraries added to dependencies
- [ ] `internal/tui/` package with shared components
- [ ] `init` command uses styled output
- [ ] `plan` command uses styled output
- [ ] `exec` command shows real-time progress with bubbletea
- [ ] Non-interactive fallback works correctly
- [ ] Existing tests pass

---
issue: 51
status: open
type: task
labels:
  - "effort: easy"
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

# Add Stub Commands for CLI Interface

## Summary

Implement stub commands (`init`, `plan`, `exec`) that match the CLI interface defined in the specification. These commands should be functional skeletons that accept all specified flags and arguments but only print placeholder messages.

## Background

Currently, the project only has demo commands (`panic`, `stderr`, `stdout`). We need to scaffold the main CLI commands before implementing the actual logic.

## Requirements

### Commands to implement

#### `tuna init <AssistantID>`
- **Purpose**: Initialize project structure for a new assistant
- **Arguments**: `AssistantID` (required) - name of the assistant directory to create
- **Stub behavior**: Print message like `"Initializing assistant: {AssistantID}..."`

#### `tuna plan <AssistantID> [flags]`
- **Purpose**: Create an execution plan
- **Arguments**: `AssistantID` (required)
- **Flags**:
  | Flag            | Short | Type   | Default                    | Description                   |
  |-----------------|-------|--------|----------------------------|-------------------------------|
  | `--models`      | `-m`  | string | `claude-sonnet-4-20250514` | Comma-separated list of models |
  | `--temperature` |       | float  | `0.7`                      | Temperature setting           |
  | `--max-tokens`  |       | int    | `4096`                     | Max tokens for response       |
- **Stub behavior**: Print received arguments and flag values

#### `tuna exec <PlanID> [flags]`
- **Purpose**: Execute a plan
- **Arguments**: `PlanID` (required)
- **Flags**:
  | Flag         | Short | Type | Default | Description                                          |
  |--------------|-------|------|---------|------------------------------------------------------|
  | `--parallel` | `-p`  | int  | `1`     | Number of parallel requests                          |
  | `--dry-run`  |       | bool | `false` | Show what would be executed without making API calls |
  | `--continue` |       | bool | `false` | Continue from last checkpoint if interrupted         |
- **Stub behavior**: Print received arguments and flag values

## Implementation Notes

1. Create new files under `internal/command/`:
   - `init.go`
   - `plan.go`
   - `exec.go`

2. Register commands in `root.go`

3. Follow existing code style from `demo/` commands

4. Update root command metadata:
   - `Use`: `tuna`
   - `Short`: `Prompt engineering automation tool`
   - `Long`: Descriptive help text

## Acceptance Criteria

- [ ] All three commands are registered and visible in `tuna --help`
- [ ] Each command accepts its specified arguments and flags
- [ ] Running a command prints its received configuration
- [ ] `tuna <cmd> --help` shows proper usage with all flags documented
- [ ] Code follows existing project conventions

---
issue: null
status: open
type: task
labels:
  - "effort: easy"
  - "impact: low"
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

# Implement `tuna init` command

## Description

Implement the `init` command that creates a directory structure for a new Assistant.

## Usage

```bash
tuna init <AssistantID>
```

## Expected behavior

The command should create the following directory structure:

```
<AssistantID>/
├── Input/
│   └── example_query.md
├── Output/
│   └── .gitkeep
└── System prompt/
    └── fragment_001.md
```

### File contents

**Input/example_query.md:**
```markdown
# Example Query

Write your user query here.
```

**System prompt/fragment_001.md:**
```markdown
# Fragment 001

Write your system prompt fragment here.
```

**Output/.gitkeep:**
Empty file to preserve the directory in git.

### AssistantID validation

AssistantID must be a valid directory name for the filesystem:
- Must not be empty
- Must not contain: `/ \ : * ? " < > |`
- Must not be `.` or `..`
- Must not exceed 255 characters

### Behavior with existing structure

If the directory already partially exists, the command should **complete the missing parts**:
- Create missing subdirectories (Input/, Output/, System prompt/)
- Create missing template files only if directory is empty
- Skip existing files (do not overwrite)
- Report what was created and what was skipped

## Acceptance criteria

- [ ] Command validates that AssistantID is provided
- [ ] Command validates AssistantID against filesystem naming rules
- [ ] Command creates all required directories
- [ ] Command creates template files with placeholder content
- [ ] Command completes partial structure if directory exists
- [ ] Command skips existing files without overwriting
- [ ] Command outputs summary: created/skipped items

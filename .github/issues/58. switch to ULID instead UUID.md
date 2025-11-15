---
issue: null
status: todo
type: task
labels:
  - "effort: easy"
  - "impact: medium"
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

# Switch to ULID instead of UUID for Plan IDs

## Summary

Replace UUID v4 with ULID (Universally Unique Lexicographically Sortable Identifier) for generating plan IDs. ULID preserves uniqueness while being lexicographically sortable by timestamp, which enables chronological ordering of plans.

## Motivation

Currently, plan IDs are generated using UUID v4, which produces random identifiers with no inherent ordering. This makes it difficult to identify recent plans or sort them chronologically.

ULID solves this by encoding a timestamp in the first 10 characters while maintaining uniqueness guarantees.

**Key benefits:**

1. **Chronological sorting** — ULID encodes timestamp in the first 10 characters, allowing natural `ls` ordering from oldest to newest
2. **Future-proof** — enables interactive plan picker in `tuna exec` without arguments (sorted newest to oldest)
3. **Same guarantees** — 128-bit, URL-safe, case-insensitive, maintains uniqueness

**ULID format:** `01ARZ3NDEKTSV4RRFFQ69G5FAV` (26 characters, Crockford Base32)

**UUID format:** `d9c35d53-288b-4bd4-ae44-572336ef7713` (36 characters with hyphens)

## Scope

### Files to Modify

| File                         | Action | Description                                                      |
|------------------------------|--------|------------------------------------------------------------------|
| `go.mod`                     | Modify | Replace `github.com/google/uuid` with `github.com/oklog/ulid/v2` |
| `internal/plan/plan.go`      | Modify | Change `uuid.New().String()` to ULID generation                  |
| `internal/plan/plan_test.go` | Modify | Update ID format validation (26 chars instead of 36)             |
| `CLAUDE.md`                  | Modify | Update documentation: Plan ID is ULID, not UUID                  |

### Code Changes

#### 1. Update `go.mod`

```bash
go get github.com/oklog/ulid/v2
go mod tidy  # This will remove unused github.com/google/uuid
```

#### 2. Update `internal/plan/plan.go`

Replace:
```go
import (
    "github.com/google/uuid"
)

// ...
planID := uuid.New().String()
```

With:
```go
import (
    "crypto/rand"
    "time"

    "github.com/oklog/ulid/v2"
)

// ...
planID := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
```

#### 3. Update `internal/plan/plan_test.go`

Replace:
```go
// Verify UUID format (should be 36 characters: 8-4-4-4-12)
if len(result.PlanID) != 36 {
    t.Errorf("Invalid UUID format: %s", result.PlanID)
}
```

With:
```go
// Verify ULID format (should be 26 characters, Crockford Base32)
if len(result.PlanID) != 26 {
    t.Errorf("Invalid ULID format: %s", result.PlanID)
}
```

#### 4. Update `CLAUDE.md`

Replace all mentions of "UUID" with "ULID" in the Plan section:

- Line 33: `- Plan ID (UUID)` → `- Plan ID (ULID)`

## Acceptance Criteria

- [ ] `github.com/oklog/ulid/v2` added to dependencies
- [ ] `github.com/google/uuid` removed from dependencies
- [ ] `tuna plan` generates ULID-based plan IDs (26 characters)
- [ ] Plans are sortable chronologically by folder name (`ls Output/` shows oldest first)
- [ ] All existing tests pass with updated assertions
- [ ] Documentation updated to reflect ULID usage

## Testing

```bash
# Run unit tests
go test ./internal/plan/...

# Manual verification
tuna plan test-assistant -m "gpt-4"
# Check that plan_id in Output/<plan_id>/plan.toml is 26 characters

# Verify sorting (create multiple plans with delay)
tuna plan test-assistant -m "gpt-4"
sleep 1
tuna plan test-assistant -m "gpt-4"
ls test-assistant/Output/
# Should show plans in chronological order
```

## References

- [ULID Spec](https://github.com/ulid/spec)
- [oklog/ulid Go library](https://github.com/oklog/ulid)
- [Video: Why ULID?](https://youtu.be/otW7nLd8P04)

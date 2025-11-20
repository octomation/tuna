---
issue: null
status: open
type: feature
labels:
  - "effort: hard"
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

# Implement `tuna view` Command

## Context

After executing a plan with multiple models, users need to review and compare LLM responses. Currently, this requires manually navigating directories and opening individual files. We need an interactive terminal UI that allows browsing responses grouped by input query, switching between models, and rating responses as good or bad.

The command should integrate with existing Charm-based TUI components (`internal/tui/`) and store ratings in metadata files alongside responses.

## Specification

### Usage

```bash
tuna view <PlanID>
```

### Behavior

1. Find `plan.toml` by PlanID using glob: `*/Output/<PlanID>/plan.toml`
2. Load all responses from `<AssistantID>/Output/<PlanID>/<model_hash>/`
3. Group responses by input query
4. Render interactive TUI for browsing and rating

### Navigation

| Key           | Action                                           |
|---------------|--------------------------------------------------|
| `↑` / `k`     | Previous input query                             |
| `↓` / `j`     | Next input query                                 |
| `←` / `h`     | Move focus to previous column (model response)   |
| `→` / `l`     | Move focus to next column (model response)       |
| `Space`       | Toggle good/bad rating for focused response      |
| `g`           | Mark focused response as good                    |
| `b`           | Mark focused response as bad                     |
| `u`           | Clear rating (unrate) for focused response       |
| `Enter`       | View focused response in full-screen pager       |
| `q` / `Esc`   | Quit viewer                                      |
| `?`           | Show help                                        |

### Interface Layout

All model responses are displayed side-by-side in columns for easy comparison. The focused column is highlighted with a distinct border. If columns don't fit on screen, horizontal scrolling is available.

**2-column layout (fits on screen):**

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Plan: 01HXYZ...  │  Query: 3/10  │  Models: 2                                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Input: query_003.md                                                          │
│ What are the key differences between REST and GraphQL?                       │
├──────────────────────────────────┬───────────────────────────────────────────┤
│ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ │ ┌──────────────────────────────────────┐  │
│ ┃ claude-sonnet      ✓ Good    ┃ │ │ gpt-4o                               │  │
│ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫ │ ├──────────────────────────────────────┤  │
│ ┃                              ┃ │ │                                      │  │
│ ┃ # REST vs GraphQL            ┃ │ │ # Comparing REST and GraphQL         │  │
│ ┃                              ┃ │ │                                      │  │
│ ┃ ## Key Differences           ┃ │ │ These are two popular API            │  │
│ ┃                              ┃ │ │ paradigms with distinct...           │  │
│ ┃ 1. **Data Fetching**         ┃ │ │                                      │  │
│ ┃    - REST: Multiple...       ┃ │ │ **Key Differences:**                 │  │
│ ┃    - GraphQL: Single...      ┃ │ │ 1. Data fetching approach            │  │
│ ┃                              ┃ │ │ 2. Type system                       │  │
│ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ │ └──────────────────────────────────────┘  │
├──────────────────────────────────┴───────────────────────────────────────────┤
│ ← → focus   ↑ ↓ query   Space: toggle   g: good   b: bad   q: quit   ?: help │
└──────────────────────────────────────────────────────────────────────────────┘
```

**3+ columns with horizontal scroll:**

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Plan: 01HXYZ...  │  Query: 3/10  │  Models: 4  │  Showing: 1-3 of 4  ◀━━▶    │
├──────────────────────────────────────────────────────────────────────────────┤
│ Input: query_003.md                                                          │
│ What are the key differences between REST and GraphQL?                       │
├──────────────────────┬──────────────────────┬────────────────────────────────┤
│ ┏━━━━━━━━━━━━━━━━━━┓ │ ┌──────────────────┐ │ ┌────────────────────────────┐ │
│ ┃ claude-sonnet    ┃ │ │ gpt-4o           │ │ │ llama-3.3-70b              │ │
│ ┃ ✓ Good  [1/4]    ┃ │ │        [2/4]     │ │ │ ✗ Bad  [3/4]               │ │
│ ┣━━━━━━━━━━━━━━━━━━┫ │ ├──────────────────┤ │ ├────────────────────────────┤ │
│ ┃                  ┃ │ │                  │ │ │                            │ │
│ ┃ # REST vs GQL    ┃ │ │ # REST & GraphQL │ │ │ REST and GraphQL are       │ │
│ ┃                  ┃ │ │                  │ │ │ both API technologies...   │ │
│ ┃ ## Key Diffs     ┃ │ │ Two approaches   │ │ │                            │ │
│ ┃ ...              ┃ │ │ ...              │ │ │ ...                        │ │
│ ┗━━━━━━━━━━━━━━━━━━┛ │ └──────────────────┘ │ └────────────────────────────┘ │
├──────────────────────┴──────────────────────┴────────────────────────────────┤
│ ← → focus/scroll   ↑ ↓ query   Space: toggle   g: good   b: bad   ?: help    │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Visual Indicators

- **Focused column**: Bold/double border (`┏━━┓`) with accent color
- **Unfocused columns**: Normal border (`┌──┐`)
- `✓` Green checkmark for "good" responses
- `✗` Red cross for "bad" responses
- No indicator for unrated responses
- `[N/M]` Position indicator showing current model position
- `◀━━▶` Horizontal scroll indicator when more columns exist off-screen

### Metadata Storage

Store ratings in YAML front matter within the response file itself:

```
Output/{plan_id}/{model_hash}/
├── query_001_response.md   # Contains front matter with rating
├── query_002_response.md
└── ...
```

**Response file format:**

```markdown
---
rating: good
rated_at: 2024-01-15T10:30:00Z
---

# Response content here...

The actual LLM response follows the front matter.
```

**Important:**
- Front matter is **not rendered** when displaying the response
- Use `glamour` to render markdown content (excluding front matter)
- When saving a rating, update or add the front matter block
- If response has no front matter yet, prepend it on first rating

## Implementation Steps

### Phase 1: Core Viewer

#### 1. Create response loader

**File:** `internal/view/loader.go`

```go
package view

import (
    "os"
    "path/filepath"
    "time"

    "github.com/pelletier/go-toml/v2"
    "go.octolab.org/toolset/tuna/internal/exec"
    "go.octolab.org/toolset/tuna/internal/plan"
)

type ResponseGroup struct {
    QueryID   string
    InputPath string
    InputText string
    Responses []ModelResponse
}

type ModelResponse struct {
    Model     string
    ModelHash string
    FilePath  string
    Content   string
    Rating    Rating
    RatedAt   time.Time
}

type Rating string

const (
    RatingNone Rating = ""
    RatingGood Rating = "good"
    RatingBad  Rating = "bad"
)

func LoadResponses(planPath string) ([]ResponseGroup, error) {
    p, err := plan.LoadFromPath(planPath)
    if err != nil {
        return nil, err
    }

    assistantDir := plan.AssistantDir(planPath)
    outputDir := filepath.Dir(planPath)

    var groups []ResponseGroup
    for _, query := range p.Queries {
        group := ResponseGroup{
            QueryID:   query.ID,
            InputPath: filepath.Join(assistantDir, "Input", query.ID),
        }

        // Read input content
        content, err := os.ReadFile(group.InputPath)
        if err != nil {
            return nil, err
        }
        group.InputText = string(content)

        // Load responses for each model
        for _, model := range p.Assistant.LLM.Models {
            hash := exec.ModelHash(model)
            respPath := filepath.Join(outputDir, hash, responseFileName(query.ID))

            resp := ModelResponse{
                Model:     model,
                ModelHash: hash,
                FilePath:  respPath,
            }

            // Parse response: extracts metadata from front matter,
            // returns content without front matter for rendering
            if meta, content, err := ParseResponse(respPath); err == nil {
                resp.Content = content  // Already stripped of front matter
                resp.Rating = meta.Rating
                resp.RatedAt = meta.RatedAt
            }

            group.Responses = append(group.Responses, resp)
        }

        groups = append(groups, group)
    }

    return groups, nil
}

func responseFileName(queryID string) string {
    base := strings.TrimSuffix(queryID, filepath.Ext(queryID))
    return base + "_response.md"
}
```

#### 2. Create metadata handler (YAML front matter)

**File:** `internal/view/metadata.go`

```go
package view

import (
    "os"
    "regexp"
    "strings"
    "time"

    "gopkg.in/yaml.v3"
)

type Metadata struct {
    Rating  Rating    `yaml:"rating,omitempty"`
    RatedAt time.Time `yaml:"rated_at,omitempty"`
}

// Front matter regex: matches ---\n...\n---\n at the start of file
var frontMatterRegex = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n`)

// ParseResponse splits a response file into metadata and content.
// Content is returned without front matter for rendering.
func ParseResponse(filePath string) (*Metadata, string, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, "", err
    }

    content := string(data)
    meta := &Metadata{}

    // Try to extract front matter
    if matches := frontMatterRegex.FindStringSubmatch(content); len(matches) == 2 {
        if err := yaml.Unmarshal([]byte(matches[1]), meta); err != nil {
            // Invalid YAML, treat as no metadata
            return &Metadata{}, content, nil
        }
        // Remove front matter from content for rendering
        content = frontMatterRegex.ReplaceAllString(content, "")
    }

    return meta, strings.TrimLeft(content, "\n"), nil
}

// SaveRating updates or adds front matter with the rating.
func SaveRating(filePath string, rating Rating) error {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }

    content := string(data)
    meta := Metadata{
        Rating:  rating,
        RatedAt: time.Now(),
    }

    // Check if front matter exists
    if matches := frontMatterRegex.FindStringSubmatch(content); len(matches) == 2 {
        // Parse existing front matter and update
        var existing Metadata
        yaml.Unmarshal([]byte(matches[1]), &existing)
        existing.Rating = rating
        existing.RatedAt = time.Now()
        meta = existing

        // Remove old front matter
        content = frontMatterRegex.ReplaceAllString(content, "")
    }

    // Build new front matter
    yamlData, err := yaml.Marshal(meta)
    if err != nil {
        return err
    }

    newContent := "---\n" + string(yamlData) + "---\n\n" + strings.TrimLeft(content, "\n")

    return os.WriteFile(filePath, []byte(newContent), 0644)
}

// StripFrontMatter removes front matter from content for display.
func StripFrontMatter(content string) string {
    return strings.TrimLeft(frontMatterRegex.ReplaceAllString(content, ""), "\n")
}
```

#### 3. Create TUI model

**File:** `internal/tui/view/model.go`

```go
package view

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/lipgloss"

    "go.octolab.org/toolset/tuna/internal/tui"
    viewpkg "go.octolab.org/toolset/tuna/internal/view"
)

// Column styles
var (
    focusedBorder = lipgloss.NewStyle().
        Border(lipgloss.DoubleBorder()).
        BorderForeground(tui.ColorCyan)

    unfocusedBorder = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(tui.ColorGray)
)

type Model struct {
    planID       string
    groups       []viewpkg.ResponseGroup
    queryIndex   int
    focusIndex   int      // Currently focused column
    scrollOffset int      // Horizontal scroll offset (first visible column)
    viewports    []viewport.Model
    width        int
    height       int
    columnWidth  int
    visibleCols  int      // Number of columns that fit on screen
    showHelp     bool
    renderer     *glamour.TermRenderer
}

func New(planID string, groups []viewpkg.ResponseGroup) Model {
    renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())

    return Model{
        planID:      planID,
        groups:      groups,
        columnWidth: 40, // Default, recalculated on resize
        renderer:    renderer,
    }
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "esc":
            return m, tea.Quit
        case "up", "k":
            if m.queryIndex > 0 {
                m.queryIndex--
                m.focusIndex = 0
                m.scrollOffset = 0
                m.updateViewports()
            }
        case "down", "j":
            if m.queryIndex < len(m.groups)-1 {
                m.queryIndex++
                m.focusIndex = 0
                m.scrollOffset = 0
                m.updateViewports()
            }
        case "left", "h":
            if m.focusIndex > 0 {
                m.focusIndex--
                // Scroll left if focus goes off-screen
                if m.focusIndex < m.scrollOffset {
                    m.scrollOffset = m.focusIndex
                }
            }
        case "right", "l":
            responses := m.groups[m.queryIndex].Responses
            if m.focusIndex < len(responses)-1 {
                m.focusIndex++
                // Scroll right if focus goes off-screen
                if m.focusIndex >= m.scrollOffset+m.visibleCols {
                    m.scrollOffset = m.focusIndex - m.visibleCols + 1
                }
            }
        case " ":
            m.toggleRating()
        case "g":
            m.setRating(viewpkg.RatingGood)
        case "b":
            m.setRating(viewpkg.RatingBad)
        case "u":
            m.setRating(viewpkg.RatingNone)
        case "?":
            m.showHelp = !m.showHelp
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.calculateLayout()
        m.updateViewports()
    }

    // Update focused viewport for scrolling
    if m.focusIndex < len(m.viewports) {
        var cmd tea.Cmd
        m.viewports[m.focusIndex], cmd = m.viewports[m.focusIndex].Update(msg)
        return m, cmd
    }

    return m, nil
}

func (m *Model) calculateLayout() {
    // Calculate column width and visible columns based on terminal width
    minColWidth := 30
    maxColWidth := 60
    padding := 4 // Border + spacing

    // Try to fit as many columns as possible
    m.columnWidth = (m.width - padding) / 2
    if m.columnWidth < minColWidth {
        m.columnWidth = minColWidth
        m.visibleCols = 1
    } else if m.columnWidth > maxColWidth {
        m.columnWidth = maxColWidth
        m.visibleCols = (m.width - padding) / (m.columnWidth + 2)
    } else {
        m.visibleCols = 2
    }

    // Cap visible columns by actual model count
    if len(m.groups) > 0 {
        modelCount := len(m.groups[m.queryIndex].Responses)
        if m.visibleCols > modelCount {
            m.visibleCols = modelCount
        }
    }
}

func (m *Model) updateViewports() {
    if len(m.groups) == 0 {
        return
    }

    responses := m.groups[m.queryIndex].Responses
    m.viewports = make([]viewport.Model, len(responses))

    vpHeight := m.height - 12 // Header + input + footer

    for i, resp := range responses {
        vp := viewport.New(m.columnWidth-4, vpHeight)
        rendered, _ := m.renderer.Render(resp.Content)
        vp.SetContent(rendered)
        m.viewports[i] = vp
    }
}

func (m *Model) toggleRating() {
    resp := &m.groups[m.queryIndex].Responses[m.focusIndex]
    switch resp.Rating {
    case viewpkg.RatingNone:
        m.setRating(viewpkg.RatingGood)
    case viewpkg.RatingGood:
        m.setRating(viewpkg.RatingBad)
    case viewpkg.RatingBad:
        m.setRating(viewpkg.RatingNone)
    }
}

func (m *Model) setRating(rating viewpkg.Rating) {
    resp := &m.groups[m.queryIndex].Responses[m.focusIndex]
    resp.Rating = rating
    // Save rating to YAML front matter in the response file
    viewpkg.SaveRating(resp.FilePath, rating)
}

func (m Model) View() string {
    if m.showHelp {
        return m.viewHelp()
    }

    var sb strings.Builder

    sb.WriteString(m.viewHeader())
    sb.WriteString("\n")
    sb.WriteString(m.viewInput())
    sb.WriteString("\n")
    sb.WriteString(m.viewColumns())
    sb.WriteString("\n")
    sb.WriteString(m.viewFooter())

    return sb.String()
}

func (m Model) viewHeader() string {
    group := m.groups[m.queryIndex]
    modelCount := len(group.Responses)

    planPart := tui.Muted.Render(fmt.Sprintf("Plan: %s", truncate(m.planID, 12)))
    queryPart := fmt.Sprintf("Query: %d/%d", m.queryIndex+1, len(m.groups))
    modelsPart := fmt.Sprintf("Models: %d", modelCount)

    // Show scroll indicator if needed
    scrollPart := ""
    if modelCount > m.visibleCols {
        scrollPart = fmt.Sprintf("Showing: %d-%d of %d",
            m.scrollOffset+1,
            min(m.scrollOffset+m.visibleCols, modelCount),
            modelCount)
        if m.scrollOffset > 0 {
            scrollPart = "◀ " + scrollPart
        }
        if m.scrollOffset+m.visibleCols < modelCount {
            scrollPart = scrollPart + " ▶"
        }
    }

    return fmt.Sprintf("%s  │  %s  │  %s  │  %s", planPart, queryPart, modelsPart, scrollPart)
}

func (m Model) viewInput() string {
    group := m.groups[m.queryIndex]
    header := tui.Bold.Render(fmt.Sprintf("Input: %s", group.QueryID))
    content := tui.Muted.Render(truncate(group.InputText, m.width-4))
    return fmt.Sprintf("%s\n%s", header, content)
}

func (m Model) viewColumns() string {
    if len(m.groups) == 0 {
        return ""
    }

    group := m.groups[m.queryIndex]
    responses := group.Responses

    // Render visible columns
    var columns []string
    endIdx := min(m.scrollOffset+m.visibleCols, len(responses))

    for i := m.scrollOffset; i < endIdx; i++ {
        resp := responses[i]
        isFocused := (i == m.focusIndex)
        col := m.renderColumn(resp, i, len(responses), isFocused)
        columns = append(columns, col)
    }

    // Join columns horizontally
    return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (m Model) renderColumn(resp viewpkg.ModelResponse, idx, total int, focused bool) string {
    // Header: model name + rating + position
    ratingStr := ""
    switch resp.Rating {
    case viewpkg.RatingGood:
        ratingStr = tui.Success.Render("✓ Good")
    case viewpkg.RatingBad:
        ratingStr = tui.Error.Render("✗ Bad")
    }

    header := fmt.Sprintf("%s\n%s [%d/%d]",
        truncate(resp.Model, m.columnWidth-10),
        ratingStr,
        idx+1, total)

    // Content from viewport
    content := ""
    if idx < len(m.viewports) {
        content = m.viewports[idx].View()
    }

    fullContent := header + "\n\n" + content

    // Apply border style based on focus
    var style lipgloss.Style
    if focused {
        style = focusedBorder.Width(m.columnWidth)
    } else {
        style = unfocusedBorder.Width(m.columnWidth)
    }

    return style.Render(fullContent)
}

func (m Model) viewFooter() string {
    return tui.Muted.Render("← → focus   ↑ ↓ query   Space: toggle   g: good   b: bad   q: quit   ?: help")
}

func (m Model) viewHelp() string {
    return `
Keyboard Shortcuts
──────────────────

Navigation:
  ↑ / k       Previous query
  ↓ / j       Next query
  ← / h       Focus previous column (scrolls if needed)
  → / l       Focus next column (scrolls if needed)

Rating (applies to focused column):
  Space       Toggle rating (none → good → bad → none)
  g           Mark as good
  b           Mark as bad
  u           Clear rating

Other:
  Enter       View full response in pager
  ?           Toggle this help
  q / Esc     Quit

Press any key to close help...
`
}

func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    return s[:max-3] + "..."
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

#### 4. Create view command

**File:** `internal/command/view.go`

```go
package command

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"

    "go.octolab.org/toolset/tuna/internal/plan"
    "go.octolab.org/toolset/tuna/internal/view"
    viewtui "go.octolab.org/toolset/tuna/internal/tui/view"
)

func View() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "view <PlanID>",
        Short: "View and rate LLM responses",
        Long: `View opens an interactive terminal UI for browsing LLM responses.

Navigation:
  ↑/↓         Switch between input queries
  ←/→         Switch between model responses
  Space/g/b   Rate responses as good or bad
  q           Quit`,
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            planID := args[0]

            cwd, err := os.Getwd()
            if err != nil {
                return fmt.Errorf("failed to get working directory: %w", err)
            }

            _, planPath, err := plan.Load(cwd, planID)
            if err != nil {
                return err
            }

            groups, err := view.LoadResponses(planPath)
            if err != nil {
                return fmt.Errorf("failed to load responses: %w", err)
            }

            if len(groups) == 0 {
                return fmt.Errorf("no responses found for plan %s", planID)
            }

            model := viewtui.New(planID, groups)
            p := tea.NewProgram(model, tea.WithAltScreen())

            if _, err := p.Run(); err != nil {
                return fmt.Errorf("viewer error: %w", err)
            }

            return nil
        },
    }

    return cmd
}
```

#### 5. Register command

**File:** `internal/command/root.go`

Add `View()` to `AddCommand()` calls.

### Phase 2: Polish

#### 6. Add scrollable viewport

Update `internal/tui/view/model.go`:
- Handle `PageUp`, `PageDown` keys
- Show scroll indicators when content overflows

#### 7. Add pager mode

- `Enter` opens full-screen response view
- `Esc` returns to normal view

#### 8. Add unit tests

**File:** `internal/view/loader_test.go`
**File:** `internal/view/metadata_test.go`
**File:** `internal/tui/view/model_test.go`

## File Changes

| File                           | Action |
|--------------------------------|--------|
| `go.mod`                       | Modify |
| `internal/view/loader.go`      | Create |
| `internal/view/metadata.go`    | Create |
| `internal/view/loader_test.go` | Create |
| `internal/view/metadata_test.go` | Create |
| `internal/tui/view/model.go`   | Create |
| `internal/command/view.go`     | Create |
| `internal/command/root.go`     | Modify |

**New dependency:** `gopkg.in/yaml.v3` for YAML front matter parsing.

## Acceptance Criteria

- [ ] `tuna view <PlanID>` opens interactive viewer
- [ ] Multiple model responses displayed side-by-side in columns
- [ ] Focused column highlighted with distinct border style
- [ ] Navigate between queries with ↑/↓
- [ ] Move focus between columns with ←/→
- [ ] Horizontal scroll when columns exceed screen width
- [ ] Scroll indicator shows visible range (e.g., "Showing: 1-2 of 4")
- [ ] Markdown responses render properly with glamour
- [ ] Each column independently scrollable (vertical)
- [ ] Space/g/b keys update ratings for focused column
- [ ] Ratings persist to YAML front matter in response files
- [ ] Front matter not rendered (stripped before display)
- [ ] Ratings visible on re-opening viewer
- [ ] Clean exit with q/Esc
- [ ] Help available with ?

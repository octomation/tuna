// Package view provides the TUI model for viewing and rating LLM responses.
package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"go.octolab.org/toolset/tuna/internal/tui"
	"go.octolab.org/toolset/tuna/internal/view"
)

// Column styles
var (
	focusedBorder = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(tui.ColorCyan)

	unfocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(tui.ColorGray)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(tui.ColorCyan)

	goodRatingStyle = lipgloss.NewStyle().
			Foreground(tui.ColorGreen)

	badRatingStyle = lipgloss.NewStyle().
			Foreground(tui.ColorRed)
)

// Model is the bubbletea model for the response viewer.
type Model struct {
	planID        string
	groups        []view.ResponseGroup
	queryIndex    int
	focusIndex    int // Currently focused column
	scrollOffset  int // Horizontal scroll offset (first visible column)
	viewports     []viewport.Model
	width         int
	height        int
	columnWidth   int
	visibleCols   int  // Number of columns that fit on screen
	showHelp      bool
	inputExpanded bool // Whether input query section is expanded
	mdRenderer    *glamour.TermRenderer

	// Cache for rendered markdown content (key: "queryIdx:respIdx:width")
	renderCache     map[string]string
	lastColumnWidth int // Track width changes for cache invalidation
}

// New creates a new view TUI model.
func New(planID string, groups []view.ResponseGroup) Model {
	// Create markdown renderer - use DarkStyle for faster init (no terminal detection)
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(0), // We'll handle wrapping ourselves
	)

	return Model{
		planID:      planID,
		groups:      groups,
		columnWidth: 40, // Default, recalculated on resize
		mdRenderer:  renderer,
		renderCache: make(map[string]string),
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showHelp {
			// Any key closes help
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit

		case "k": // Only k for previous query (not up arrow)
			if m.queryIndex > 0 {
				m.queryIndex--
				m.focusIndex = 0
				m.scrollOffset = 0
				m.updateViewports()
			}

		case "j": // Only j for next query (not down arrow)
			if m.queryIndex < len(m.groups)-1 {
				m.queryIndex++
				m.focusIndex = 0
				m.scrollOffset = 0
				m.updateViewports()
			}

		case "up": // Scroll content up in focused column
			if m.focusIndex < len(m.viewports) {
				m.viewports[m.focusIndex].LineUp(3)
			}

		case "down": // Scroll content down in focused column
			if m.focusIndex < len(m.viewports) {
				m.viewports[m.focusIndex].LineDown(3)
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
			if len(m.groups) > 0 {
				responses := m.groups[m.queryIndex].Responses
				if m.focusIndex < len(responses)-1 {
					m.focusIndex++
					// Scroll right if focus goes off-screen
					if m.focusIndex >= m.scrollOffset+m.visibleCols {
						m.scrollOffset = m.focusIndex - m.visibleCols + 1
					}
				}
			}

		case " ":
			m.toggleRating()

		case "g":
			m.setRating(view.RatingGood)

		case "b":
			m.setRating(view.RatingBad)

		case "u":
			m.setRating(view.RatingNone)

		case "?":
			m.showHelp = !m.showHelp

		case "tab":
			m.inputExpanded = !m.inputExpanded
			m.updateViewports() // Recalculate column heights

		case "pgup":
			if m.focusIndex < len(m.viewports) {
				m.viewports[m.focusIndex].HalfViewUp()
			}

		case "pgdown":
			if m.focusIndex < len(m.viewports) {
				m.viewports[m.focusIndex].HalfViewDown()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.calculateLayout()
		m.updateViewports()

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			// Scroll content up in focused column
			if m.focusIndex < len(m.viewports) {
				m.viewports[m.focusIndex].LineUp(3)
			}
		case tea.MouseButtonWheelDown:
			// Scroll content down in focused column
			if m.focusIndex < len(m.viewports) {
				m.viewports[m.focusIndex].LineDown(3)
			}
		case tea.MouseButtonLeft:
			// Only handle press, not release (to avoid double-toggle)
			if msg.Action != tea.MouseActionPress {
				break
			}

			// Check if click is in the input area (header is ~2 lines, input section follows)
			inputAreaStart := 2 // After header
			inputAreaEnd := inputAreaStart + m.inputHeight()

			if msg.Y >= inputAreaStart && msg.Y < inputAreaEnd {
				// Click on input area - toggle expand/collapse
				m.inputExpanded = !m.inputExpanded
				m.updateViewports()
			} else if msg.Y >= inputAreaEnd {
				// Click on column area - focus the column
				if len(m.groups) > 0 && m.queryIndex < len(m.groups) {
					clickedCol := m.getColumnAtX(msg.X)
					if clickedCol >= 0 {
						m.focusIndex = clickedCol
					}
				}
			}
		}
	}

	// Update focused viewport for scrolling within column
	if len(m.viewports) > 0 && m.focusIndex < len(m.viewports) {
		var cmd tea.Cmd
		m.viewports[m.focusIndex], cmd = m.viewports[m.focusIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

// getColumnAtX returns the column index at the given X coordinate, or -1 if none.
func (m Model) getColumnAtX(x int) int {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return -1
	}

	responses := m.groups[m.queryIndex].Responses
	if len(responses) == 0 {
		return -1
	}

	// Each column has width = m.columnWidth + 1 (for gap between columns)
	colWidthWithGap := m.columnWidth + 1

	// Calculate which visible column was clicked
	visibleColIndex := x / colWidthWithGap

	// Convert to actual column index (accounting for scroll offset)
	actualColIndex := m.scrollOffset + visibleColIndex

	// Validate the index
	if actualColIndex < 0 || actualColIndex >= len(responses) {
		return -1
	}

	// Also check if we're within visible columns
	if visibleColIndex >= m.visibleCols {
		return -1
	}

	return actualColIndex
}

func (m *Model) calculateLayout() {
	// Layout rules:
	// - Maximum 2 columns visible at once
	// - Columns fill all available horizontal space
	// - If more than 2 models, horizontal scrolling is enabled
	const maxVisibleCols = 2

	// Get model count for current query
	modelCount := 0
	if len(m.groups) > 0 && m.queryIndex < len(m.groups) {
		modelCount = len(m.groups[m.queryIndex].Responses)
	}

	// Determine number of visible columns: min(modelCount, maxVisibleCols)
	m.visibleCols = modelCount
	if m.visibleCols > maxVisibleCols {
		m.visibleCols = maxVisibleCols
	}
	if m.visibleCols < 1 {
		m.visibleCols = 1
	}

	// Calculate column width to fill available space
	// Account for borders (2 chars per column) and gaps between columns
	borderWidth := 2 * m.visibleCols
	gapWidth := m.visibleCols - 1 // 1 char gap between columns
	availableWidth := m.width - borderWidth - gapWidth

	m.columnWidth = availableWidth / m.visibleCols
	if m.columnWidth < 20 {
		m.columnWidth = 20 // Absolute minimum for readability
	}
}

func (m *Model) updateViewports() {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return
	}

	responses := m.groups[m.queryIndex].Responses
	m.viewports = make([]viewport.Model, len(responses))

	// Calculate viewport height: total height - header(2) - input section - column header(2) - footer(1) - borders(2)
	inputH := m.inputHeight()
	vpHeight := m.height - inputH - 7
	if vpHeight < 5 {
		vpHeight = 5
	}

	// Calculate content width inside viewport (minus borders)
	contentWidth := m.columnWidth - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Invalidate cache if column width changed
	if m.lastColumnWidth != contentWidth {
		m.renderCache = make(map[string]string)
		m.lastColumnWidth = contentWidth

		// Recreate renderer with proper word wrap width
		m.mdRenderer, _ = glamour.NewTermRenderer(
			glamour.WithStylePath("dark"),
			glamour.WithWordWrap(contentWidth),
		)
	}

	for i, resp := range responses {
		vp := viewport.New(contentWidth, vpHeight)

		// Check cache first
		cacheKey := fmt.Sprintf("%d:%d:%d", m.queryIndex, i, contentWidth)
		content, cached := m.renderCache[cacheKey]

		if !cached {
			// Render markdown content
			if m.mdRenderer != nil && resp.Content != "" {
				rendered, err := m.mdRenderer.Render(resp.Content)
				if err == nil {
					content = strings.TrimSpace(rendered)
				} else {
					// Fallback to plain text
					content = wrapText(resp.Content, contentWidth)
				}
			} else {
				content = wrapText(resp.Content, contentWidth)
			}
			// Store in cache
			m.renderCache[cacheKey] = content
		}

		vp.SetContent(content)
		m.viewports[i] = vp
	}
}

func (m *Model) toggleRating() {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return
	}
	responses := m.groups[m.queryIndex].Responses
	if m.focusIndex >= len(responses) {
		return
	}

	resp := &m.groups[m.queryIndex].Responses[m.focusIndex]
	switch resp.Rating {
	case view.RatingNone:
		m.setRating(view.RatingGood)
	case view.RatingGood:
		m.setRating(view.RatingBad)
	case view.RatingBad:
		m.setRating(view.RatingNone)
	}
}

func (m *Model) setRating(rating view.Rating) {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return
	}
	responses := m.groups[m.queryIndex].Responses
	if m.focusIndex >= len(responses) {
		return
	}

	resp := &m.groups[m.queryIndex].Responses[m.focusIndex]
	resp.Rating = rating
	// Save rating to YAML front matter in the response file
	view.SaveRating(resp.FilePath, rating)
}

// View renders the model.
func (m Model) View() string {
	if m.showHelp {
		return m.viewHelp()
	}

	if len(m.groups) == 0 {
		return "No responses to display.\n\nPress 'q' to quit."
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
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return ""
	}

	group := m.groups[m.queryIndex]
	modelCount := len(group.Responses)

	planPart := tui.Muted.Render(fmt.Sprintf("Plan: %s", truncate(m.planID, 12)))
	queryPart := fmt.Sprintf("Query: %d/%d", m.queryIndex+1, len(m.groups))
	modelsPart := fmt.Sprintf("Models: %d", modelCount)

	// Show scroll indicator if needed
	scrollPart := ""
	if modelCount > m.visibleCols {
		endIdx := m.scrollOffset + m.visibleCols
		if endIdx > modelCount {
			endIdx = modelCount
		}
		scrollPart = fmt.Sprintf("Showing: %d-%d of %d",
			m.scrollOffset+1,
			endIdx,
			modelCount)
		if m.scrollOffset > 0 {
			scrollPart = "<< " + scrollPart
		}
		if m.scrollOffset+m.visibleCols < modelCount {
			scrollPart = scrollPart + " >>"
		}
	}

	parts := []string{planPart, queryPart, modelsPart}
	if scrollPart != "" {
		parts = append(parts, scrollPart)
	}

	return headerStyle.Render(strings.Join(parts, "  |  "))
}

// inputHeight returns the number of lines used by the input section.
func (m Model) inputHeight() int {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return 2 // header + empty line
	}

	if m.inputExpanded {
		// Count actual lines in input, but cap at 30% of screen height
		lines := strings.Count(m.groups[m.queryIndex].InputText, "\n") + 1
		maxLines := m.height * 30 / 100
		if maxLines < 3 {
			maxLines = 3
		}
		if lines > maxLines {
			lines = maxLines
		}
		return lines + 2 // +2 for header and border/spacing
	}

	return 4 // header + 2 lines of content + hint
}

func (m Model) viewInput() string {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return ""
	}

	// Handle case when width is not yet initialized
	width := m.width
	if width < 20 {
		width = 80 // Default fallback
	}

	group := m.groups[m.queryIndex]

	// Build header with expand/collapse indicator
	expandIndicator := "[Tab to expand]"
	if m.inputExpanded {
		expandIndicator = "[Tab to collapse]"
	}
	header := fmt.Sprintf("%s  %s",
		tui.Bold.Render(fmt.Sprintf("Input: %s", group.QueryID)),
		tui.Muted.Render(expandIndicator))

	// Safe line truncation helper
	truncateLine := func(line string, maxLen int) string {
		if maxLen < 10 {
			maxLen = 10
		}
		if len(line) <= maxLen {
			return line
		}
		return line[:maxLen-3] + "..."
	}

	// Show content based on expanded state
	var content string
	if m.inputExpanded {
		// Show full content (up to 30% of screen height)
		maxLines := m.height * 30 / 100
		if maxLines < 3 {
			maxLines = 3
		}
		lines := strings.Split(group.InputText, "\n")
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, tui.Muted.Render("... (truncated)"))
		}
		// Wrap long lines
		var wrappedLines []string
		for _, line := range lines {
			wrappedLines = append(wrappedLines, truncateLine(line, width-6))
		}
		content = strings.Join(wrappedLines, "\n")
	} else {
		// Show first 2 lines collapsed
		lines := strings.Split(group.InputText, "\n")
		previewLines := 2
		if len(lines) < previewLines {
			previewLines = len(lines)
		}
		var preview []string
		for i := 0; i < previewLines; i++ {
			preview = append(preview, truncateLine(lines[i], width-6))
		}
		if len(lines) > 2 {
			preview = append(preview, tui.Muted.Render(fmt.Sprintf("... (+%d more lines)", len(lines)-2)))
		}
		content = strings.Join(preview, "\n")
	}

	// Add a border around input
	boxWidth := width - 4
	if boxWidth < 10 {
		boxWidth = 10
	}
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.ColorGray).
		Width(boxWidth).
		Padding(0, 1)

	return header + "\n" + inputStyle.Render(content)
}

func (m Model) viewColumns() string {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return ""
	}

	group := m.groups[m.queryIndex]
	responses := group.Responses

	if len(responses) == 0 {
		return tui.Muted.Render("No model responses found.")
	}

	// Show loading state if viewports not yet initialized
	if len(m.viewports) == 0 {
		return tui.Muted.Render("Loading responses...")
	}

	// Render visible columns
	var columns []string
	endIdx := m.scrollOffset + m.visibleCols
	if endIdx > len(responses) {
		endIdx = len(responses)
	}

	for i := m.scrollOffset; i < endIdx; i++ {
		resp := responses[i]
		isFocused := (i == m.focusIndex)
		col := m.renderColumn(resp, i, len(responses), isFocused)
		columns = append(columns, col)
	}

	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (m Model) renderColumn(resp view.ModelResponse, idx, total int, focused bool) string {
	// Header: model name + rating + position
	modelName := truncate(resp.Model, m.columnWidth-20)

	ratingStr := ""
	switch resp.Rating {
	case view.RatingGood:
		ratingStr = goodRatingStyle.Render(" [Good]")
	case view.RatingBad:
		ratingStr = badRatingStyle.Render(" [Bad]")
	}

	posStr := tui.Muted.Render(fmt.Sprintf(" [%d/%d]", idx+1, total))

	header := fmt.Sprintf("%s%s%s", modelName, ratingStr, posStr)

	// Content from viewport
	content := ""
	if idx < len(m.viewports) {
		content = m.viewports[idx].View()
	} else if resp.Content != "" {
		// Fallback if viewport not ready
		content = truncate(resp.Content, m.columnWidth*3)
	} else {
		content = tui.Muted.Render("(no response)")
	}

	// Separator line
	separatorWidth := m.columnWidth - 2
	if separatorWidth < 5 {
		separatorWidth = 5
	}
	separator := strings.Repeat("─", separatorWidth)

	fullContent := header + "\n" + separator + "\n" + content

	// Column height: total height - header(2) - input section - footer(1) - border(2)
	inputH := m.inputHeight()
	colHeight := m.height - inputH - 5
	if colHeight < 5 {
		colHeight = 5
	}

	// Apply border style based on focus
	var style lipgloss.Style
	if focused {
		style = focusedBorder.Width(m.columnWidth).Height(colHeight)
	} else {
		style = unfocusedBorder.Width(m.columnWidth).Height(colHeight)
	}

	return style.Render(fullContent)
}

func (m Model) viewFooter() string {
	return tui.Muted.Render("h/l: focus  j/k: query  ↑↓/scroll: content  Tab: input  g/b: rate  q: quit  ?: help")
}

func (m Model) viewHelp() string {
	help := `
Keyboard Shortcuts
------------------

Query Navigation:
  k            Previous query
  j            Next query

Column Navigation:
  h / ←        Focus previous column
  l / →        Focus next column
  Click        Focus clicked column

Content Scrolling:
  ↑ / ↓        Scroll content in focused column
  Mouse wheel  Scroll content in focused column
  PgUp/PgDn    Scroll half page

Input:
  Tab          Expand/collapse input query section
  Click        Expand/collapse input query section

Rating (applies to focused column):
  Space        Toggle rating (none → good → bad → none)
  g            Mark as good
  b            Mark as bad
  u            Clear rating

Other:
  ?            Toggle this help
  q / Esc      Quit

Press any key to close help...
`
	return headerStyle.Render("Help") + help
}

func truncate(s string, max int) string {
	if max < 4 {
		max = 4
	}
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// wrapText wraps text to fit within a given width.
func wrapText(text string, width int) string {
	if width < 10 {
		width = 10
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Handle empty lines
		if len(line) == 0 {
			continue
		}

		// Simple word wrapping
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			}
		}
		result.WriteString(currentLine)
	}

	return result.String()
}

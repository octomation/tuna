// Package view provides the TUI model for viewing and rating LLM responses.
package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	planID       string
	groups       []view.ResponseGroup
	queryIndex   int
	focusIndex   int                // Currently focused column
	scrollOffset int                // Horizontal scroll offset (first visible column)
	viewports    []viewport.Model
	width        int
	height       int
	columnWidth  int
	visibleCols  int // Number of columns that fit on screen
	showHelp     bool
}

// New creates a new view TUI model.
func New(planID string, groups []view.ResponseGroup) Model {
	return Model{
		planID:      planID,
		groups:      groups,
		columnWidth: 40, // Default, recalculated on resize
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
	}

	// Update focused viewport for scrolling within column
	if len(m.viewports) > 0 && m.focusIndex < len(m.viewports) {
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
	availableWidth := m.width - padding
	m.columnWidth = availableWidth / 2
	if m.columnWidth < minColWidth {
		m.columnWidth = minColWidth
		m.visibleCols = 1
	} else if m.columnWidth > maxColWidth {
		m.columnWidth = maxColWidth
		m.visibleCols = availableWidth / (m.columnWidth + 2)
	} else {
		m.visibleCols = 2
	}

	// Cap visible columns by actual model count
	if len(m.groups) > 0 && m.queryIndex < len(m.groups) {
		modelCount := len(m.groups[m.queryIndex].Responses)
		if m.visibleCols > modelCount {
			m.visibleCols = modelCount
		}
	}

	// Ensure at least 1 visible column
	if m.visibleCols < 1 {
		m.visibleCols = 1
	}
}

func (m *Model) updateViewports() {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return
	}

	responses := m.groups[m.queryIndex].Responses
	m.viewports = make([]viewport.Model, len(responses))

	// Calculate viewport height: total height - header - input - footer
	vpHeight := m.height - 10
	if vpHeight < 5 {
		vpHeight = 5
	}

	// Calculate content width inside viewport (minus borders and padding)
	contentWidth := m.columnWidth - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	for i, resp := range responses {
		vp := viewport.New(contentWidth, vpHeight)
		// Wrap content to fit the column width
		wrapped := wrapText(resp.Content, contentWidth)
		vp.SetContent(wrapped)
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

func (m Model) viewInput() string {
	if len(m.groups) == 0 || m.queryIndex >= len(m.groups) {
		return ""
	}

	group := m.groups[m.queryIndex]
	header := tui.Bold.Render(fmt.Sprintf("Input: %s", group.QueryID))

	// Show first line or truncated input
	inputPreview := strings.Split(group.InputText, "\n")[0]
	inputPreview = truncate(inputPreview, m.width-10)
	content := tui.Muted.Render(inputPreview)

	return fmt.Sprintf("%s\n%s", header, content)
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
	modelName := truncate(resp.Model, m.columnWidth-15)

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

	fullContent := header + "\n" + strings.Repeat("-", m.columnWidth-4) + "\n" + content

	// Apply border style based on focus
	var style lipgloss.Style
	if focused {
		style = focusedBorder.Width(m.columnWidth).Height(m.height - 8)
	} else {
		style = unfocusedBorder.Width(m.columnWidth).Height(m.height - 8)
	}

	return style.Render(fullContent)
}

func (m Model) viewFooter() string {
	return tui.Muted.Render("< > focus   ^ v query   Space: toggle   g: good   b: bad   u: unrate   q: quit   ?: help")
}

func (m Model) viewHelp() string {
	help := `
Keyboard Shortcuts
------------------

Navigation:
  Up / k       Previous query
  Down / j     Next query
  Left / h     Focus previous column (scrolls if needed)
  Right / l    Focus next column (scrolls if needed)
  PgUp/PgDn    Scroll content in focused column

Rating (applies to focused column):
  Space        Toggle rating (none -> good -> bad -> none)
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

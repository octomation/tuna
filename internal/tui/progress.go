package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

// Progress is a thread-safe progress bar wrapper.
type Progress struct {
	mu       sync.Mutex
	total    int
	current  int
	model    progress.Model
	message  string
	width    int
}

// NewProgress creates a new progress bar with the given total count.
func NewProgress(total int) *Progress {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return &Progress{
		total: total,
		model: p,
		width: 40,
	}
}

// SetMessage sets the current progress message.
func (p *Progress) SetMessage(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = msg
}

// Increment increases the progress by 1.
func (p *Progress) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current < p.total {
		p.current++
	}
}

// Current returns the current progress value.
func (p *Progress) Current() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

// Total returns the total count.
func (p *Progress) Total() int {
	return p.total
}

// Percent returns the current progress as a percentage (0.0 to 1.0).
func (p *Progress) Percent() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.total == 0 {
		return 0
	}
	return float64(p.current) / float64(p.total)
}

// View returns the progress bar as a string.
func (p *Progress) View() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	percent := float64(0)
	if p.total > 0 {
		percent = float64(p.current) / float64(p.total)
	}

	bar := p.model.ViewAs(percent)
	stats := Muted.Render(fmt.Sprintf(" %d/%d", p.current, p.total))

	if p.message != "" {
		return fmt.Sprintf("%s%s %s", bar, stats, p.message)
	}
	return bar + stats
}

// SimpleProgress renders a simple text-based progress bar for non-interactive mode.
func SimpleProgress(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s] %d/%d (%.0f%%)", bar, current, total, percent*100)
}

// ProgressStyle returns a styled progress indicator for inline use.
func ProgressStyle(current, total int) string {
	percent := float64(0)
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}

	var style lipgloss.Style
	switch {
	case percent >= 100:
		style = Success
	case percent >= 50:
		style = Info
	default:
		style = Warning
	}

	return style.Render(fmt.Sprintf("%d/%d", current, total))
}

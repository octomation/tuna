package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorGreen  = lipgloss.Color("2")
	ColorYellow = lipgloss.Color("3")
	ColorRed    = lipgloss.Color("1")
	ColorBlue   = lipgloss.Color("4")
	ColorGray   = lipgloss.Color("8")
	ColorCyan   = lipgloss.Color("6")
)

// Text styles
var (
	// Success style for successful operations (green)
	Success = lipgloss.NewStyle().Foreground(ColorGreen)

	// Warning style for warnings (yellow)
	Warning = lipgloss.NewStyle().Foreground(ColorYellow)

	// Error style for errors (red)
	Error = lipgloss.NewStyle().Foreground(ColorRed)

	// Info style for informational messages (blue)
	Info = lipgloss.NewStyle().Foreground(ColorBlue)

	// Muted style for less important text (gray)
	Muted = lipgloss.NewStyle().Foreground(ColorGray)

	// Bold style for emphasis
	Bold = lipgloss.NewStyle().Bold(true)

	// Title style for headers
	Title = lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
)

// Symbols for list items
var (
	SymbolCreated = Success.Render("+")
	SymbolSkipped = Muted.Render("-")
	SymbolError   = Error.Render("✗")
	SymbolSuccess = Success.Render("✓")
	SymbolPending = Muted.Render("○")
	SymbolRunning = Info.Render("●")
)

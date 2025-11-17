package tui

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// spinnerModel is a bubbletea model for running a function with a spinner.
type spinnerModel struct {
	spinner spinner.Model
	message string
	err     error
	done    bool
	fn      func() error
}

type spinnerDoneMsg struct {
	err error
}

func newSpinnerModel(message string, fn func() error) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	return spinnerModel{
		spinner: s,
		message: message,
		fn:      fn,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			err := m.fn()
			return spinnerDoneMsg{err: err}
		},
	)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinnerDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

// RunWithSpinner executes fn while showing a spinner with the given message.
// In non-interactive mode, it just prints the message and runs fn.
func RunWithSpinner(message string, fn func() error) error {
	if !IsInteractive() {
		fmt.Println(message + "...")
		return fn()
	}

	model := newSpinnerModel(message, fn)
	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("spinner error: %w", err)
	}

	if m, ok := finalModel.(spinnerModel); ok && m.err != nil {
		return m.err
	}

	return nil
}

// RunWithSpinnerOutput executes fn while showing a spinner, capturing any output.
// Returns the error from fn (if any).
func RunWithSpinnerOutput(w io.Writer, message string, fn func() error) error {
	if !IsInteractive() {
		fmt.Fprintln(w, message+"...")
		return fn()
	}

	model := newSpinnerModel(message, fn)
	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("spinner error: %w", err)
	}

	if m, ok := finalModel.(spinnerModel); ok && m.err != nil {
		return m.err
	}

	return nil
}

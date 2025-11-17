// Package exec provides the TUI model for plan execution.
package exec

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"go.octolab.org/toolset/tuna/internal/tui"
)

// TaskStatus represents the status of a single task.
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskComplete
	TaskFailed
)

// Task represents a single execution task (model + query combination).
type Task struct {
	Model    string
	QueryID  string
	Status   TaskStatus
	Error    error
	Tokens   TokenUsage
	Duration time.Duration
}

// TokenUsage holds token counts.
type TokenUsage struct {
	Prompt int
	Output int
}

// Model is the bubbletea model for execution progress.
type Model struct {
	tasks       []Task
	current     int
	totalTokens TokenUsage
	startTime   time.Time
	spinner     spinner.Model
	progress    progress.Model
	done        bool
	width       int
	err         error
}

// New creates a new execution TUI model.
func New(models []string, queries []string) Model {
	// Create tasks for all model/query combinations
	var tasks []Task
	for _, model := range models {
		for _, query := range queries {
			tasks = append(tasks, Task{
				Model:   model,
				QueryID: query,
				Status:  TaskPending,
			})
		}
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tui.ColorCyan)

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return Model{
		tasks:     tasks,
		startTime: time.Now(),
		spinner:   s,
		progress:  p,
		width:     80,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Messages for updating the model from the executor.

// TaskStartMsg signals that a task has started.
type TaskStartMsg struct {
	Model   string
	QueryID string
}

// TaskDoneMsg signals that a task has completed.
type TaskDoneMsg struct {
	Model    string
	QueryID  string
	Tokens   TokenUsage
	Duration time.Duration
}

// TaskErrorMsg signals that a task has failed.
type TaskErrorMsg struct {
	Model   string
	QueryID string
	Err     error
}

// ExecutionDoneMsg signals that all tasks are complete.
type ExecutionDoneMsg struct {
	Err error
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.progress.Width = msg.Width - 10
		if m.progress.Width > 60 {
			m.progress.Width = 60
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case TaskStartMsg:
		for i := range m.tasks {
			if m.tasks[i].Model == msg.Model && m.tasks[i].QueryID == msg.QueryID {
				m.tasks[i].Status = TaskRunning
				m.current = i
				break
			}
		}

	case TaskDoneMsg:
		for i := range m.tasks {
			if m.tasks[i].Model == msg.Model && m.tasks[i].QueryID == msg.QueryID {
				m.tasks[i].Status = TaskComplete
				m.tasks[i].Tokens = msg.Tokens
				m.tasks[i].Duration = msg.Duration
				m.totalTokens.Prompt += msg.Tokens.Prompt
				m.totalTokens.Output += msg.Tokens.Output
				break
			}
		}

	case TaskErrorMsg:
		for i := range m.tasks {
			if m.tasks[i].Model == msg.Model && m.tasks[i].QueryID == msg.QueryID {
				m.tasks[i].Status = TaskFailed
				m.tasks[i].Error = msg.Err
				break
			}
		}

	case ExecutionDoneMsg:
		m.done = true
		m.err = msg.Err
		return m, tea.Quit
	}

	return m, nil
}

// View renders the model.
func (m Model) View() string {
	if m.done {
		return m.viewDone()
	}

	var sb strings.Builder

	// Title
	sb.WriteString(tui.Title.Render("Executing plan"))
	sb.WriteString("\n\n")

	// Progress bar
	completed := m.completedCount()
	percent := float64(completed) / float64(len(m.tasks))
	sb.WriteString(m.progress.ViewAs(percent))
	sb.WriteString(tui.Muted.Render(fmt.Sprintf(" %d/%d", completed, len(m.tasks))))
	sb.WriteString("\n\n")

	// Current task
	if m.current < len(m.tasks) && m.tasks[m.current].Status == TaskRunning {
		task := m.tasks[m.current]
		sb.WriteString(m.spinner.View())
		sb.WriteString(" ")
		sb.WriteString(tui.Info.Render(task.Model))
		sb.WriteString(" ")
		sb.WriteString(tui.Muted.Render("→"))
		sb.WriteString(" ")
		sb.WriteString(task.QueryID)
		sb.WriteString("\n")
	}

	// Stats
	sb.WriteString("\n")
	elapsed := time.Since(m.startTime).Round(time.Second)
	sb.WriteString(tui.Muted.Render(fmt.Sprintf("Elapsed: %s", elapsed)))
	sb.WriteString("  ")
	sb.WriteString(tui.Muted.Render(fmt.Sprintf("Tokens: %d prompt + %d output",
		m.totalTokens.Prompt, m.totalTokens.Output)))
	sb.WriteString("\n")

	// Recent completed tasks (show last 3)
	recentCompleted := m.recentCompleted(3)
	if len(recentCompleted) > 0 {
		sb.WriteString("\n")
		for _, task := range recentCompleted {
			sb.WriteString("  ")
			sb.WriteString(tui.SymbolSuccess)
			sb.WriteString(" ")
			sb.WriteString(tui.Muted.Render(task.Model))
			sb.WriteString(" → ")
			sb.WriteString(task.QueryID)
			sb.WriteString("\n")
		}
	}

	// Errors
	errors := m.failedTasks()
	if len(errors) > 0 {
		sb.WriteString("\n")
		sb.WriteString(tui.Error.Render("Errors:"))
		sb.WriteString("\n")
		for _, task := range errors {
			sb.WriteString("  ")
			sb.WriteString(tui.SymbolError)
			sb.WriteString(" ")
			sb.WriteString(task.Model)
			sb.WriteString(" → ")
			sb.WriteString(task.QueryID)
			sb.WriteString(": ")
			sb.WriteString(task.Error.Error())
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (m Model) viewDone() string {
	var sb strings.Builder

	completed := m.completedCount()
	failed := len(m.failedTasks())
	elapsed := time.Since(m.startTime).Round(time.Second)

	if failed == 0 {
		sb.WriteString(tui.RenderSuccess("Execution complete"))
	} else {
		sb.WriteString(tui.RenderWarning(fmt.Sprintf("Execution complete with %d errors", failed)))
	}
	sb.WriteString("\n\n")

	// Stats
	sb.WriteString(tui.RenderKeyValue("Tasks", fmt.Sprintf("%d/%d completed", completed, len(m.tasks))))
	sb.WriteString("\n")
	sb.WriteString(tui.RenderKeyValue("Tokens", fmt.Sprintf("%d prompt + %d output = %d total",
		m.totalTokens.Prompt, m.totalTokens.Output, m.totalTokens.Prompt+m.totalTokens.Output)))
	sb.WriteString("\n")
	sb.WriteString(tui.RenderKeyValue("Elapsed", elapsed.String()))
	sb.WriteString("\n")

	return sb.String()
}

func (m Model) completedCount() int {
	count := 0
	for _, task := range m.tasks {
		if task.Status == TaskComplete || task.Status == TaskFailed {
			count++
		}
	}
	return count
}

func (m Model) recentCompleted(n int) []Task {
	var completed []Task
	for i := len(m.tasks) - 1; i >= 0 && len(completed) < n; i-- {
		if m.tasks[i].Status == TaskComplete {
			completed = append([]Task{m.tasks[i]}, completed...)
		}
	}
	return completed
}

func (m Model) failedTasks() []Task {
	var failed []Task
	for _, task := range m.tasks {
		if task.Status == TaskFailed {
			failed = append(failed, task)
		}
	}
	return failed
}

// Tasks returns all tasks.
func (m Model) Tasks() []Task {
	return m.tasks
}

// TotalTokens returns the total token usage.
func (m Model) TotalTokens() TokenUsage {
	return m.totalTokens
}

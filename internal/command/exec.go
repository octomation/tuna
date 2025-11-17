package command

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/tuna/internal/config"
	"go.octolab.org/toolset/tuna/internal/exec"
	"go.octolab.org/toolset/tuna/internal/llm"
	"go.octolab.org/toolset/tuna/internal/plan"
	"go.octolab.org/toolset/tuna/internal/tui"
	tuiexec "go.octolab.org/toolset/tuna/internal/tui/exec"
)

// Exec returns a cobra.Command to execute a plan.
//
//	$ tuna exec <PlanID> [flags]
func Exec() *cobra.Command {
	var (
		parallel   int
		dryRun     bool
		continueOp bool
	)

	command := cobra.Command{
		Use:   "exec <PlanID>",
		Short: "Execute a plan",
		Long: `Execute runs the specified plan, sending queries to the configured models.

Configuration is loaded from (in order of priority):
  1. .tuna.toml in current directory or parent directories
  2. ~/.config/tuna.toml
  3. Environment variables (deprecated): LLM_API_TOKEN, LLM_BASE_URL

Use 'tuna config show' to see the current configuration.`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			planID := args[0]

			// Warn about unimplemented flags
			if parallel > 1 {
				cmd.PrintErrln("Warning: --parallel is not yet implemented, using default (1)")
			}
			if continueOp {
				cmd.PrintErrln("Warning: --continue is not yet implemented")
			}

			// Get working directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			// Load plan
			p, planPath, err := plan.Load(cwd, planID)
			if err != nil {
				return err
			}

			assistantDir := plan.AssistantDir(planPath)

			// Dry run mode
			if dryRun {
				executor := exec.New(p, assistantDir, nil, exec.Options{DryRun: true})
				cmd.Print(executor.DryRun())
				return nil
			}

			// Load configuration
			cfgResult, err := config.Load()
			if err != nil {
				return err
			}

			// Show deprecation warning if using environment variables
			if cfgResult.Deprecated {
				cmd.PrintErrln(config.DeprecationWarning())
			}

			// Create router
			router, err := llm.NewRouter(cfgResult.Config)
			if err != nil {
				return err
			}

			// Execute with TUI or non-interactive mode
			if tui.IsInteractive() {
				return executeWithTUI(cmd, p, assistantDir, router, planID, parallel, continueOp)
			}
			return executeNonInteractive(cmd, p, assistantDir, router, planID, parallel, continueOp)
		},
	}

	command.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel requests")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without making API calls")
	command.Flags().BoolVar(&continueOp, "continue", false, "Continue from last checkpoint if interrupted")

	return &command
}

func executeWithTUI(cmd *cobra.Command, p *plan.Plan, assistantDir string, router llm.ChatClient, planID string, parallel int, continueOp bool) error {
	// Create TUI model
	models := p.Assistant.LLM.Models
	queries := make([]string, len(p.Queries))
	for i, q := range p.Queries {
		queries[i] = q.ID
	}

	model := tuiexec.New(models, queries)
	program := tea.NewProgram(model, tea.WithAltScreen())

	// Create executor with progress callback
	executor := exec.New(p, assistantDir, router, exec.Options{
		Parallel: parallel,
		Continue: continueOp,
		OnProgress: func(event exec.ProgressEvent) {
			switch event.Type {
			case exec.EventTaskStart:
				program.Send(tuiexec.TaskStartMsg{
					Model:   event.Model,
					QueryID: event.QueryID,
				})
			case exec.EventTaskDone:
				program.Send(tuiexec.TaskDoneMsg{
					Model:   event.Model,
					QueryID: event.QueryID,
					Tokens: tuiexec.TokenUsage{
						Prompt: event.Tokens.Prompt,
						Output: event.Tokens.Output,
					},
					Duration: event.Duration,
				})
			case exec.EventTaskError:
				program.Send(tuiexec.TaskErrorMsg{
					Model:   event.Model,
					QueryID: event.QueryID,
					Err:     event.Err,
				})
			}
		},
	})

	// Run executor in background
	var summary *exec.ExecutionSummary
	var execErr error

	go func() {
		ctx := context.Background()
		summary, execErr = executor.Execute(ctx)
		program.Send(tuiexec.ExecutionDoneMsg{Err: execErr})
	}()

	// Run TUI
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Print final summary (already shown in TUI, but add results list)
	if summary != nil && len(summary.Results) > 0 {
		cmd.Println()
		cmd.Println(tui.Bold.Render("Output files:"))
		for _, result := range summary.Results {
			cmd.Printf("  %s %s\n", tui.SymbolSuccess, result.OutputPath)
		}
	}

	return execErr
}

func executeNonInteractive(cmd *cobra.Command, p *plan.Plan, assistantDir string, router llm.ChatClient, planID string, parallel int, continueOp bool) error {
	// Execute
	executor := exec.New(p, assistantDir, router, exec.Options{
		Parallel: parallel,
		Continue: continueOp,
		OnProgress: func(event exec.ProgressEvent) {
			// Simple progress output for non-interactive mode
			switch event.Type {
			case exec.EventTaskStart:
				cmd.Printf("  Processing %s with %s...\n", event.QueryID, event.Model)
			case exec.EventTaskDone:
				cmd.Printf("  ✓ %s -> %s (%d tokens)\n", event.QueryID, event.Model,
					event.Tokens.Prompt+event.Tokens.Output)
			case exec.EventTaskError:
				cmd.Printf("  ✗ %s -> %s: %v\n", event.QueryID, event.Model, event.Err)
			}
		},
	})

	ctx := context.Background()
	summary, err := executor.Execute(ctx)
	if err != nil {
		return err
	}

	// Print summary
	cmd.Printf("\nExecution complete\n\n")
	cmd.Printf("Plan:      %s\n", planID)
	cmd.Printf("Queries:   %d\n", summary.TotalQueries)
	cmd.Printf("Models:    %d\n", summary.TotalModels)
	cmd.Printf("Tokens:    %d prompt + %d output = %d total\n\n",
		summary.TotalTokens.Prompt,
		summary.TotalTokens.Output,
		summary.TotalTokens.Prompt+summary.TotalTokens.Output)

	cmd.Println("Results:")
	for _, result := range summary.Results {
		cmd.Printf("  + %s -> %s\n", result.QueryID, result.OutputPath)
	}

	if len(summary.Errors) > 0 {
		cmd.Println("\nErrors:")
		for _, err := range summary.Errors {
			cmd.Printf("  x %s\n", err)
		}
	}

	return nil
}

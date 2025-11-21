package command

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/tuna/internal/plan"
	"go.octolab.org/toolset/tuna/internal/tui"
	"go.octolab.org/toolset/tuna/internal/view"
	viewtui "go.octolab.org/toolset/tuna/internal/tui/view"
)

// View returns the view command.
func View() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <PlanID>",
		Short: "View and rate LLM responses",
		Long: `View opens an interactive terminal UI for browsing LLM responses.

After executing a plan with multiple models, use this command to review
and compare responses. You can navigate between queries and models,
and rate responses as good or bad.

Navigation:
  Up/Down      Switch between input queries
  Left/Right   Switch between model responses
  Space/g/b    Rate responses as good or bad
  u            Clear rating
  q            Quit`,
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

			// Non-interactive mode: print summary
			if !tui.IsInteractive() {
				return printViewSummary(planID, groups)
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

// printViewSummary prints a non-interactive summary of responses.
func printViewSummary(planID string, groups []view.ResponseGroup) error {
	fmt.Printf("Plan: %s\n", planID)
	fmt.Printf("Queries: %d\n", len(groups))

	if len(groups) > 0 {
		fmt.Printf("Models: %d\n\n", len(groups[0].Responses))
	}

	for i, group := range groups {
		fmt.Printf("Query %d/%d: %s\n", i+1, len(groups), group.QueryID)

		for _, resp := range group.Responses {
			ratingStr := "(unrated)"
			if resp.Rating == view.RatingGood {
				ratingStr = "[Good]"
			} else if resp.Rating == view.RatingBad {
				ratingStr = "[Bad]"
			}

			contentPreview := ""
			if len(resp.Content) > 50 {
				contentPreview = resp.Content[:50] + "..."
			} else if resp.Content != "" {
				contentPreview = resp.Content
			} else {
				contentPreview = "(no response)"
			}

			// Remove newlines from preview
			for _, r := range "\n\r" {
				for {
					idx := -1
					for i, c := range contentPreview {
						if c == r {
							idx = i
							break
						}
					}
					if idx == -1 {
						break
					}
					contentPreview = contentPreview[:idx] + " " + contentPreview[idx+1:]
				}
			}

			fmt.Printf("  - %s %s: %s\n", resp.Model, ratingStr, contentPreview)
		}
		fmt.Println()
	}

	fmt.Println("Run without --no-tui flag to view responses interactively.")
	return nil
}

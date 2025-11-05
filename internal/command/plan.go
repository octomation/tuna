package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.octolab.org/template/tool/internal/plan"
)

// Plan returns a cobra.Command to create an execution plan.
//
//	$ tuna plan <AssistantID> [flags]
func Plan() *cobra.Command {
	var (
		models      string
		temperature float64
		maxTokens   int
	)

	command := cobra.Command{
		Use:   "plan <AssistantID>",
		Short: "Create an execution plan",
		Long: `Plan creates a TOML configuration file that defines an execution session.

The plan includes:
  - Plan ID (UUID v4)
  - Compiled system prompt (from System prompt/ directory)
  - List of input queries (from Input/ directory)
  - Target models and execution parameters

Output: <AssistantID>/Output/<plan_id>/plan.toml`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assistantID := args[0]

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			cfg := plan.Config{
				Models:      plan.ParseModels(models),
				Temperature: temperature,
				MaxTokens:   maxTokens,
			}

			result, err := plan.Generate(cwd, assistantID, cfg)
			if err != nil {
				return err
			}

			// Print summary
			cmd.Printf("Plan created: %s\n", result.PlanPath)
			cmd.Printf("  Plan ID: %s\n", result.PlanID)
			cmd.Printf("  Models:  %d\n", result.ModelsCount)
			cmd.Printf("  Queries: %d\n", result.QueriesCount)

			if result.QueriesCount == 0 {
				cmd.Println("\nWarning: No input queries found. Add .txt or .md files to Input/ directory.")
			}

			return nil
		},
	}

	command.Flags().StringVarP(&models, "models", "m", "claude-sonnet-4-20250514", "Comma-separated list of models")
	command.Flags().Float64Var(&temperature, "temperature", 0.7, "Temperature setting")
	command.Flags().IntVar(&maxTokens, "max-tokens", 4096, "Max tokens for response")

	return &command
}

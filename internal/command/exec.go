package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/tuna/internal/config"
	"go.octolab.org/toolset/tuna/internal/exec"
	"go.octolab.org/toolset/tuna/internal/llm"
	"go.octolab.org/toolset/tuna/internal/plan"
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

			// Execute
			executor := exec.New(p, assistantDir, router, exec.Options{
				DryRun:   dryRun,
				Parallel: parallel,
				Continue: continueOp,
			})

			ctx := context.Background()
			summary, err := executor.Execute(ctx)
			if err != nil {
				return err
			}

			// Print summary
			cmd.Printf("Execution complete\n\n")
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
		},
	}

	command.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel requests")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without making API calls")
	command.Flags().BoolVar(&continueOp, "continue", false, "Continue from last checkpoint if interrupted")

	return &command
}

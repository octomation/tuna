package command

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.octolab.org/template/tool/internal/exec"
	"go.octolab.org/template/tool/internal/llm"
	"go.octolab.org/template/tool/internal/plan"
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

Environment variables required:
  LLM_API_TOKEN  API token for authentication
  LLM_BASE_URL   Base URL for OpenAI-compatible API

MVP: Executes only the first query with the first model.`,

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

			// Load LLM config from environment
			llmCfg, err := llm.ConfigFromEnv()
			if err != nil {
				return err
			}

			// Create LLM client
			llmClient := llm.NewClient(llmCfg)

			// Execute
			executor := exec.New(p, assistantDir, llmClient, exec.Options{
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

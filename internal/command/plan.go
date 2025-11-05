package command

import (
	"github.com/spf13/cobra"
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
  - Plan ID (UUID)
  - Assistant ID
  - Compiled system prompt
  - List of input files
  - Target models list
  - Execution parameters`,

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Creating plan for assistant: %s\n", args[0])
			cmd.Printf("  Models:      %s\n", models)
			cmd.Printf("  Temperature: %.1f\n", temperature)
			cmd.Printf("  Max tokens:  %d\n", maxTokens)
		},
	}

	command.Flags().StringVarP(&models, "models", "m", "claude-sonnet-4-20250514", "Comma-separated list of models")
	command.Flags().Float64Var(&temperature, "temperature", 0.7, "Temperature setting")
	command.Flags().IntVar(&maxTokens, "max-tokens", 4096, "Max tokens for response")

	return &command
}

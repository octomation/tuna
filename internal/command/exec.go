package command

import (
	"github.com/spf13/cobra"
)

// Exec returns a cobra.Command to execute a plan.
//
//	$ tuna exec <PlanID> [flags]
func Exec() *cobra.Command {
	var (
		parallel    int
		dryRun      bool
		continueExe bool
	)

	command := cobra.Command{
		Use:   "exec <PlanID>",
		Short: "Execute a plan",
		Long: `Execute runs the specified plan, sending queries to the configured models
and saving responses to the output directory.

The execution can be parallelized and supports resumption from checkpoints
if interrupted.`,

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Executing plan: %s\n", args[0])
			cmd.Printf("  Parallel: %d\n", parallel)
			cmd.Printf("  Dry run:  %t\n", dryRun)
			cmd.Printf("  Continue: %t\n", continueExe)
		},
	}

	command.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel requests")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be executed without making API calls")
	command.Flags().BoolVar(&continueExe, "continue", false, "Continue from last checkpoint if interrupted")

	return &command
}

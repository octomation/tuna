package command

import (
	"github.com/spf13/cobra"

	"go.octolab.org/toolset/tuna/internal/tui"
)

// New returns the new root command.
func New() *cobra.Command {
	var noTUI bool

	command := cobra.Command{
		Use:   "tuna",
		Short: "Prompt engineering automation tool",
		Long: `Tuna is a CLI utility that automates the routine of testing and comparing
LLM prompts across multiple models. It helps teams iterate on system prompts
efficiently by organizing inputs, outputs, and execution plans.`,

		SilenceErrors: false,
		SilenceUsage:  true,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if noTUI {
				tui.SetNonInteractive()
			}
		},
	}

	command.PersistentFlags().BoolVar(&noTUI, "no-tui", false, "Disable interactive TUI")

	/* configure instance */
	command.AddCommand(
		Init(),
		Plan(),
		Exec(),
		View(),
		Config(),
	)

	return &command
}

package command

import (
	"github.com/spf13/cobra"
)

// New returns the new root command.
func New() *cobra.Command {
	command := cobra.Command{
		Use:   "tuna",
		Short: "Prompt engineering automation tool",
		Long: `Tuna is a CLI utility that automates the routine of testing and comparing
LLM prompts across multiple models. It helps teams iterate on system prompts
efficiently by organizing inputs, outputs, and execution plans.`,

		SilenceErrors: false,
		SilenceUsage:  true,
	}

	/* configure instance */
	command.AddCommand(
		Init(),
		Plan(),
		Exec(),
		Config(),
	)

	return &command
}

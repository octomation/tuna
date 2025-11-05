package command

import (
	"github.com/spf13/cobra"
)

// Init returns a cobra.Command to initialize project structure for a new assistant.
//
//	$ tuna init <AssistantID>
func Init() *cobra.Command {
	command := cobra.Command{
		Use:   "init <AssistantID>",
		Short: "Initialize project structure for a new assistant",
		Long: `Initialize creates the directory structure for a new assistant:

  AssistantID/
  ├── Input/          # User query files
  ├── Output/         # Generated responses
  └── System prompt/  # System prompt fragments`,

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Initializing assistant: %s...\n", args[0])
		},
	}

	return &command
}

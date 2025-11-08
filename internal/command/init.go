package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/tuna/internal/assistant"
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
  ├── Input/           # User query files
  │   └── example_query.md
  ├── Output/          # Generated responses
  │   └── .gitkeep
  └── System prompt/   # System prompt fragments
      └── fragment_001.md

If the directory already exists, missing parts will be completed.
Existing files will not be overwritten.`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assistantID := args[0]

			// Get current working directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			result, err := assistant.Init(cwd, assistantID)
			if err != nil {
				return err
			}

			// Print summary
			if len(result.Created) > 0 {
				cmd.Println("Created:")
				for _, item := range result.Created {
					cmd.Printf("  + %s\n", item)
				}
			}

			if len(result.Skipped) > 0 {
				cmd.Println("Skipped (already exists):")
				for _, item := range result.Skipped {
					cmd.Printf("  - %s\n", item)
				}
			}

			if len(result.Created) == 0 && len(result.Skipped) > 0 {
				cmd.Println("\nAssistant structure already complete.")
			} else {
				cmd.Printf("\nAssistant '%s' initialized successfully.\n", assistantID)
			}

			return nil
		},
	}

	return &command
}

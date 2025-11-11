package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/tuna/internal/config"
	"go.octolab.org/toolset/tuna/internal/llm"
)

// Config returns a cobra.Command for configuration management.
//
//	$ tuna config <subcommand>
func Config() *cobra.Command {
	command := cobra.Command{
		Use:   "config",
		Short: "Manage tuna configuration",
		Long: `Configuration management commands for tuna.

Subcommands:
  show      Display current configuration
  validate  Validate configuration file
  resolve   Show which provider will be used for a model`,
	}

	command.AddCommand(
		configShow(),
		configValidate(),
		configResolve(),
	)

	return &command
}

// configShow displays current configuration.
func configShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Long: `Display the current tuna configuration.

Shows which configuration file is being used and its contents,
including providers, aliases, and the default provider.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := config.Load()
			if err != nil {
				return err
			}

			cfg := result.Config

			// Show source
			cmd.Printf("Configuration source: %s\n", result.Source)
			if result.Deprecated {
				cmd.Println("Status: Using deprecated environment variables")
			}
			cmd.Println()

			// Show default provider
			cmd.Printf("Default provider: %s\n\n", cfg.DefaultProvider)

			// Show providers
			cmd.Println("Providers:")
			for _, p := range cfg.Providers {
				cmd.Printf("  %s:\n", p.Name)
				cmd.Printf("    Base URL:    %s\n", p.BaseURL)
				cmd.Printf("    API Token:   $%s\n", p.APITokenEnv)
				if p.RateLimit != "" {
					cmd.Printf("    Rate Limit:  %s\n", p.RateLimit)
				}
				if len(p.Models) > 0 {
					cmd.Printf("    Models:      %s\n", strings.Join(p.Models, ", "))
				}
				cmd.Println()
			}

			// Show aliases
			if len(cfg.Aliases) > 0 {
				cmd.Println("Aliases:")
				// Sort aliases for consistent output
				aliases := make([]string, 0, len(cfg.Aliases))
				for alias := range cfg.Aliases {
					aliases = append(aliases, alias)
				}
				sort.Strings(aliases)
				for _, alias := range aliases {
					cmd.Printf("  %s -> %s\n", alias, cfg.Aliases[alias])
				}
			}

			return nil
		},
	}
}

// configValidate validates configuration.
func configValidate() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long: `Validate the tuna configuration file.

Checks for:
  - Valid TOML syntax
  - Required fields (default_provider, providers)
  - Valid rate limit formats
  - No duplicate provider names
  - Default provider exists in providers list`,

		RunE: func(cmd *cobra.Command, args []string) error {
			// Find config file
			configPath, err := config.FindConfigFile()
			if err != nil {
				return fmt.Errorf("no configuration file found\n\nCreate .tuna.toml in your project or ~/.config/tuna.toml")
			}

			// Try to load and validate
			_, err = config.LoadFromFile(configPath)
			if err != nil {
				return err
			}

			cmd.Printf("Configuration is valid: %s\n", configPath)
			return nil
		},
	}
}

// configResolve shows which provider will be used for a model.
func configResolve() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve <model>",
		Short: "Show which provider will be used for a model",
		Long: `Resolve a model name to its full name and provider.

This command shows:
  - The full model name (if an alias was used)
  - Which provider will handle requests for this model

Examples:
  tuna config resolve sonnet
  tuna config resolve gpt-4o
  tuna config resolve unknown-model`,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			model := args[0]

			// Load config
			result, err := config.Load()
			if err != nil {
				return err
			}

			// Create router (we need to check env vars for this)
			router, err := llm.NewRouter(result.Config)
			if err != nil {
				// If router creation fails (missing env vars), we can still
				// do basic resolution without actually creating the clients
				return resolveWithoutRouter(cmd, result.Config, model)
			}

			fullName, provider := router.ResolveModel(model)

			if model != fullName {
				cmd.Printf("%s -> %s -> %s\n", model, fullName, provider)
			} else {
				cmd.Printf("%s -> %s\n", fullName, provider)
			}

			// Check if this model is explicitly mapped or using default
			isDefault := true
			for _, p := range result.Config.Providers {
				for _, m := range p.Models {
					if m == fullName {
						isDefault = false
						break
					}
				}
			}
			if isDefault {
				cmd.Printf("  (using default provider)\n")
			}

			return nil
		},
	}
}

// resolveWithoutRouter resolves model without creating actual clients.
func resolveWithoutRouter(cmd *cobra.Command, cfg *config.Config, model string) error {
	// Resolve alias
	fullName := model
	if name, ok := cfg.Aliases[model]; ok {
		fullName = name
	}

	// Find provider
	provider := cfg.DefaultProvider
	for _, p := range cfg.Providers {
		for _, m := range p.Models {
			if m == fullName {
				provider = p.Name
				break
			}
		}
	}

	if model != fullName {
		cmd.Printf("%s -> %s -> %s\n", model, fullName, provider)
	} else {
		cmd.Printf("%s -> %s\n", fullName, provider)
	}

	// Check if using default
	isDefault := provider == cfg.DefaultProvider
	for _, p := range cfg.Providers {
		if p.Name == provider {
			for _, m := range p.Models {
				if m == fullName {
					isDefault = false
					break
				}
			}
		}
	}
	if isDefault {
		cmd.Printf("  (using default provider)\n")
	}

	return nil
}

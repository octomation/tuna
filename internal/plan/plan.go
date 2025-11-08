package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"

	"go.octolab.org/toolset/tuna/internal/assistant"
)

// Config holds the plan configuration from CLI flags.
type Config struct {
	Models      []string
	Temperature float64
	MaxTokens   int
}

// Plan represents the generated plan structure.
type Plan struct {
	PlanID      string    `toml:"plan_id"`
	AssistantID string    `toml:"assistant_id"`
	Assistant   Assistant `toml:"assistant"`
	Queries     []Query   `toml:"query"`
}

// Assistant holds assistant configuration.
type Assistant struct {
	SystemPrompt string `toml:"system_prompt,multiline"`
	LLM          LLM    `toml:"llm"`
}

// LLM holds LLM configuration.
type LLM struct {
	Models      []string `toml:"models"`
	MaxTokens   int      `toml:"max_tokens"`
	Temperature float64  `toml:"temperature"`
}

// Query represents an input query entry.
type Query struct {
	ID string `toml:"id"`
}

// Result contains the result of plan generation.
type Result struct {
	PlanPath     string
	PlanID       string
	ModelsCount  int
	QueriesCount int
}

// Generate creates a new execution plan for the given assistant.
func Generate(baseDir, assistantID string, cfg Config) (*Result, error) {
	assistantDir := filepath.Join(baseDir, assistantID)

	// Validate assistant directory exists
	if _, err := os.Stat(assistantDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("assistant directory not found: %s", assistantDir)
	}

	// Generate plan ID
	planID := uuid.New().String()

	// Compile system prompt
	systemPrompt, err := assistant.CompileSystemPrompt(assistantDir)
	if err != nil {
		return nil, err
	}

	// Collect queries
	inputDir := filepath.Join(assistantDir, "Input")
	queryFiles, err := assistant.ListFiles(inputDir, assistant.DefaultFilter())
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read input directory: %w", err)
	}

	queries := make([]Query, len(queryFiles))
	for i, filename := range queryFiles {
		queries[i] = Query{ID: filename}
	}

	// Build plan
	plan := Plan{
		PlanID:      planID,
		AssistantID: assistantID,
		Assistant: Assistant{
			SystemPrompt: systemPrompt,
			LLM: LLM{
				Models:      cfg.Models,
				MaxTokens:   cfg.MaxTokens,
				Temperature: cfg.Temperature,
			},
		},
		Queries: queries,
	}

	// Create output directory
	outputDir := filepath.Join(assistantDir, "Output", planID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write plan.toml
	planPath := filepath.Join(outputDir, "plan.toml")
	data, err := toml.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err := os.WriteFile(planPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write plan.toml: %w", err)
	}

	return &Result{
		PlanPath:     planPath,
		PlanID:       planID,
		ModelsCount:  len(cfg.Models),
		QueriesCount: len(queries),
	}, nil
}

// ParseModels splits comma-separated models string into a slice.
func ParseModels(modelsStr string) []string {
	if modelsStr == "" {
		return nil
	}

	parts := strings.Split(modelsStr, ",")
	models := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			models = append(models, trimmed)
		}
	}
	return models
}

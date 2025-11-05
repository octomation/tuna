package exec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.octolab.org/template/tool/internal/llm"
	"go.octolab.org/template/tool/internal/plan"
)

// Options holds execution options.
type Options struct {
	DryRun   bool
	Parallel int
	Continue bool
}

// Result holds execution result.
type Result struct {
	Response     string
	Model        string
	QueryID      string
	PromptTokens int
	OutputTokens int
}

// Executor handles plan execution.
type Executor struct {
	plan         *plan.Plan
	assistantDir string
	llmClient    *llm.Client
	options      Options
}

// New creates a new executor for the given plan.
func New(p *plan.Plan, assistantDir string, llmClient *llm.Client, opts Options) *Executor {
	return &Executor{
		plan:         p,
		assistantDir: assistantDir,
		llmClient:    llmClient,
		options:      opts,
	}
}

// DryRun prints what would be executed without making API calls.
func (e *Executor) DryRun() string {
	var output string

	output += fmt.Sprintf("Plan ID:      %s\n", e.plan.PlanID)
	output += fmt.Sprintf("Assistant ID: %s\n", e.plan.AssistantID)
	output += "\n"

	output += "Models:\n"
	for i, model := range e.plan.Assistant.LLM.Models {
		marker := "  "
		if i == 0 {
			marker = "* " // MVP: first model will be used
		}
		output += fmt.Sprintf("  %s%s\n", marker, model)
	}
	output += "\n"

	output += "Queries:\n"
	for i, query := range e.plan.Queries {
		marker := "  "
		if i == 0 {
			marker = "* " // MVP: first query will be used
		}
		output += fmt.Sprintf("  %s%s\n", marker, query.ID)
	}
	output += "\n"

	output += "LLM Parameters:\n"
	output += fmt.Sprintf("  Temperature: %.1f\n", e.plan.Assistant.LLM.Temperature)
	output += fmt.Sprintf("  Max tokens:  %d\n", e.plan.Assistant.LLM.MaxTokens)
	output += "\n"

	output += "(MVP: only first model and first query will be executed)\n"

	return output
}

// Execute runs the plan (MVP: first query with first model).
func (e *Executor) Execute(ctx context.Context) (*Result, error) {
	// Validate plan has required data
	if len(e.plan.Assistant.LLM.Models) == 0 {
		return nil, fmt.Errorf("no models specified in plan")
	}
	if len(e.plan.Queries) == 0 {
		return nil, fmt.Errorf("no queries specified in plan")
	}

	// MVP: use first model and first query
	model := e.plan.Assistant.LLM.Models[0]
	queryID := e.plan.Queries[0].ID

	// Read query file
	queryPath := filepath.Join(e.assistantDir, "Input", queryID)
	queryContent, err := os.ReadFile(queryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read query file %s: %w", queryPath, err)
	}

	// Make LLM request
	resp, err := e.llmClient.Chat(ctx, llm.ChatRequest{
		Model:        model,
		SystemPrompt: e.plan.Assistant.SystemPrompt,
		UserMessage:  string(queryContent),
		Temperature:  e.plan.Assistant.LLM.Temperature,
		MaxTokens:    e.plan.Assistant.LLM.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		Response:     resp.Content,
		Model:        resp.Model,
		QueryID:      queryID,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

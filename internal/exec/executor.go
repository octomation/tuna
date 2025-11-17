package exec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.octolab.org/toolset/tuna/internal/llm"
	"go.octolab.org/toolset/tuna/internal/plan"
)

// ProgressCallback is called during execution to report progress.
type ProgressCallback func(event ProgressEvent)

// ProgressEvent represents an execution progress event.
type ProgressEvent struct {
	Type     ProgressEventType
	Model    string
	QueryID  string
	Tokens   TokenUsage
	Duration time.Duration
	Err      error
}

// ProgressEventType indicates the type of progress event.
type ProgressEventType int

const (
	EventTaskStart ProgressEventType = iota
	EventTaskDone
	EventTaskError
)

// TokenUsage holds token counts for prompt and output.
type TokenUsage struct {
	Prompt int
	Output int
}

// Options holds execution options.
type Options struct {
	DryRun   bool
	Parallel int
	Continue bool
	OnProgress ProgressCallback
}

// Result holds execution result for a single query-model pair.
type Result struct {
	Response     string
	Model        string
	QueryID      string
	OutputPath   string // Path where response was saved
	PromptTokens int
	OutputTokens int
}

// ExecutionSummary holds results for the entire plan execution.
type ExecutionSummary struct {
	Results      []Result
	TotalQueries int
	TotalModels  int
	TotalTokens  struct {
		Prompt int
		Output int
	}
	Errors []error
}

// Executor handles plan execution.
type Executor struct {
	plan         *plan.Plan
	assistantDir string
	llmClient    llm.ChatClient
	options      Options
}

// New creates a new executor for the given plan.
func New(p *plan.Plan, assistantDir string, llmClient llm.ChatClient, opts Options) *Executor {
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
	output += fmt.Sprintf("Assistant ID: %s\n\n", e.plan.AssistantID)

	output += "Execution matrix:\n"
	for _, model := range e.plan.Assistant.LLM.Models {
		hash := ModelHash(model)
		output += fmt.Sprintf("\n  Model: %s (hash: %s)\n", model, hash)
		for _, query := range e.plan.Queries {
			baseName := strings.TrimSuffix(query.ID, filepath.Ext(query.ID))
			outputPath := fmt.Sprintf("Output/%s/%s/%s_response.md",
				e.plan.PlanID, hash, baseName)
			output += fmt.Sprintf("    %s -> %s\n", query.ID, outputPath)
		}
	}

	output += "\nLLM Parameters:\n"
	output += fmt.Sprintf("  Temperature: %.1f\n", e.plan.Assistant.LLM.Temperature)
	output += fmt.Sprintf("  Max tokens:  %d\n\n", e.plan.Assistant.LLM.MaxTokens)

	total := len(e.plan.Assistant.LLM.Models) * len(e.plan.Queries)
	output += fmt.Sprintf("Total requests: %d (%d models x %d queries)\n",
		total, len(e.plan.Assistant.LLM.Models), len(e.plan.Queries))

	return output
}

// Execute runs the plan for all queries and all models.
func (e *Executor) Execute(ctx context.Context) (*ExecutionSummary, error) {
	// Validate plan has required data
	if len(e.plan.Assistant.LLM.Models) == 0 {
		return nil, fmt.Errorf("no models specified in plan")
	}
	if len(e.plan.Queries) == 0 {
		return nil, fmt.Errorf("no queries specified in plan")
	}

	writer := NewResponseWriter(e.assistantDir, e.plan.PlanID)
	summary := &ExecutionSummary{
		TotalQueries: len(e.plan.Queries),
		TotalModels:  len(e.plan.Assistant.LLM.Models),
	}

	// Iterate over all models
	for _, model := range e.plan.Assistant.LLM.Models {
		// Iterate over all queries
		for _, query := range e.plan.Queries {
			// Notify start
			if e.options.OnProgress != nil {
				e.options.OnProgress(ProgressEvent{
					Type:    EventTaskStart,
					Model:   model,
					QueryID: query.ID,
				})
			}

			start := time.Now()
			result, err := e.executeOne(ctx, model, query.ID, writer)
			duration := time.Since(start)

			if err != nil {
				summary.Errors = append(summary.Errors, fmt.Errorf(
					"model=%s query=%s: %w", model, query.ID, err,
				))
				// Notify error
				if e.options.OnProgress != nil {
					e.options.OnProgress(ProgressEvent{
						Type:     EventTaskError,
						Model:    model,
						QueryID:  query.ID,
						Duration: duration,
						Err:      err,
					})
				}
				continue
			}

			summary.Results = append(summary.Results, *result)
			summary.TotalTokens.Prompt += result.PromptTokens
			summary.TotalTokens.Output += result.OutputTokens

			// Notify done
			if e.options.OnProgress != nil {
				e.options.OnProgress(ProgressEvent{
					Type:    EventTaskDone,
					Model:   model,
					QueryID: query.ID,
					Tokens: TokenUsage{
						Prompt: result.PromptTokens,
						Output: result.OutputTokens,
					},
					Duration: duration,
				})
			}
		}
	}

	return summary, nil
}

// executeOne runs a single query with a single model.
func (e *Executor) executeOne(ctx context.Context, model, queryID string, writer *ResponseWriter) (*Result, error) {
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

	// Save response to file
	outputPath, err := writer.Write(model, queryID, resp.Content)
	if err != nil {
		return nil, err
	}

	return &Result{
		Response:     resp.Content,
		Model:        resp.Model,
		QueryID:      queryID,
		OutputPath:   outputPath,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

// Models returns the list of models from the plan.
func (e *Executor) Models() []string {
	return e.plan.Assistant.LLM.Models
}

// QueryIDs returns the list of query IDs from the plan.
func (e *Executor) QueryIDs() []string {
	ids := make([]string, len(e.plan.Queries))
	for i, q := range e.plan.Queries {
		ids[i] = q.ID
	}
	return ids
}

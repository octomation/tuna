package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.octolab.org/toolset/tuna/internal/response"
)

// ResponseWriter handles saving LLM responses to files.
type ResponseWriter struct {
	baseDir string // {AssistantID}/Output/{plan_id}
}

// NewResponseWriter creates a writer for the given plan output directory.
func NewResponseWriter(assistantDir, planID string) *ResponseWriter {
	return &ResponseWriter{
		baseDir: filepath.Join(assistantDir, "Output", planID),
	}
}

// WriteOptions contains metadata to embed in the response file.
type WriteOptions struct {
	ProviderURL  string
	Model        string
	Duration     time.Duration
	InputTokens  int
	OutputTokens int
}

// Write saves a response to the appropriate file with metadata.
// Path: {baseDir}/{model_hash}/{query_id}_response.md
// Note: This completely overwrites any existing file, including previous ratings.
func (w *ResponseWriter) Write(model, queryID, content string, opts WriteOptions) (string, error) {
	modelDir := filepath.Join(w.baseDir, ModelHash(model))

	// Create model directory if not exists
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build response filename: query_001.md -> query_001_response.md
	baseName := strings.TrimSuffix(queryID, filepath.Ext(queryID))
	responseFile := baseName + "_response.md"
	responsePath := filepath.Join(modelDir, responseFile)

	// Build metadata (rating fields nil = null in YAML)
	meta := &response.Metadata{
		Provider:   opts.ProviderURL,
		Model:      opts.Model,
		Duration:   opts.Duration,
		Input:      opts.InputTokens,
		Output:     opts.OutputTokens,
		ExecutedAt: time.Now(),
		Rating:     nil, // Will be set by tuna view
		RatedAt:    nil, // Will be set by tuna view
	}

	// Format content with metadata
	formatted, err := response.Format(meta, content)
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	// Write response content
	if err := os.WriteFile(responsePath, []byte(formatted), 0644); err != nil {
		return "", fmt.Errorf("failed to write response file: %w", err)
	}

	return responsePath, nil
}

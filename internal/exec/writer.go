package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Write saves a response to the appropriate file.
// Path: {baseDir}/{model_hash}/{query_id}_response.md
func (w *ResponseWriter) Write(model, queryID, content string) (string, error) {
	modelDir := filepath.Join(w.baseDir, ModelHash(model))

	// Create model directory if not exists
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build response filename: query_001.md -> query_001_response.md
	baseName := strings.TrimSuffix(queryID, filepath.Ext(queryID))
	responseFile := baseName + "_response.md"
	responsePath := filepath.Join(modelDir, responseFile)

	// Write response content
	if err := os.WriteFile(responsePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write response file: %w", err)
	}

	return responsePath, nil
}

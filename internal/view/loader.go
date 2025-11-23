// Package view provides functionality for loading and displaying LLM responses.
package view

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.octolab.org/toolset/tuna/internal/exec"
	"go.octolab.org/toolset/tuna/internal/plan"
)

// ResponseGroup represents all model responses for a single input query.
type ResponseGroup struct {
	QueryID   string
	InputPath string
	InputText string
	Responses []ModelResponse
}

// ModelResponse represents a single model's response to a query.
type ModelResponse struct {
	Model     string
	ModelHash string
	FilePath  string
	Content   string
	// Execution metadata
	Provider   string
	Duration   time.Duration
	Input      int
	Output     int
	ExecutedAt time.Time
	// Rating metadata
	Rating  Rating
	RatedAt time.Time
}

// Rating represents the user's rating of a response.
type Rating string

const (
	RatingNone Rating = ""
	RatingGood Rating = "good"
	RatingBad  Rating = "bad"
)

// LoadResponses loads all responses for a plan from disk.
func LoadResponses(planPath string) ([]ResponseGroup, error) {
	p, err := plan.LoadFromPath(planPath)
	if err != nil {
		return nil, err
	}

	assistantDir := plan.AssistantDir(planPath)
	outputDir := filepath.Dir(planPath)

	var groups []ResponseGroup
	for _, query := range p.Queries {
		group := ResponseGroup{
			QueryID:   query.ID,
			InputPath: filepath.Join(assistantDir, "Input", query.ID),
		}

		// Read input content
		content, err := os.ReadFile(group.InputPath)
		if err != nil {
			return nil, err
		}
		group.InputText = string(content)

		// Load responses for each model
		for _, model := range p.Assistant.LLM.Models {
			hash := exec.ModelHash(model)
			respPath := filepath.Join(outputDir, hash, responseFileName(query.ID))

			resp := ModelResponse{
				Model:     model,
				ModelHash: hash,
				FilePath:  respPath,
			}

			// Parse response: extracts metadata from front matter,
			// returns content without front matter for rendering
			if meta, respContent, err := ParseResponse(respPath); err == nil {
				resp.Content = respContent // Already stripped of front matter
				// Execution metadata
				resp.Provider = meta.Provider
				resp.Duration = meta.Duration
				resp.Input = meta.Input
				resp.Output = meta.Output
				resp.ExecutedAt = meta.ExecutedAt
				// Rating metadata
				if meta.Rating != nil {
					resp.Rating = Rating(*meta.Rating)
				}
				if meta.RatedAt != nil {
					resp.RatedAt = *meta.RatedAt
				}
			}

			group.Responses = append(group.Responses, resp)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// responseFileName converts a query ID to a response filename.
// e.g., "query_001.md" -> "query_001_response.md"
func responseFileName(queryID string) string {
	base := strings.TrimSuffix(queryID, filepath.Ext(queryID))
	return base + "_response.md"
}

package view

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/tuna/internal/exec"
)

func TestLoadResponses(t *testing.T) {
	// Create a complete test structure:
	// testdata/
	// └── TestAssistant/
	//     ├── Input/
	//     │   └── query_001.md
	//     └── Output/
	//         └── test-plan-id/
	//             ├── plan.toml
	//             └── {model_hash}/
	//                 └── query_001_response.md

	dir := t.TempDir()
	assistantDir := filepath.Join(dir, "TestAssistant")
	inputDir := filepath.Join(assistantDir, "Input")
	planID := "test-plan-123"
	outputDir := filepath.Join(assistantDir, "Output", planID)

	// Create directories
	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Create input file
	inputContent := "What is the meaning of life?"
	require.NoError(t, os.WriteFile(
		filepath.Join(inputDir, "query_001.md"),
		[]byte(inputContent),
		0644,
	))

	// Create plan.toml
	planContent := `plan_id = "test-plan-123"
assistant_id = "TestAssistant"

[assistant]
system_prompt = "You are a helpful assistant."

[assistant.llm]
models = ["gpt-4", "claude-3"]
max_tokens = 4096
temperature = 0.7

[[query]]
id = "query_001.md"
`
	planPath := filepath.Join(outputDir, "plan.toml")
	require.NoError(t, os.WriteFile(planPath, []byte(planContent), 0644))

	// Create response directories for each model
	model1 := "gpt-4"
	model2 := "claude-3"
	hash1 := exec.ModelHash(model1)
	hash2 := exec.ModelHash(model2)

	model1Dir := filepath.Join(outputDir, hash1)
	model2Dir := filepath.Join(outputDir, hash2)
	require.NoError(t, os.MkdirAll(model1Dir, 0755))
	require.NoError(t, os.MkdirAll(model2Dir, 0755))

	// Create response files
	response1 := `---
rating: good
---

# GPT-4 Response

The meaning of life is 42.`
	response2 := "# Claude Response\n\nThe meaning of life is subjective."

	require.NoError(t, os.WriteFile(
		filepath.Join(model1Dir, "query_001_response.md"),
		[]byte(response1),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(model2Dir, "query_001_response.md"),
		[]byte(response2),
		0644,
	))

	// Test LoadResponses
	groups, err := LoadResponses(planPath)
	require.NoError(t, err)
	require.Len(t, groups, 1)

	group := groups[0]
	assert.Equal(t, "query_001.md", group.QueryID)
	assert.Equal(t, inputContent, group.InputText)
	require.Len(t, group.Responses, 2)

	// Check first response (gpt-4)
	resp1 := group.Responses[0]
	assert.Equal(t, model1, resp1.Model)
	assert.Equal(t, hash1, resp1.ModelHash)
	assert.Equal(t, RatingGood, resp1.Rating)
	assert.Contains(t, resp1.Content, "GPT-4 Response")
	assert.NotContains(t, resp1.Content, "rating:") // Front matter stripped

	// Check second response (claude-3)
	resp2 := group.Responses[1]
	assert.Equal(t, model2, resp2.Model)
	assert.Equal(t, hash2, resp2.ModelHash)
	assert.Equal(t, RatingNone, resp2.Rating)
	assert.Contains(t, resp2.Content, "Claude Response")
}

func TestLoadResponses_MissingResponse(t *testing.T) {
	dir := t.TempDir()
	assistantDir := filepath.Join(dir, "TestAssistant")
	inputDir := filepath.Join(assistantDir, "Input")
	planID := "test-plan-456"
	outputDir := filepath.Join(assistantDir, "Output", planID)

	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	// Create input
	require.NoError(t, os.WriteFile(
		filepath.Join(inputDir, "query_001.md"),
		[]byte("Test query"),
		0644,
	))

	// Create plan.toml
	planContent := `plan_id = "test-plan-456"
assistant_id = "TestAssistant"

[assistant]
system_prompt = "Test"

[assistant.llm]
models = ["gpt-4"]
max_tokens = 4096
temperature = 0.7

[[query]]
id = "query_001.md"
`
	planPath := filepath.Join(outputDir, "plan.toml")
	require.NoError(t, os.WriteFile(planPath, []byte(planContent), 0644))

	// Don't create response file - test that it handles missing responses

	groups, err := LoadResponses(planPath)
	require.NoError(t, err)
	require.Len(t, groups, 1)

	// Response should exist but with empty content
	assert.Len(t, groups[0].Responses, 1)
	assert.Empty(t, groups[0].Responses[0].Content)
	assert.Equal(t, RatingNone, groups[0].Responses[0].Rating)
}

func TestLoadResponses_InvalidPlan(t *testing.T) {
	dir := t.TempDir()
	planPath := filepath.Join(dir, "invalid.toml")
	require.NoError(t, os.WriteFile(planPath, []byte("invalid toml [[["), 0644))

	_, err := LoadResponses(planPath)
	assert.Error(t, err)
}

func TestLoadResponses_PlanNotFound(t *testing.T) {
	_, err := LoadResponses("/nonexistent/plan.toml")
	assert.Error(t, err)
}

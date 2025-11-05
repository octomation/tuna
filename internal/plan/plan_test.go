package plan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestGenerate(t *testing.T) {
	t.Run("creates plan successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		assistantDir := filepath.Join(tmpDir, "test-assistant")

		// Setup assistant structure
		if err := os.MkdirAll(filepath.Join(assistantDir, "Input"), 0755); err != nil {
			t.Fatalf("Failed to create Input dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "Output"), 0755); err != nil {
			t.Fatalf("Failed to create Output dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "System prompt"), 0755); err != nil {
			t.Fatalf("Failed to create System prompt dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(assistantDir, "Input", "query.md"), []byte("test query"), 0644); err != nil {
			t.Fatalf("Failed to create query file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(assistantDir, "System prompt", "prompt.md"), []byte("test prompt"), 0644); err != nil {
			t.Fatalf("Failed to create prompt file: %v", err)
		}

		cfg := Config{
			Models:      []string{"gpt-4", "claude-3"},
			Temperature: 0.5,
			MaxTokens:   2048,
		}

		result, err := Generate(tmpDir, "test-assistant", cfg)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		// Verify plan.toml was created
		if _, err := os.Stat(result.PlanPath); os.IsNotExist(err) {
			t.Error("plan.toml was not created")
		}

		// Verify UUID format (should be 36 characters: 8-4-4-4-12)
		if len(result.PlanID) != 36 {
			t.Errorf("Invalid UUID format: %s", result.PlanID)
		}

		if result.ModelsCount != 2 {
			t.Errorf("Expected 2 models, got %d", result.ModelsCount)
		}

		if result.QueriesCount != 1 {
			t.Errorf("Expected 1 query, got %d", result.QueriesCount)
		}

		// Verify plan.toml content
		data, err := os.ReadFile(result.PlanPath)
		if err != nil {
			t.Fatalf("Failed to read plan.toml: %v", err)
		}

		var plan Plan
		if err := toml.Unmarshal(data, &plan); err != nil {
			t.Fatalf("Failed to unmarshal plan.toml: %v", err)
		}

		if plan.PlanID != result.PlanID {
			t.Errorf("PlanID mismatch: expected %s, got %s", result.PlanID, plan.PlanID)
		}
		if plan.AssistantID != "test-assistant" {
			t.Errorf("AssistantID mismatch: expected test-assistant, got %s", plan.AssistantID)
		}
		if len(plan.Assistant.LLM.Models) != 2 {
			t.Errorf("Expected 2 models in plan, got %d", len(plan.Assistant.LLM.Models))
		}
		if plan.Assistant.LLM.Temperature != 0.5 {
			t.Errorf("Expected temperature 0.5, got %f", plan.Assistant.LLM.Temperature)
		}
		if plan.Assistant.LLM.MaxTokens != 2048 {
			t.Errorf("Expected max_tokens 2048, got %d", plan.Assistant.LLM.MaxTokens)
		}
		if len(plan.Queries) != 1 || plan.Queries[0].ID != "query.md" {
			t.Errorf("Expected query.md, got %v", plan.Queries)
		}
	})

	t.Run("fails for missing assistant directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := Generate(tmpDir, "nonexistent", Config{})
		if err == nil {
			t.Error("Expected error for missing assistant directory")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected error to mention 'not found', got: %v", err)
		}
	})

	t.Run("fails for empty system prompt", func(t *testing.T) {
		tmpDir := t.TempDir()
		assistantDir := filepath.Join(tmpDir, "test-assistant")

		if err := os.MkdirAll(filepath.Join(assistantDir, "Input"), 0755); err != nil {
			t.Fatalf("Failed to create Input dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "Output"), 0755); err != nil {
			t.Fatalf("Failed to create Output dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "System prompt"), 0755); err != nil {
			t.Fatalf("Failed to create System prompt dir: %v", err)
		}

		_, err := Generate(tmpDir, "test-assistant", Config{})
		if err == nil {
			t.Error("Expected error for empty system prompt directory")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("Expected error to mention 'empty', got: %v", err)
		}
	})

	t.Run("succeeds with empty Input directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		assistantDir := filepath.Join(tmpDir, "test-assistant")

		if err := os.MkdirAll(filepath.Join(assistantDir, "Input"), 0755); err != nil {
			t.Fatalf("Failed to create Input dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "Output"), 0755); err != nil {
			t.Fatalf("Failed to create Output dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "System prompt"), 0755); err != nil {
			t.Fatalf("Failed to create System prompt dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(assistantDir, "System prompt", "prompt.md"), []byte("test prompt"), 0644); err != nil {
			t.Fatalf("Failed to create prompt file: %v", err)
		}

		cfg := Config{
			Models:      []string{"gpt-4"},
			Temperature: 0.7,
			MaxTokens:   4096,
		}

		result, err := Generate(tmpDir, "test-assistant", cfg)
		if err != nil {
			t.Fatalf("Generate() should succeed with empty Input, error = %v", err)
		}

		if result.QueriesCount != 0 {
			t.Errorf("Expected 0 queries, got %d", result.QueriesCount)
		}
	})

	t.Run("creates correct output directory structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		assistantDir := filepath.Join(tmpDir, "test-assistant")

		if err := os.MkdirAll(filepath.Join(assistantDir, "Input"), 0755); err != nil {
			t.Fatalf("Failed to create Input dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "Output"), 0755); err != nil {
			t.Fatalf("Failed to create Output dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(assistantDir, "System prompt"), 0755); err != nil {
			t.Fatalf("Failed to create System prompt dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(assistantDir, "System prompt", "prompt.md"), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create prompt file: %v", err)
		}

		cfg := Config{
			Models:      []string{"model"},
			Temperature: 0.7,
			MaxTokens:   4096,
		}

		result, err := Generate(tmpDir, "test-assistant", cfg)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		// Verify path structure: <AssistantID>/Output/<plan_id>/plan.toml
		expectedPath := filepath.Join(assistantDir, "Output", result.PlanID, "plan.toml")
		if result.PlanPath != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, result.PlanPath)
		}
	})
}

func TestParseModels(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single model",
			input:    "gpt-4",
			expected: []string{"gpt-4"},
		},
		{
			name:     "multiple models",
			input:    "gpt-4,claude-3",
			expected: []string{"gpt-4", "claude-3"},
		},
		{
			name:     "models with spaces",
			input:    "gpt-4, claude-3 , gemini",
			expected: []string{"gpt-4", "claude-3", "gemini"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only spaces and commas",
			input:    "  ,  ,  ",
			expected: nil,
		},
		{
			name:     "trailing comma",
			input:    "gpt-4,",
			expected: []string{"gpt-4"},
		},
		{
			name:     "leading comma",
			input:    ",gpt-4",
			expected: []string{"gpt-4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseModels(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseModels(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i, v := range tt.expected {
				if result[i] != v {
					t.Errorf("ParseModels(%q)[%d] = %q, want %q", tt.input, i, result[i], v)
				}
			}
		})
	}
}

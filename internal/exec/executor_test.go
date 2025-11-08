package exec

import (
	"context"
	"strings"
	"testing"

	"go.octolab.org/template/tool/internal/plan"
)

func TestDryRun(t *testing.T) {
	p := &plan.Plan{
		PlanID:      "test-plan",
		AssistantID: "test-assistant",
		Assistant: plan.Assistant{
			SystemPrompt: "Test prompt",
			LLM: plan.LLM{
				Models:      []string{"gpt-4", "claude-3"},
				MaxTokens:   4096,
				Temperature: 0.7,
			},
		},
		Queries: []plan.Query{
			{ID: "query1.md"},
			{ID: "query2.md"},
		},
	}

	executor := New(p, "/tmp/assistant", nil, Options{DryRun: true})
	output := executor.DryRun()

	expectedStrings := []string{
		"Plan ID:      test-plan",
		"Assistant ID: test-assistant",
		"Execution matrix:",
		"gpt-4",
		"claude-3",
		"query1.md",
		"query2.md",
		"Temperature: 0.7",
		"Max tokens:  4096",
		"Total requests: 4 (2 models x 2 queries)",
		"_response.md",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("DryRun output missing %q\nOutput:\n%s", expected, output)
		}
	}

	// Verify model hashes are shown
	gpt4Hash := ModelHash("gpt-4")
	if !strings.Contains(output, gpt4Hash) {
		t.Errorf("DryRun output missing gpt-4 hash %q", gpt4Hash)
	}
}

func TestExecuteValidation(t *testing.T) {
	t.Run("fails for empty models", func(t *testing.T) {
		p := &plan.Plan{
			Assistant: plan.Assistant{
				LLM: plan.LLM{Models: []string{}},
			},
			Queries: []plan.Query{{ID: "q.md"}},
		}

		executor := New(p, "/tmp", nil, Options{})
		_, err := executor.Execute(context.Background())
		if err == nil {
			t.Error("Expected error for empty models")
		}
		if !strings.Contains(err.Error(), "no models specified") {
			t.Errorf("Expected 'no models specified' error, got: %v", err)
		}
	})

	t.Run("fails for empty queries", func(t *testing.T) {
		p := &plan.Plan{
			Assistant: plan.Assistant{
				LLM: plan.LLM{Models: []string{"gpt-4"}},
			},
			Queries: []plan.Query{},
		}

		executor := New(p, "/tmp", nil, Options{})
		_, err := executor.Execute(context.Background())
		if err == nil {
			t.Error("Expected error for empty queries")
		}
		if !strings.Contains(err.Error(), "no queries specified") {
			t.Errorf("Expected 'no queries specified' error, got: %v", err)
		}
	})
}

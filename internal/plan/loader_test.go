package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("loads valid plan", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create plan structure
		planID := "test-plan-id"
		planDir := filepath.Join(tmpDir, "test-assistant", "Output", planID)
		if err := os.MkdirAll(planDir, 0755); err != nil {
			t.Fatal(err)
		}

		planContent := `
plan_id = "test-plan-id"
assistant_id = "test-assistant"

[assistant]
system_prompt = "Test prompt"

[assistant.llm]
models = ["gpt-4"]
max_tokens = 1000
temperature = 0.5

[[query]]
id = "query.md"
`
		if err := os.WriteFile(filepath.Join(planDir, "plan.toml"), []byte(planContent), 0644); err != nil {
			t.Fatal(err)
		}

		plan, planPath, err := Load(tmpDir, planID)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if plan.PlanID != planID {
			t.Errorf("PlanID = %q, want %q", plan.PlanID, planID)
		}
		if plan.AssistantID != "test-assistant" {
			t.Errorf("AssistantID = %q, want %q", plan.AssistantID, "test-assistant")
		}
		if planPath == "" {
			t.Error("planPath should not be empty")
		}
	})

	t.Run("returns error for missing plan", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, _, err := Load(tmpDir, "nonexistent")
		if err == nil {
			t.Error("Expected error for missing plan")
		}
	})

	t.Run("returns error for plan_id mismatch", func(t *testing.T) {
		tmpDir := t.TempDir()

		planDir := filepath.Join(tmpDir, "assistant", "Output", "wrong-id")
		if err := os.MkdirAll(planDir, 0755); err != nil {
			t.Fatal(err)
		}

		planContent := `plan_id = "different-id"`
		if err := os.WriteFile(filepath.Join(planDir, "plan.toml"), []byte(planContent), 0644); err != nil {
			t.Fatal(err)
		}

		_, _, err := Load(tmpDir, "wrong-id")
		if err == nil {
			t.Error("Expected error for plan_id mismatch")
		}
	})
}

func TestAssistantDir(t *testing.T) {
	tests := []struct {
		planPath string
		expected string
	}{
		{"/base/Assistant/Output/plan-id/plan.toml", "/base/Assistant"},
		{"./My Assistant/Output/123/plan.toml", "My Assistant"},
	}

	for _, tt := range tests {
		result := AssistantDir(tt.planPath)
		if result != tt.expected {
			t.Errorf("AssistantDir(%q) = %q, want %q", tt.planPath, result, tt.expected)
		}
	}
}

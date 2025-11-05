package assistant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileSystemPrompt(t *testing.T) {
	t.Run("compiles fragments in order", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptDir := filepath.Join(tmpDir, SystemPromptDir)
		if err := os.MkdirAll(promptDir, 0755); err != nil {
			t.Fatalf("Failed to create prompt dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(promptDir, "fragment_002.md"), []byte("Second"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(promptDir, "fragment_001.md"), []byte("First"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(promptDir, ".hidden.md"), []byte("Hidden"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		result, err := CompileSystemPrompt(tmpDir)
		if err != nil {
			t.Fatalf("CompileSystemPrompt() error = %v", err)
		}

		if !strings.Contains(result, "--- fragment_001.md ---") {
			t.Error("Expected fragment_001.md delimiter")
		}
		if !strings.Contains(result, "--- fragment_002.md ---") {
			t.Error("Expected fragment_002.md delimiter")
		}
		if strings.Contains(result, "hidden") {
			t.Error("Hidden file should be excluded")
		}

		// Check order: fragment_001 should come before fragment_002
		idx1 := strings.Index(result, "fragment_001")
		idx2 := strings.Index(result, "fragment_002")
		if idx1 > idx2 {
			t.Error("Fragments should be sorted alphabetically")
		}

		// Check content is included
		if !strings.Contains(result, "First") {
			t.Error("Expected 'First' content in result")
		}
		if !strings.Contains(result, "Second") {
			t.Error("Expected 'Second' content in result")
		}
	})

	t.Run("returns error for empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(tmpDir, SystemPromptDir), 0755); err != nil {
			t.Fatalf("Failed to create prompt dir: %v", err)
		}

		_, err := CompileSystemPrompt(tmpDir)
		if err == nil {
			t.Error("Expected error for empty system prompt directory")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("Expected error to mention 'empty', got: %v", err)
		}
	})

	t.Run("returns error for missing directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := CompileSystemPrompt(tmpDir)
		if err == nil {
			t.Error("Expected error for missing system prompt directory")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected error to mention 'not found', got: %v", err)
		}
	})

	t.Run("adds trailing newline if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptDir := filepath.Join(tmpDir, SystemPromptDir)
		if err := os.MkdirAll(promptDir, 0755); err != nil {
			t.Fatalf("Failed to create prompt dir: %v", err)
		}

		// Create file without trailing newline
		if err := os.WriteFile(filepath.Join(promptDir, "prompt.md"), []byte("No newline"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		result, err := CompileSystemPrompt(tmpDir)
		if err != nil {
			t.Fatalf("CompileSystemPrompt() error = %v", err)
		}

		if !strings.HasSuffix(result, "\n") {
			t.Error("Expected result to end with newline")
		}
	})

	t.Run("preserves existing trailing newline", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptDir := filepath.Join(tmpDir, SystemPromptDir)
		if err := os.MkdirAll(promptDir, 0755); err != nil {
			t.Fatalf("Failed to create prompt dir: %v", err)
		}

		// Create file with trailing newline
		if err := os.WriteFile(filepath.Join(promptDir, "prompt.md"), []byte("With newline\n"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		result, err := CompileSystemPrompt(tmpDir)
		if err != nil {
			t.Fatalf("CompileSystemPrompt() error = %v", err)
		}

		// Should not add extra newline
		if strings.HasSuffix(result, "\n\n") {
			t.Error("Should not add extra newline when file already has one")
		}
	})

	t.Run("handles multiple files with blank line separator", func(t *testing.T) {
		tmpDir := t.TempDir()
		promptDir := filepath.Join(tmpDir, SystemPromptDir)
		if err := os.MkdirAll(promptDir, 0755); err != nil {
			t.Fatalf("Failed to create prompt dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(promptDir, "a.md"), []byte("Content A\n"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(promptDir, "b.md"), []byte("Content B\n"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		result, err := CompileSystemPrompt(tmpDir)
		if err != nil {
			t.Fatalf("CompileSystemPrompt() error = %v", err)
		}

		// Check that there's a blank line between fragments
		if !strings.Contains(result, "Content A\n\n--- b.md ---") {
			t.Errorf("Expected blank line between fragments, got:\n%s", result)
		}
	})
}

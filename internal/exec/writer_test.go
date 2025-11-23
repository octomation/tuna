package exec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.octolab.org/toolset/tuna/internal/response"
)

func TestResponseWriter(t *testing.T) {
	defaultOpts := WriteOptions{
		ProviderURL:  "https://api.openai.com/v1",
		Model:        "gpt-4",
		Duration:     1500 * time.Millisecond,
		InputTokens:  100,
		OutputTokens: 200,
	}

	t.Run("writes response to correct path with metadata", func(t *testing.T) {
		tmpDir := t.TempDir()

		writer := NewResponseWriter(tmpDir, "test-plan-id")

		// Write response
		path, err := writer.Write("gpt-4", "query_001.md", "Test response content", defaultOpts)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Verify path structure
		expectedPath := filepath.Join(tmpDir, "Output", "test-plan-id",
			ModelHash("gpt-4"), "query_001_response.md")
		if path != expectedPath {
			t.Errorf("Write() path = %q, want %q", path, expectedPath)
		}

		// Verify file was created
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}

		// Should have front matter
		if !strings.HasPrefix(string(content), "---\n") {
			t.Error("Expected file to start with front matter")
		}

		// Parse and verify metadata
		meta, parsedContent, err := response.Parse(path)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if meta.Provider != defaultOpts.ProviderURL {
			t.Errorf("Provider = %q, want %q", meta.Provider, defaultOpts.ProviderURL)
		}
		if meta.Model != defaultOpts.Model {
			t.Errorf("Model = %q, want %q", meta.Model, defaultOpts.Model)
		}
		if meta.Duration != defaultOpts.Duration {
			t.Errorf("Duration = %v, want %v", meta.Duration, defaultOpts.Duration)
		}
		if meta.Input != defaultOpts.InputTokens {
			t.Errorf("Input = %d, want %d", meta.Input, defaultOpts.InputTokens)
		}
		if meta.Output != defaultOpts.OutputTokens {
			t.Errorf("Output = %d, want %d", meta.Output, defaultOpts.OutputTokens)
		}
		if meta.ExecutedAt.IsZero() {
			t.Error("ExecutedAt should be set")
		}
		if meta.Rating != nil {
			t.Error("Rating should be nil")
		}
		if parsedContent != "Test response content" {
			t.Errorf("Content = %q, want %q", parsedContent, "Test response content")
		}
	})

	t.Run("creates directories if not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		writer := NewResponseWriter(tmpDir, "new-plan")

		path, err := writer.Write("claude-3", "test.md", "content", defaultOpts)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File was not created at %q", path)
		}
	})

	t.Run("handles query IDs without extension", func(t *testing.T) {
		tmpDir := t.TempDir()

		writer := NewResponseWriter(tmpDir, "plan-id")

		path, err := writer.Write("gpt-4", "query_no_ext", "content", defaultOpts)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Should still append _response.md
		if filepath.Base(path) != "query_no_ext_response.md" {
			t.Errorf("Expected filename query_no_ext_response.md, got %s", filepath.Base(path))
		}
	})

	t.Run("handles different extensions", func(t *testing.T) {
		tmpDir := t.TempDir()

		writer := NewResponseWriter(tmpDir, "plan-id")

		testCases := []struct {
			queryID      string
			expectedFile string
		}{
			{"query.md", "query_response.md"},
			{"query.txt", "query_response.md"},
			{"my-query.markdown", "my-query_response.md"},
		}

		for _, tc := range testCases {
			path, err := writer.Write("gpt-4", tc.queryID, "content", defaultOpts)
			if err != nil {
				t.Fatalf("Write(%q) error = %v", tc.queryID, err)
			}

			if filepath.Base(path) != tc.expectedFile {
				t.Errorf("Write(%q) filename = %q, want %q",
					tc.queryID, filepath.Base(path), tc.expectedFile)
			}
		}
	})

	t.Run("different models create different directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		writer := NewResponseWriter(tmpDir, "plan-id")

		path1, _ := writer.Write("gpt-4", "query.md", "content1", defaultOpts)
		path2, _ := writer.Write("claude-3", "query.md", "content2", defaultOpts)

		dir1 := filepath.Dir(path1)
		dir2 := filepath.Dir(path2)

		if dir1 == dir2 {
			t.Errorf("Different models should create different directories: %q == %q", dir1, dir2)
		}
	})

	t.Run("overwrites existing file including ratings", func(t *testing.T) {
		tmpDir := t.TempDir()
		writer := NewResponseWriter(tmpDir, "plan-id")

		// Write initial response
		path, err := writer.Write("gpt-4", "query.md", "original content", defaultOpts)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Simulate rating by modifying the file
		meta, content, _ := response.Parse(path)
		rating := "good"
		ratedAt := time.Now()
		meta.Rating = &rating
		meta.RatedAt = &ratedAt
		formatted, _ := response.Format(meta, content)
		os.WriteFile(path, []byte(formatted), 0644)

		// Re-execute (overwrite)
		newOpts := WriteOptions{
			ProviderURL:  "https://api.openai.com/v1",
			Model:        "gpt-4-turbo",
			Duration:     2000 * time.Millisecond,
			InputTokens:  150,
			OutputTokens: 250,
		}
		_, err = writer.Write("gpt-4", "query.md", "new content", newOpts)
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Verify rating was reset
		meta, parsedContent, _ := response.Parse(path)
		if meta.Rating != nil {
			t.Errorf("Rating should be nil after re-execution, got %v", *meta.Rating)
		}
		if meta.RatedAt != nil {
			t.Error("RatedAt should be nil after re-execution")
		}
		if parsedContent != "new content" {
			t.Errorf("Content = %q, want %q", parsedContent, "new content")
		}
		if meta.Model != "gpt-4-turbo" {
			t.Errorf("Model = %q, want %q", meta.Model, "gpt-4-turbo")
		}
	})
}

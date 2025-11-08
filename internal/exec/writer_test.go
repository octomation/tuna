package exec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	t.Run("writes response to correct path", func(t *testing.T) {
		tmpDir := t.TempDir()

		writer := NewResponseWriter(tmpDir, "test-plan-id")

		// Write response
		path, err := writer.Write("gpt-4", "query_001.md", "Test response content")
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}

		// Verify path structure
		expectedPath := filepath.Join(tmpDir, "Output", "test-plan-id",
			ModelHash("gpt-4"), "query_001_response.md")
		if path != expectedPath {
			t.Errorf("Write() path = %q, want %q", path, expectedPath)
		}

		// Verify file was created with correct content
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		if string(content) != "Test response content" {
			t.Errorf("File content = %q, want %q", string(content), "Test response content")
		}
	})

	t.Run("creates directories if not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		writer := NewResponseWriter(tmpDir, "new-plan")

		path, err := writer.Write("claude-3", "test.md", "content")
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

		path, err := writer.Write("gpt-4", "query_no_ext", "content")
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
			path, err := writer.Write("gpt-4", tc.queryID, "content")
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

		path1, _ := writer.Write("gpt-4", "query.md", "content1")
		path2, _ := writer.Write("claude-3", "query.md", "content2")

		dir1 := filepath.Dir(path1)
		dir2 := filepath.Dir(path2)

		if dir1 == dir2 {
			t.Errorf("Different models should create different directories: %q == %q", dir1, dir2)
		}
	})
}

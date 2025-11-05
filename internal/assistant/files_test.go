package assistant

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFiles(t *testing.T) {
	t.Run("filters and sorts files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test files
		files := []string{
			"b_query.md",
			"a_query.txt",
			".hidden.md",
			"c_query.md",
			"readme.pdf", // wrong extension
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", f, err)
			}
		}
		if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}

		result, err := ListFiles(tmpDir, DefaultFilter())
		if err != nil {
			t.Fatalf("ListFiles() error = %v", err)
		}

		expected := []string{"a_query.txt", "b_query.md", "c_query.md"}
		if len(result) != len(expected) {
			t.Fatalf("Expected %d files, got %d: %v", len(expected), len(result), result)
		}
		for i, name := range expected {
			if result[i] != name {
				t.Errorf("Expected %s at index %d, got %s", name, i, result[i])
			}
		}
	})

	t.Run("returns empty slice for empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		result, err := ListFiles(tmpDir, DefaultFilter())
		if err != nil {
			t.Fatalf("ListFiles() error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		_, err := ListFiles("/non/existent/directory", DefaultFilter())
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("handles case-insensitive extensions", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Use different filenames to avoid case-insensitive FS collisions (macOS)
		files := []string{
			"query1.MD",
			"query2.TXT",
			"query3.Md",
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", f, err)
			}
		}

		result, err := ListFiles(tmpDir, DefaultFilter())
		if err != nil {
			t.Fatalf("ListFiles() error = %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 files (case-insensitive), got %d: %v", len(result), result)
		}
	})

	t.Run("ignores .DS_Store and .gitkeep", func(t *testing.T) {
		tmpDir := t.TempDir()

		files := []string{
			".DS_Store",
			".gitkeep",
			"valid.md",
		}
		for _, f := range files {
			if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create file %s: %v", f, err)
			}
		}

		result, err := ListFiles(tmpDir, DefaultFilter())
		if err != nil {
			t.Fatalf("ListFiles() error = %v", err)
		}

		if len(result) != 1 || result[0] != "valid.md" {
			t.Errorf("Expected only valid.md, got %v", result)
		}
	})
}

func TestDefaultFilter(t *testing.T) {
	filter := DefaultFilter()

	if len(filter.Extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(filter.Extensions))
	}
	if filter.Extensions[0] != ".txt" || filter.Extensions[1] != ".md" {
		t.Errorf("Expected [.txt, .md], got %v", filter.Extensions)
	}
	if !filter.IgnoreHidden {
		t.Error("Expected IgnoreHidden to be true")
	}
}

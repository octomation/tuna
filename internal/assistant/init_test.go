package assistant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("creates full structure", func(t *testing.T) {
		tmpDir := t.TempDir()

		result, err := Init(tmpDir, "test-assistant")
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		// Check directories created
		expectedDirs := []string{"Input", "Output", "System prompt"}
		for _, dir := range expectedDirs {
			path := filepath.Join(tmpDir, "test-assistant", dir)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Directory %s was not created", dir)
			}
		}

		// Check files created
		expectedFiles := []string{
			filepath.Join("Input", "example_query.md"),
			filepath.Join("Output", ".gitkeep"),
			filepath.Join("System prompt", "fragment_001.md"),
		}
		for _, file := range expectedFiles {
			path := filepath.Join(tmpDir, "test-assistant", file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("File %s was not created", file)
			}
		}

		if len(result.Created) != 6 { // 3 dirs + 3 files
			t.Errorf("Expected 6 created items, got %d", len(result.Created))
		}
	})

	t.Run("creates files with correct content", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := Init(tmpDir, "content-test")
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		// Check example_query.md content
		queryPath := filepath.Join(tmpDir, "content-test", "Input", "example_query.md")
		content, err := os.ReadFile(queryPath)
		if err != nil {
			t.Fatalf("Failed to read example_query.md: %v", err)
		}
		if string(content) != ExampleQueryContent {
			t.Errorf("example_query.md content mismatch\ngot: %q\nwant: %q", string(content), ExampleQueryContent)
		}

		// Check fragment_001.md content
		fragmentPath := filepath.Join(tmpDir, "content-test", "System prompt", "fragment_001.md")
		content, err = os.ReadFile(fragmentPath)
		if err != nil {
			t.Fatalf("Failed to read fragment_001.md: %v", err)
		}
		if string(content) != Fragment001Content {
			t.Errorf("fragment_001.md content mismatch\ngot: %q\nwant: %q", string(content), Fragment001Content)
		}

		// Check .gitkeep is empty
		gitkeepPath := filepath.Join(tmpDir, "content-test", "Output", ".gitkeep")
		content, err = os.ReadFile(gitkeepPath)
		if err != nil {
			t.Fatalf("Failed to read .gitkeep: %v", err)
		}
		if string(content) != "" {
			t.Errorf(".gitkeep should be empty, got: %q", string(content))
		}
	})

	t.Run("completes partial structure", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create partial structure
		if err := os.MkdirAll(filepath.Join(tmpDir, "partial", "Input"), 0755); err != nil {
			t.Fatalf("Failed to create partial structure: %v", err)
		}

		result, err := Init(tmpDir, "partial")
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		// Input should be skipped (exists)
		// Output and System prompt should be created
		if len(result.Skipped) < 1 {
			t.Error("Expected at least 1 skipped item")
		}

		// Verify Output and System prompt were created
		outputPath := filepath.Join(tmpDir, "partial", "Output")
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Error("Output directory was not created")
		}

		systemPromptPath := filepath.Join(tmpDir, "partial", "System prompt")
		if _, err := os.Stat(systemPromptPath); os.IsNotExist(err) {
			t.Error("System prompt directory was not created")
		}
	})

	t.Run("skips template file when directory not empty", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create structure with custom file in Input
		inputDir := filepath.Join(tmpDir, "existing", "Input")
		if err := os.MkdirAll(inputDir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(inputDir, "custom.md"), []byte("custom"), 0644); err != nil {
			t.Fatalf("Failed to create custom file: %v", err)
		}

		result, err := Init(tmpDir, "existing")
		if err != nil {
			t.Fatalf("Init() error = %v", err)
		}

		// example_query.md should be skipped because Input is not empty
		found := false
		for _, item := range result.Skipped {
			if strings.Contains(item, "example_query.md") && strings.Contains(item, "directory not empty") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected example_query.md to be skipped with 'directory not empty' reason")
		}

		// Verify example_query.md was NOT created
		examplePath := filepath.Join(tmpDir, "existing", "Input", "example_query.md")
		if _, err := os.Stat(examplePath); err == nil {
			t.Error("example_query.md should not have been created")
		}
	})

	t.Run("skips existing files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create full structure first
		_, err := Init(tmpDir, "double-init")
		if err != nil {
			t.Fatalf("First Init() error = %v", err)
		}

		// Run init again
		result, err := Init(tmpDir, "double-init")
		if err != nil {
			t.Fatalf("Second Init() error = %v", err)
		}

		// All items should be skipped
		if len(result.Created) != 0 {
			t.Errorf("Expected 0 created items on second init, got %d", len(result.Created))
		}
		if len(result.Skipped) != 6 { // 3 dirs + 3 files
			t.Errorf("Expected 6 skipped items, got %d", len(result.Skipped))
		}
	})

	t.Run("rejects invalid ID", func(t *testing.T) {
		tmpDir := t.TempDir()

		invalidIDs := []string{
			"",
			".",
			"..",
			"foo/bar",
			"foo\\bar",
			"foo:bar",
		}

		for _, id := range invalidIDs {
			_, err := Init(tmpDir, id)
			if err == nil {
				t.Errorf("Expected error for invalid ID %q", id)
			}
		}
	})
}

package assistant

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitResult contains the result of initialization.
type InitResult struct {
	Created []string
	Skipped []string
}

// Template files content.
const (
	ExampleQueryContent = `# Example Query

Write your user query here.
`
	Fragment001Content = `# Fragment 001

Write your system prompt fragment here.
`
)

// Init creates the directory structure for a new assistant.
func Init(baseDir, assistantID string) (*InitResult, error) {
	if err := ValidateID(assistantID); err != nil {
		return nil, fmt.Errorf("invalid assistant ID: %w", err)
	}

	result := &InitResult{}
	root := filepath.Join(baseDir, assistantID)

	// Define structure
	dirs := []string{
		filepath.Join(root, "Input"),
		filepath.Join(root, "Output"),
		filepath.Join(root, "System prompt"),
	}

	files := []struct {
		path    string
		content string
		dir     string // parent dir to check if empty
	}{
		{filepath.Join(root, "Input", "example_query.md"), ExampleQueryContent, filepath.Join(root, "Input")},
		{filepath.Join(root, "Output", ".gitkeep"), "", filepath.Join(root, "Output")},
		{filepath.Join(root, "System prompt", "fragment_001.md"), Fragment001Content, filepath.Join(root, "System prompt")},
	}

	// Create directories
	for _, dir := range dirs {
		if err := createDir(dir, result); err != nil {
			return nil, err
		}
	}

	// Create files (only if directory is empty or file doesn't exist)
	for _, f := range files {
		if err := createFile(f.path, f.content, f.dir, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func createDir(path string, result *InitResult) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
		result.Created = append(result.Created, path)
	} else if err != nil {
		return fmt.Errorf("failed to check directory %s: %w", path, err)
	} else {
		result.Skipped = append(result.Skipped, path)
	}
	return nil
}

func createFile(path, content, parentDir string, result *InitResult) error {
	// Skip if file already exists
	if _, err := os.Stat(path); err == nil {
		result.Skipped = append(result.Skipped, path)
		return nil
	}

	// Check if parent directory is empty (skip template if not empty)
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", parentDir, err)
	}
	if len(entries) > 0 {
		result.Skipped = append(result.Skipped, path+" (directory not empty)")
		return nil
	}

	// Create the file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	result.Created = append(result.Created, path)
	return nil
}

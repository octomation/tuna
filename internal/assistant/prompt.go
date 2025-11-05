package assistant

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SystemPromptDir is the name of the system prompt directory.
const SystemPromptDir = "System prompt"

// CompileSystemPrompt reads and concatenates all prompt fragments.
// Each fragment is prefixed with "--- <filename> ---" delimiter.
func CompileSystemPrompt(assistantDir string) (string, error) {
	promptDir := filepath.Join(assistantDir, SystemPromptDir)

	files, err := ListFiles(promptDir, DefaultFilter())
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("system prompt directory not found: %s", promptDir)
		}
		return "", fmt.Errorf("failed to read system prompt directory: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("system prompt directory is empty: %s", promptDir)
	}

	var builder strings.Builder
	for i, filename := range files {
		if i > 0 {
			builder.WriteString("\n")
		}

		// Write delimiter
		builder.WriteString(fmt.Sprintf("--- %s ---\n", filename))

		// Read and write content
		content, err := os.ReadFile(filepath.Join(promptDir, filename))
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", filename, err)
		}
		builder.Write(content)

		// Ensure trailing newline
		if len(content) > 0 && content[len(content)-1] != '\n' {
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseResponse_NoFrontMatter(t *testing.T) {
	// Create temp file without front matter
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := "# Response\n\nThis is a response without front matter."
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	meta, parsed, err := ParseResponse(filePath)
	require.NoError(t, err)
	assert.Empty(t, meta.Rating)
	assert.True(t, meta.RatedAt.IsZero())
	assert.Equal(t, content, parsed)
}

func TestParseResponse_WithFrontMatter(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := `---
rating: good
rated_at: 2024-01-15T10:30:00Z
---

# Response

This is a response with front matter.`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	meta, parsed, err := ParseResponse(filePath)
	require.NoError(t, err)
	assert.Equal(t, "good", meta.Rating)
	assert.False(t, meta.RatedAt.IsZero())
	assert.Equal(t, 2024, meta.RatedAt.Year())
	assert.Equal(t, time.January, meta.RatedAt.Month())
	assert.Equal(t, 15, meta.RatedAt.Day())

	// Content should not contain front matter
	assert.NotContains(t, parsed, "---")
	assert.NotContains(t, parsed, "rating:")
	assert.Contains(t, parsed, "# Response")
}

func TestParseResponse_WithExecutionMetadata(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := `---
provider: https://openrouter.ai/api/v1
model: anthropic/claude-sonnet-4
duration: 2.45s
input: 100t
output: 200t
executed_at: 2024-01-15T10:30:00Z
rating: good
rated_at: 2024-01-15T11:00:00Z
---

# Response content`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	meta, parsed, err := ParseResponse(filePath)
	require.NoError(t, err)

	assert.Equal(t, "https://openrouter.ai/api/v1", meta.Provider)
	assert.Equal(t, "anthropic/claude-sonnet-4", meta.Model)
	assert.Equal(t, 2450*time.Millisecond, meta.Duration)
	assert.Equal(t, 100, meta.Input)
	assert.Equal(t, 200, meta.Output)
	assert.Equal(t, "good", meta.Rating)
	assert.Contains(t, parsed, "# Response content")
}

func TestParseResponse_BadRating(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := `---
rating: bad
---

Bad response.`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	meta, parsed, err := ParseResponse(filePath)
	require.NoError(t, err)
	assert.Equal(t, "bad", meta.Rating)
	assert.Contains(t, parsed, "Bad response.")
}

func TestParseResponse_FileNotFound(t *testing.T) {
	_, _, err := ParseResponse("/nonexistent/path/response.md")
	assert.Error(t, err)
}

func TestSaveRating_NewFrontMatter(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	originalContent := "# Response\n\nOriginal content."
	require.NoError(t, os.WriteFile(filePath, []byte(originalContent), 0644))

	err := SaveRating(filePath, RatingGood)
	require.NoError(t, err)

	// Read back and verify
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	content := string(data)

	assert.True(t, strings.HasPrefix(content, "---\n"))
	assert.Contains(t, content, "rating: good")
	assert.Contains(t, content, "rated_at:")
	assert.Contains(t, content, "# Response")
	assert.Contains(t, content, "Original content.")
}

func TestSaveRating_UpdateExisting(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := `---
rating: good
rated_at: 2024-01-15T10:30:00Z
---

# Response`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	err := SaveRating(filePath, RatingBad)
	require.NoError(t, err)

	// Read back and verify
	meta, parsed, err := ParseResponse(filePath)
	require.NoError(t, err)
	assert.Equal(t, "bad", meta.Rating)
	assert.Contains(t, parsed, "# Response")
}

func TestSaveRating_PreservesExecutionMetadata(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := `---
provider: https://openrouter.ai/api/v1
model: claude-sonnet-4
duration: 2.45s
input: 100t
output: 200t
executed_at: 2024-01-15T10:30:00Z
rating: null
rated_at: null
---

# Response`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	err := SaveRating(filePath, RatingGood)
	require.NoError(t, err)

	// Read back and verify execution metadata is preserved
	meta, parsed, err := ParseResponse(filePath)
	require.NoError(t, err)

	assert.Equal(t, "https://openrouter.ai/api/v1", meta.Provider)
	assert.Equal(t, "claude-sonnet-4", meta.Model)
	assert.Equal(t, 2450*time.Millisecond, meta.Duration)
	assert.Equal(t, 100, meta.Input)
	assert.Equal(t, 200, meta.Output)
	assert.Equal(t, "good", meta.Rating)
	assert.False(t, meta.RatedAt.IsZero())
	assert.Contains(t, parsed, "# Response")
}

func TestSaveRating_Unrate_PreservesExecutionMetadata(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := `---
provider: https://openrouter.ai/api/v1
model: claude-sonnet-4
duration: 2.45s
input: 100t
output: 200t
executed_at: 2024-01-15T10:30:00Z
rating: good
rated_at: 2024-01-15T11:00:00Z
---

# Response`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	err := SaveRating(filePath, RatingNone)
	require.NoError(t, err)

	// Read back and verify - execution metadata preserved, rating removed
	meta, parsed, err := ParseResponse(filePath)
	require.NoError(t, err)

	assert.Equal(t, "https://openrouter.ai/api/v1", meta.Provider)
	assert.Equal(t, "claude-sonnet-4", meta.Model)
	assert.Empty(t, meta.Rating)
	assert.True(t, meta.RatedAt.IsZero())
	assert.Contains(t, parsed, "# Response")
}

func TestSaveRating_Unrate_NoExecutionMetadata(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")
	content := `---
rating: good
rated_at: 2024-01-15T10:30:00Z
---

# Response`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	err := SaveRating(filePath, RatingNone)
	require.NoError(t, err)

	// Read back and verify - should have no front matter
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	content = string(data)

	assert.False(t, strings.HasPrefix(content, "---\n"))
	assert.Contains(t, content, "# Response")
}

func TestStripFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no front matter",
			input:    "# Hello\n\nWorld",
			expected: "# Hello\n\nWorld",
		},
		{
			name:     "with front matter",
			input:    "---\nrating: good\n---\n\n# Hello\n\nWorld",
			expected: "# Hello\n\nWorld",
		},
		{
			name:     "front matter with leading newlines",
			input:    "---\nrating: good\n---\n\n\n\n# Hello",
			expected: "# Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripFrontMatter(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseFileName(t *testing.T) {
	tests := []struct {
		queryID  string
		expected string
	}{
		{"query_001.md", "query_001_response.md"},
		{"query_001.txt", "query_001_response.md"},
		{"test.md", "test_response.md"},
		{"8c651548-9b43-4bfa-98d4-c698ad1befc3.md", "8c651548-9b43-4bfa-98d4-c698ad1befc3_response.md"},
	}

	for _, tt := range tests {
		t.Run(tt.queryID, func(t *testing.T) {
			result := responseFileName(tt.queryID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

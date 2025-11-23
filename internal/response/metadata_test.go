package response

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestParseContent(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMeta    *Metadata
		wantContent string
	}{
		{
			name:        "no front matter",
			input:       "# Hello\n\nWorld",
			wantMeta:    &Metadata{},
			wantContent: "# Hello\n\nWorld",
		},
		{
			name: "with execution metadata",
			input: `---
provider: https://openrouter.ai/api/v1
model: claude-sonnet-4
duration: 2.45s
input: 100t
output: 200t
executed_at: 2024-01-15T10:30:00Z
---

# Response`,
			wantMeta: &Metadata{
				Provider:   "https://openrouter.ai/api/v1",
				Model:      "claude-sonnet-4",
				Duration:   2450 * time.Millisecond,
				Input:      100,
				Output:     200,
				ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantContent: "# Response",
		},
		{
			name: "with rating only",
			input: `---
rating: good
rated_at: 2024-01-15T11:00:00Z
---

Content`,
			wantMeta: &Metadata{
				Rating:  ptr("good"),
				RatedAt: ptr(time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)),
			},
			wantContent: "Content",
		},
		{
			name: "with all metadata",
			input: `---
provider: https://openrouter.ai/api/v1
model: anthropic/claude-sonnet-4
duration: 1500ms
input: 500t
output: 300t
executed_at: 2024-01-15T10:30:00Z
rating: good
rated_at: 2024-01-15T11:00:00Z
---

# Full response`,
			wantMeta: &Metadata{
				Provider:   "https://openrouter.ai/api/v1",
				Model:      "anthropic/claude-sonnet-4",
				Duration:   1500 * time.Millisecond,
				Input:      500,
				Output:     300,
				ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Rating:     ptr("good"),
				RatedAt:    ptr(time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)),
			},
			wantContent: "# Full response",
		},
		{
			name: "null rating fields",
			input: `---
provider: https://api.openai.com/v1
model: gpt-4o
rating: null
rated_at: null
---

Response`,
			wantMeta: &Metadata{
				Provider: "https://api.openai.com/v1",
				Model:    "gpt-4o",
				Rating:   nil,
				RatedAt:  nil,
			},
			wantContent: "Response",
		},
		{
			name:        "invalid YAML returns empty metadata",
			input:       "---\n: invalid yaml\n---\n\nContent",
			wantMeta:    &Metadata{},
			wantContent: "---\n: invalid yaml\n---\n\nContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, content, err := ParseContent(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantMeta.Provider, meta.Provider, "provider")
			assert.Equal(t, tt.wantMeta.Model, meta.Model, "model")
			assert.Equal(t, tt.wantMeta.Duration, meta.Duration, "duration")
			assert.Equal(t, tt.wantMeta.Input, meta.Input, "input")
			assert.Equal(t, tt.wantMeta.Output, meta.Output, "output")
			assert.Equal(t, tt.wantMeta.Rating, meta.Rating, "rating")
			assert.Equal(t, tt.wantContent, content, "content")
		})
	}
}

func TestParse(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "response.md")

	content := `---
provider: https://openrouter.ai/api/v1
model: claude-sonnet-4
duration: 2.45s
input: 100t
output: 200t
---

# Response content`

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	meta, parsed, err := Parse(filePath)
	require.NoError(t, err)
	assert.Equal(t, "https://openrouter.ai/api/v1", meta.Provider)
	assert.Equal(t, "claude-sonnet-4", meta.Model)
	assert.Equal(t, 2450*time.Millisecond, meta.Duration)
	assert.Equal(t, 100, meta.Input)
	assert.Equal(t, 200, meta.Output)
	assert.Equal(t, "# Response content", parsed)
}

func TestParse_FileNotFound(t *testing.T) {
	_, _, err := Parse("/nonexistent/path/response.md")
	assert.Error(t, err)
}

func TestFormat(t *testing.T) {
	meta := &Metadata{
		Provider:   "https://openrouter.ai/api/v1",
		Model:      "claude-sonnet-4",
		Duration:   2450 * time.Millisecond,
		Input:      100,
		Output:     200,
		ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	result, err := Format(meta, "# Response")
	require.NoError(t, err)

	// Verify it starts with front matter
	assert.True(t, len(result) > 0 && result[:4] == "---\n")

	// Re-parse to verify round-trip
	parsed, content, err := ParseContent(result)
	require.NoError(t, err)

	assert.Equal(t, meta.Provider, parsed.Provider)
	assert.Equal(t, meta.Model, parsed.Model)
	assert.Equal(t, meta.Duration, parsed.Duration)
	assert.Equal(t, meta.Input, parsed.Input)
	assert.Equal(t, meta.Output, parsed.Output)
	assert.Equal(t, "# Response", content)
}

func TestFormat_WithRating(t *testing.T) {
	meta := &Metadata{
		Provider:   "https://openrouter.ai/api/v1",
		Model:      "claude-sonnet-4",
		Duration:   1500 * time.Millisecond,
		Input:      100,
		Output:     200,
		ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Rating:     ptr("good"),
		RatedAt:    ptr(time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)),
	}

	result, err := Format(meta, "Content")
	require.NoError(t, err)

	// Verify round-trip
	parsed, content, err := ParseContent(result)
	require.NoError(t, err)

	assert.Equal(t, meta.Rating, parsed.Rating)
	assert.NotNil(t, parsed.RatedAt)
	assert.Equal(t, "Content", content)
}

func TestFormat_EmptyMetadata(t *testing.T) {
	meta := &Metadata{}

	result, err := Format(meta, "# Just content\n\nNo metadata here.")
	require.NoError(t, err)

	// Should not have front matter
	assert.False(t, len(result) >= 4 && result[:4] == "---\n")
	assert.Equal(t, "# Just content\n\nNo metadata here.", result)
}

func TestFormat_NilMetadata(t *testing.T) {
	result, err := Format(nil, "Content")
	require.NoError(t, err)
	assert.Equal(t, "Content", result)
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		meta     *Metadata
		expected bool
	}{
		{
			name:     "empty metadata",
			meta:     &Metadata{},
			expected: true,
		},
		{
			name:     "with provider",
			meta:     &Metadata{Provider: "https://example.com"},
			expected: false,
		},
		{
			name:     "with model",
			meta:     &Metadata{Model: "gpt-4"},
			expected: false,
		},
		{
			name:     "with duration",
			meta:     &Metadata{Duration: 100 * time.Millisecond},
			expected: false,
		},
		{
			name:     "with input tokens",
			meta:     &Metadata{Input: 100},
			expected: false,
		},
		{
			name:     "with output tokens",
			meta:     &Metadata{Output: 50},
			expected: false,
		},
		{
			name:     "with executed_at",
			meta:     &Metadata{ExecutedAt: time.Now()},
			expected: false,
		},
		{
			name:     "with rating",
			meta:     &Metadata{Rating: ptr("good")},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.meta.IsEmpty())
		})
	}
}

func TestHasExecutionMetadata(t *testing.T) {
	tests := []struct {
		name     string
		meta     *Metadata
		expected bool
	}{
		{
			name:     "empty metadata",
			meta:     &Metadata{},
			expected: false,
		},
		{
			name:     "rating only - no execution metadata",
			meta:     &Metadata{Rating: ptr("good")},
			expected: false,
		},
		{
			name:     "with provider",
			meta:     &Metadata{Provider: "https://example.com"},
			expected: true,
		},
		{
			name:     "with model",
			meta:     &Metadata{Model: "gpt-4"},
			expected: true,
		},
		{
			name:     "with duration",
			meta:     &Metadata{Duration: 100 * time.Millisecond},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.meta.HasExecutionMetadata())
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{100 * time.Millisecond, "100ms"},
		{999 * time.Millisecond, "999ms"},
		{1000 * time.Millisecond, "1.00s"},
		{1500 * time.Millisecond, "1.50s"},
		{2450 * time.Millisecond, "2.45s"},
		{10000 * time.Millisecond, "10.00s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"100t", 100},
		{"0t", 0},
		{"12345t", 12345},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseTokens(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

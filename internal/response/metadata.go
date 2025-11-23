// Package response provides types and utilities for working with LLM response files.
package response

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Metadata holds all metadata stored in response file front matter.
type Metadata struct {
	// Execution metadata (set by tuna exec)
	Provider   string        `yaml:"provider,omitempty"`
	Model      string        `yaml:"model,omitempty"`
	Duration   time.Duration `yaml:"-"`
	Input      int           `yaml:"-"`
	Output     int           `yaml:"-"`
	ExecutedAt time.Time     `yaml:"executed_at,omitempty"`

	// Rating metadata (set by tuna view)
	Rating  *string    `yaml:"rating"`
	RatedAt *time.Time `yaml:"rated_at"`
}

// metadataYAML is used for custom YAML marshaling/unmarshaling.
type metadataYAML struct {
	Provider   string     `yaml:"provider,omitempty"`
	Model      string     `yaml:"model,omitempty"`
	Duration   string     `yaml:"duration,omitempty"`
	Input      string     `yaml:"input,omitempty"`
	Output     string     `yaml:"output,omitempty"`
	ExecutedAt *time.Time `yaml:"executed_at,omitempty"`
	Rating     *string    `yaml:"rating"`
	RatedAt    *time.Time `yaml:"rated_at"`
}

// MarshalYAML implements custom YAML marshaling for human-readable format.
func (m Metadata) MarshalYAML() (interface{}, error) {
	aux := metadataYAML{
		Provider: m.Provider,
		Model:    m.Model,
		Rating:   m.Rating,
		RatedAt:  m.RatedAt,
	}

	if m.Duration > 0 {
		aux.Duration = formatDuration(m.Duration)
	}
	if m.Input > 0 {
		aux.Input = fmt.Sprintf("%dt", m.Input)
	}
	if m.Output > 0 {
		aux.Output = fmt.Sprintf("%dt", m.Output)
	}
	if !m.ExecutedAt.IsZero() {
		aux.ExecutedAt = &m.ExecutedAt
	}

	return aux, nil
}

// UnmarshalYAML implements custom YAML unmarshaling from human-readable format.
func (m *Metadata) UnmarshalYAML(value *yaml.Node) error {
	var aux metadataYAML
	if err := value.Decode(&aux); err != nil {
		return err
	}

	m.Provider = aux.Provider
	m.Model = aux.Model
	m.Rating = aux.Rating
	m.RatedAt = aux.RatedAt

	if aux.ExecutedAt != nil {
		m.ExecutedAt = *aux.ExecutedAt
	}

	// Parse duration: "2.45s" or "2450ms" -> time.Duration
	if aux.Duration != "" {
		d, err := time.ParseDuration(aux.Duration)
		if err != nil {
			return fmt.Errorf("invalid duration %q: %w", aux.Duration, err)
		}
		m.Duration = d
	}

	// Parse tokens: "1250t" -> int
	m.Input = parseTokens(aux.Input)
	m.Output = parseTokens(aux.Output)

	return nil
}

// parseTokens parses token count from format "1250t".
func parseTokens(s string) int {
	s = strings.TrimSuffix(s, "t")
	n, _ := strconv.Atoi(s)
	return n
}

// formatDuration formats duration in a human-readable way.
// Rounds to milliseconds for cleaner output.
func formatDuration(d time.Duration) string {
	// Round to milliseconds for cleaner display
	ms := d.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	// For seconds, show up to 2 decimal places
	secs := float64(ms) / 1000
	return fmt.Sprintf("%.2fs", secs)
}

// frontMatterRegex matches YAML front matter at the start of a file.
var frontMatterRegex = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n`)

// Parse reads a response file and returns metadata and content separately.
func Parse(filePath string) (*Metadata, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}
	return ParseContent(string(data))
}

// ParseContent parses metadata and content from a string.
func ParseContent(data string) (*Metadata, string, error) {
	meta := &Metadata{}
	content := data

	if matches := frontMatterRegex.FindStringSubmatch(content); len(matches) == 2 {
		if err := yaml.Unmarshal([]byte(matches[1]), meta); err != nil {
			// Invalid YAML - return empty metadata but preserve content
			return &Metadata{}, content, nil
		}
		content = frontMatterRegex.ReplaceAllString(content, "")
	}

	return meta, strings.TrimLeft(content, "\n"), nil
}

// Format combines metadata and content into a response file format.
func Format(meta *Metadata, content string) (string, error) {
	if meta == nil || meta.IsEmpty() {
		return strings.TrimLeft(content, "\n"), nil
	}

	yamlData, err := yaml.Marshal(meta)
	if err != nil {
		return "", err
	}

	return "---\n" + string(yamlData) + "---\n\n" + strings.TrimLeft(content, "\n"), nil
}

// IsEmpty returns true if metadata has no meaningful values.
func (m *Metadata) IsEmpty() bool {
	return m.Provider == "" &&
		m.Model == "" &&
		m.Duration == 0 &&
		m.Input == 0 &&
		m.Output == 0 &&
		m.ExecutedAt.IsZero() &&
		m.Rating == nil
}

// HasExecutionMetadata returns true if execution metadata is present.
func (m *Metadata) HasExecutionMetadata() bool {
	return m.Provider != "" || m.Model != "" || m.Duration > 0 ||
		m.Input > 0 || m.Output > 0 || !m.ExecutedAt.IsZero()
}

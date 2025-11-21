package view

import (
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Metadata holds the rating information stored in YAML front matter.
type Metadata struct {
	Rating  Rating    `yaml:"rating,omitempty"`
	RatedAt time.Time `yaml:"rated_at,omitempty"`
}

// frontMatterRegex matches YAML front matter at the start of a file.
// Matches: ---\n...content...\n---\n
var frontMatterRegex = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n`)

// ParseResponse splits a response file into metadata and content.
// Content is returned without front matter for rendering.
func ParseResponse(filePath string) (*Metadata, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}

	content := string(data)
	meta := &Metadata{}

	// Try to extract front matter
	if matches := frontMatterRegex.FindStringSubmatch(content); len(matches) == 2 {
		if err := yaml.Unmarshal([]byte(matches[1]), meta); err != nil {
			// Invalid YAML, treat as no metadata
			return &Metadata{}, content, nil
		}
		// Remove front matter from content for rendering
		content = frontMatterRegex.ReplaceAllString(content, "")
	}

	return meta, strings.TrimLeft(content, "\n"), nil
}

// SaveRating updates or adds front matter with the rating.
func SaveRating(filePath string, rating Rating) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	meta := Metadata{
		Rating:  rating,
		RatedAt: time.Now(),
	}

	// If rating is empty (unrate), we still save it with empty rating
	if rating == RatingNone {
		meta.Rating = ""
		meta.RatedAt = time.Time{} // Zero time
	}

	// Check if front matter exists
	if matches := frontMatterRegex.FindStringSubmatch(content); len(matches) == 2 {
		// Parse existing front matter and update
		var existing Metadata
		yaml.Unmarshal([]byte(matches[1]), &existing)
		existing.Rating = meta.Rating
		existing.RatedAt = meta.RatedAt
		meta = existing

		// Remove old front matter
		content = frontMatterRegex.ReplaceAllString(content, "")
	}

	// If unrating and no other metadata, just save content without front matter
	if meta.Rating == "" && meta.RatedAt.IsZero() {
		return os.WriteFile(filePath, []byte(strings.TrimLeft(content, "\n")), 0644)
	}

	// Build new front matter
	yamlData, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}

	newContent := "---\n" + string(yamlData) + "---\n\n" + strings.TrimLeft(content, "\n")

	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// StripFrontMatter removes front matter from content for display.
func StripFrontMatter(content string) string {
	return strings.TrimLeft(frontMatterRegex.ReplaceAllString(content, ""), "\n")
}

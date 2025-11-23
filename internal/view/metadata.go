package view

import (
	"os"
	"regexp"
	"strings"
	"time"

	"go.octolab.org/toolset/tuna/internal/response"
)

// frontMatterRegex matches YAML front matter at the start of a file.
// Matches: ---\n...content...\n---\n
var frontMatterRegex = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n`)

// ParseResponse splits a response file into metadata and content.
// Content is returned without front matter for rendering.
func ParseResponse(filePath string) (*response.Metadata, string, error) {
	return response.Parse(filePath)
}

// SaveRating updates or adds front matter with the rating.
// Preserves execution metadata if present.
func SaveRating(filePath string, rating Rating) error {
	meta, content, err := response.Parse(filePath)
	if err != nil {
		return err
	}

	// Update rating fields
	if rating == RatingNone {
		meta.Rating = nil
		meta.RatedAt = nil
	} else {
		r := string(rating)
		t := time.Now()
		meta.Rating = &r
		meta.RatedAt = &t
	}

	// Format with updated metadata
	formatted, err := response.Format(meta, content)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(formatted), 0644)
}

// StripFrontMatter removes front matter from content for display.
func StripFrontMatter(content string) string {
	return strings.TrimLeft(frontMatterRegex.ReplaceAllString(content, ""), "\n")
}

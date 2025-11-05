package assistant

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileFilter defines criteria for filtering files.
type FileFilter struct {
	Extensions   []string // e.g., [".txt", ".md"]
	IgnoreHidden bool     // ignore files starting with "."
}

// DefaultFilter returns the standard filter for assistant files.
func DefaultFilter() FileFilter {
	return FileFilter{
		Extensions:   []string{".txt", ".md"},
		IgnoreHidden: true,
	}
}

// ListFiles returns filtered and sorted list of files in a directory.
// Returns only filenames (not full paths), sorted alphabetically.
func ListFiles(dir string, filter FileFilter) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden files
		if filter.IgnoreHidden && strings.HasPrefix(name, ".") {
			continue
		}

		// Check extension
		ext := strings.ToLower(filepath.Ext(name))
		matched := false
		for _, allowed := range filter.Extensions {
			if ext == allowed {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		files = append(files, name)
	}

	sort.Strings(files)
	return files, nil
}

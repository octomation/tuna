package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Load finds and parses a plan by its ID.
// Searches for plan.toml using glob pattern: */Output/<planID>/plan.toml
func Load(baseDir, planID string) (*Plan, string, error) {
	pattern := filepath.Join(baseDir, "*", "Output", planID, "plan.toml")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, "", fmt.Errorf("failed to search for plan: %w", err)
	}

	if len(matches) == 0 {
		return nil, "", fmt.Errorf("plan not found: %s\nRun 'tuna plan <AssistantID>' to create a plan first", planID)
	}

	if len(matches) > 1 {
		return nil, "", fmt.Errorf("multiple plans found with ID %s: %v", planID, matches)
	}

	planPath := matches[0]

	data, err := os.ReadFile(planPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read plan file: %w", err)
	}

	var plan Plan
	if err := toml.Unmarshal(data, &plan); err != nil {
		return nil, "", fmt.Errorf("failed to parse plan.toml: %w", err)
	}

	if plan.PlanID != planID {
		return nil, "", fmt.Errorf("plan_id mismatch: expected %s, got %s", planID, plan.PlanID)
	}

	return &plan, planPath, nil
}

// AssistantDir returns the assistant directory path from plan.toml path.
func AssistantDir(planPath string) string {
	// planPath: <base>/<AssistantID>/Output/<planID>/plan.toml
	// Go up 3 levels to get AssistantID directory
	return filepath.Dir(filepath.Dir(filepath.Dir(planPath)))
}

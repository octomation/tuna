// Package tui provides terminal user interface components built on Charm libraries.
package tui

import (
	"os"
	"sync"

	"github.com/mattn/go-isatty"
)

var (
	forceNonInteractive bool
	mu                  sync.RWMutex
)

// IsInteractive returns true if the terminal supports interactive TUI.
// It checks:
// - stdout is a TTY
// - CI environment variable is not set
// - non-interactive mode was not forced via SetNonInteractive
func IsInteractive() bool {
	mu.RLock()
	forced := forceNonInteractive
	mu.RUnlock()

	if forced {
		return false
	}

	// Check CI environment
	if os.Getenv("CI") != "" {
		return false
	}

	// Check if stdout is a TTY
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// SetNonInteractive forces non-interactive mode.
// This is typically called when --no-tui flag is set.
func SetNonInteractive() {
	mu.Lock()
	forceNonInteractive = true
	mu.Unlock()
}

// ResetInteractive resets the interactive mode to default (auto-detect).
// This is mainly useful for testing.
func ResetInteractive() {
	mu.Lock()
	forceNonInteractive = false
	mu.Unlock()
}

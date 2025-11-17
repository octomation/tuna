package tui

import (
	"fmt"
	"strings"
)

// RenderCreated renders a list of created items with success styling.
func RenderCreated(items []string) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(Bold.Render("Created:"))
	sb.WriteString("\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("  %s %s\n", SymbolCreated, item))
	}

	return sb.String()
}

// RenderSkipped renders a list of skipped items with muted styling.
func RenderSkipped(items []string) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(Muted.Render("Skipped (already exists):"))
	sb.WriteString("\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("  %s %s\n", SymbolSkipped, Muted.Render(item)))
	}

	return sb.String()
}

// RenderErrors renders a list of errors with error styling.
func RenderErrors(items []string) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(Error.Render("Errors:"))
	sb.WriteString("\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("  %s %s\n", SymbolError, Error.Render(item)))
	}

	return sb.String()
}

// RenderResults renders a list of results with success styling.
func RenderResults(items []string) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(Bold.Render("Results:"))
	sb.WriteString("\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("  %s %s\n", SymbolSuccess, item))
	}

	return sb.String()
}

// RenderList renders a generic list with a custom header and symbol.
func RenderList(header string, items []string, symbol string) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(Bold.Render(header))
	sb.WriteString("\n")

	for _, item := range items {
		sb.WriteString(fmt.Sprintf("  %s %s\n", symbol, item))
	}

	return sb.String()
}

// RenderKeyValue renders a key-value pair with styling.
func RenderKeyValue(key, value string) string {
	return fmt.Sprintf("%s %s", Muted.Render(key+":"), value)
}

// RenderBox renders content in a simple box.
func RenderBox(title, content string) string {
	var sb strings.Builder

	sb.WriteString(Title.Render(title))
	sb.WriteString("\n")
	sb.WriteString(content)

	return sb.String()
}

// RenderSuccess renders a success message.
func RenderSuccess(message string) string {
	return fmt.Sprintf("%s %s", SymbolSuccess, Success.Render(message))
}

// RenderWarning renders a warning message.
func RenderWarning(message string) string {
	return fmt.Sprintf("%s %s", Warning.Render("!"), Warning.Render(message))
}

// RenderError renders an error message.
func RenderError(message string) string {
	return fmt.Sprintf("%s %s", SymbolError, Error.Render(message))
}

// RenderInfo renders an info message.
func RenderInfo(message string) string {
	return fmt.Sprintf("%s %s", Info.Render("ℹ"), Info.Render(message))
}

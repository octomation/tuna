package llm

import "context"

// ChatClient defines the interface for LLM chat operations.
type ChatClient interface {
	// Chat sends a chat completion request and returns the response.
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// Compile-time interface implementation checks.
var _ ChatClient = (*Client)(nil)

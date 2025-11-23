package llm

import (
	"context"
	"fmt"
	"os"
	"time"

	api "github.com/sashabaranov/go-openai"
)

const (
	EnvAPIToken = "LLM_API_TOKEN"
	EnvBaseURL  = "LLM_BASE_URL"
)

// Config holds LLM client configuration.
type Config struct {
	APIToken string
	BaseURL  string
}

// ConfigFromEnv reads LLM configuration from environment variables.
func ConfigFromEnv() (*Config, error) {
	token := os.Getenv(EnvAPIToken)
	if token == "" {
		return nil, fmt.Errorf("missing %s environment variable\n\nSet it with:\n  export %s=your-api-token", EnvAPIToken, EnvAPIToken)
	}

	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("missing %s environment variable\n\nSet it with:\n  export %s=https://api.example.com/v1", EnvBaseURL, EnvBaseURL)
	}

	return &Config{
		APIToken: token,
		BaseURL:  baseURL,
	}, nil
}

// Client wraps the OpenAI-compatible client for LLM interactions.
type Client struct {
	client *api.Client
}

// NewClient creates a new LLM client with the given configuration.
func NewClient(cfg *Config) *Client {
	config := api.DefaultConfig(cfg.APIToken)
	config.BaseURL = cfg.BaseURL

	return &Client{
		client: api.NewClientWithConfig(config),
	}
}

// ChatRequest holds parameters for a chat completion request.
type ChatRequest struct {
	Model        string
	SystemPrompt string
	UserMessage  string
	Temperature  float64
	MaxTokens    int
}

// ChatResponse holds the response from a chat completion.
type ChatResponse struct {
	Content      string
	Model        string        // Resolved model name from API response
	ProviderURL  string        // Provider base URL (set by Router)
	PromptTokens int
	OutputTokens int
	Duration     time.Duration // Request execution time (set by Router)
}

// Chat sends a chat completion request and returns the response.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	resp, err := c.client.CreateChatCompletion(ctx, api.ChatCompletionRequest{
		Model: req.Model,
		Messages: []api.ChatCompletionMessage{
			{Role: api.ChatMessageRoleSystem, Content: req.SystemPrompt},
			{Role: api.ChatMessageRoleUser, Content: req.UserMessage},
		},
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	return &ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		Model:        resp.Model,
		PromptTokens: resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}, nil
}

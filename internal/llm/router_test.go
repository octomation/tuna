package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/tuna/internal/config"
)

func TestNewRouter(t *testing.T) {
	t.Run("creates router with valid config", func(t *testing.T) {
		t.Setenv("OPENROUTER_API_KEY", "test-key-1")
		t.Setenv("ANTHROPIC_API_KEY", "test-key-2")

		cfg := &config.Config{
			DefaultProvider: "openrouter",
			Aliases: map[string]string{
				"sonnet": "claude-sonnet-4-20250514",
				"gpt4":   "gpt-4o",
			},
			Providers: []config.Provider{
				{
					Name:        "openrouter",
					BaseURL:     "https://openrouter.ai/api/v1",
					APITokenEnv: "OPENROUTER_API_KEY",
					RateLimit:   "10rpm",
					Models:      []string{"anthropic/claude-sonnet-4", "openai/gpt-4o"},
				},
				{
					Name:        "anthropic",
					BaseURL:     "https://api.anthropic.com/v1",
					APITokenEnv: "ANTHROPIC_API_KEY",
					Models:      []string{"claude-sonnet-4-20250514"},
				},
			},
		}

		router, err := NewRouter(cfg)
		require.NoError(t, err)

		assert.Len(t, router.providers, 2)
		assert.Equal(t, "openrouter", router.defaultProvider)
		assert.Len(t, router.aliases, 2)
	})

	t.Run("returns error for missing env var", func(t *testing.T) {
		os.Unsetenv("MISSING_API_KEY")

		cfg := &config.Config{
			DefaultProvider: "test",
			Providers: []config.Provider{
				{
					Name:        "test",
					BaseURL:     "https://test.com/v1",
					APITokenEnv: "MISSING_API_KEY",
				},
			},
		}

		_, err := NewRouter(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "MISSING_API_KEY")
	})

	t.Run("creates rate limiter for provider", func(t *testing.T) {
		t.Setenv("TEST_API_KEY", "test-key")

		cfg := &config.Config{
			DefaultProvider: "test",
			Providers: []config.Provider{
				{
					Name:        "test",
					BaseURL:     "https://test.com/v1",
					APITokenEnv: "TEST_API_KEY",
					RateLimit:   "60rpm",
				},
			},
		}

		router, err := NewRouter(cfg)
		require.NoError(t, err)

		assert.NotNil(t, router.rateLimiters["test"])
	})

	t.Run("no rate limiter when not configured", func(t *testing.T) {
		t.Setenv("TEST_API_KEY", "test-key")

		cfg := &config.Config{
			DefaultProvider: "test",
			Providers: []config.Provider{
				{
					Name:        "test",
					BaseURL:     "https://test.com/v1",
					APITokenEnv: "TEST_API_KEY",
					// No RateLimit
				},
			},
		}

		router, err := NewRouter(cfg)
		require.NoError(t, err)

		assert.Nil(t, router.rateLimiters["test"])
	})
}

func TestRouter_ResolveModel(t *testing.T) {
	t.Setenv("TEST_API_KEY", "test-key")

	cfg := &config.Config{
		DefaultProvider: "default",
		Aliases: map[string]string{
			"sonnet": "claude-sonnet-4-20250514",
			"gpt4":   "gpt-4o",
		},
		Providers: []config.Provider{
			{
				Name:        "default",
				BaseURL:     "https://default.com/v1",
				APITokenEnv: "TEST_API_KEY",
			},
			{
				Name:        "anthropic",
				BaseURL:     "https://api.anthropic.com/v1",
				APITokenEnv: "TEST_API_KEY",
				Models:      []string{"claude-sonnet-4-20250514"},
			},
			{
				Name:        "openai",
				BaseURL:     "https://api.openai.com/v1",
				APITokenEnv: "TEST_API_KEY",
				Models:      []string{"gpt-4o", "gpt-4o-mini"},
			},
		},
	}

	router, err := NewRouter(cfg)
	require.NoError(t, err)

	tests := []struct {
		name             string
		model            string
		expectedModel    string
		expectedProvider string
	}{
		{
			name:             "alias resolves to model and provider",
			model:            "sonnet",
			expectedModel:    "claude-sonnet-4-20250514",
			expectedProvider: "anthropic",
		},
		{
			name:             "alias resolves to openai",
			model:            "gpt4",
			expectedModel:    "gpt-4o",
			expectedProvider: "openai",
		},
		{
			name:             "direct model name",
			model:            "gpt-4o-mini",
			expectedModel:    "gpt-4o-mini",
			expectedProvider: "openai",
		},
		{
			name:             "unknown model uses default provider",
			model:            "unknown-model",
			expectedModel:    "unknown-model",
			expectedProvider: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullName, provider := router.ResolveModel(tt.model)
			assert.Equal(t, tt.expectedModel, fullName)
			assert.Equal(t, tt.expectedProvider, provider)
		})
	}
}

func TestRouter_Chat(t *testing.T) {
	t.Run("routes request to correct provider", func(t *testing.T) {
		// Create mock servers for two providers
		anthropicCalls := 0
		anthropicServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			anthropicCalls++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id":      "chatcmpl-123",
				"model":   "claude-sonnet-4-20250514",
				"choices": []map[string]any{{"message": map[string]string{"content": "anthropic response"}}},
				"usage":   map[string]int{"prompt_tokens": 10, "completion_tokens": 20},
			})
		}))
		defer anthropicServer.Close()

		openaiCalls := 0
		openaiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			openaiCalls++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id":      "chatcmpl-456",
				"model":   "gpt-4o",
				"choices": []map[string]any{{"message": map[string]string{"content": "openai response"}}},
				"usage":   map[string]int{"prompt_tokens": 15, "completion_tokens": 25},
			})
		}))
		defer openaiServer.Close()

		t.Setenv("ANTHROPIC_KEY", "test-key-1")
		t.Setenv("OPENAI_KEY", "test-key-2")

		cfg := &config.Config{
			DefaultProvider: "openai",
			Aliases: map[string]string{
				"sonnet": "claude-sonnet-4-20250514",
			},
			Providers: []config.Provider{
				{
					Name:        "anthropic",
					BaseURL:     anthropicServer.URL,
					APITokenEnv: "ANTHROPIC_KEY",
					Models:      []string{"claude-sonnet-4-20250514"},
				},
				{
					Name:        "openai",
					BaseURL:     openaiServer.URL,
					APITokenEnv: "OPENAI_KEY",
					Models:      []string{"gpt-4o"},
				},
			},
		}

		router, err := NewRouter(cfg)
		require.NoError(t, err)

		ctx := context.Background()

		// Request using alias - should go to anthropic
		resp, err := router.Chat(ctx, ChatRequest{
			Model:        "sonnet",
			SystemPrompt: "You are helpful",
			UserMessage:  "Hello",
			Temperature:  0.7,
			MaxTokens:    100,
		})
		require.NoError(t, err)
		assert.Equal(t, "anthropic response", resp.Content)
		assert.Equal(t, 1, anthropicCalls)
		assert.Equal(t, 0, openaiCalls)

		// Request using direct model - should go to openai
		resp, err = router.Chat(ctx, ChatRequest{
			Model:        "gpt-4o",
			SystemPrompt: "You are helpful",
			UserMessage:  "Hello",
			Temperature:  0.7,
			MaxTokens:    100,
		})
		require.NoError(t, err)
		assert.Equal(t, "openai response", resp.Content)
		assert.Equal(t, 1, anthropicCalls)
		assert.Equal(t, 1, openaiCalls)
	})

	t.Run("uses default provider for unknown model", func(t *testing.T) {
		defaultCalls := 0
		defaultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defaultCalls++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id":      "chatcmpl-789",
				"model":   "unknown-model",
				"choices": []map[string]any{{"message": map[string]string{"content": "default response"}}},
				"usage":   map[string]int{"prompt_tokens": 5, "completion_tokens": 10},
			})
		}))
		defer defaultServer.Close()

		t.Setenv("DEFAULT_KEY", "test-key")

		cfg := &config.Config{
			DefaultProvider: "default",
			Providers: []config.Provider{
				{
					Name:        "default",
					BaseURL:     defaultServer.URL,
					APITokenEnv: "DEFAULT_KEY",
				},
			},
		}

		router, err := NewRouter(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		resp, err := router.Chat(ctx, ChatRequest{
			Model:       "unknown-model",
			UserMessage: "Hello",
		})
		require.NoError(t, err)
		assert.Equal(t, "default response", resp.Content)
		assert.Equal(t, 1, defaultCalls)
	})
}

func TestRouter_RateLimiting(t *testing.T) {
	t.Run("rate limiter blocks when limit exceeded", func(t *testing.T) {
		requestCount := 0
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			requestCount++
			mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id":      "chatcmpl-123",
				"model":   "test-model",
				"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
				"usage":   map[string]int{"prompt_tokens": 1, "completion_tokens": 1},
			})
		}))
		defer server.Close()

		t.Setenv("TEST_KEY", "test-key")

		cfg := &config.Config{
			DefaultProvider: "test",
			Providers: []config.Provider{
				{
					Name:        "test",
					BaseURL:     server.URL,
					APITokenEnv: "TEST_KEY",
					RateLimit:   "2rps", // 2 requests per second
				},
			},
		}

		router, err := NewRouter(cfg)
		require.NoError(t, err)

		ctx := context.Background()

		// Make 3 requests quickly - the third should be delayed
		start := time.Now()

		for range 3 {
			_, err := router.Chat(ctx, ChatRequest{
				Model:       "test-model",
				UserMessage: "Hello",
			})
			require.NoError(t, err)
		}

		elapsed := time.Since(start)

		// With 2rps, 3 requests should take at least 500ms (one request waits)
		// We use 400ms to account for timing variations
		assert.GreaterOrEqual(t, elapsed, 400*time.Millisecond, "rate limiting should have delayed requests")
		assert.Equal(t, 3, requestCount)
	})

	t.Run("rate limiter respects context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id":      "chatcmpl-123",
				"model":   "test-model",
				"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
				"usage":   map[string]int{"prompt_tokens": 1, "completion_tokens": 1},
			})
		}))
		defer server.Close()

		t.Setenv("TEST_KEY", "test-key")

		cfg := &config.Config{
			DefaultProvider: "test",
			Providers: []config.Provider{
				{
					Name:        "test",
					BaseURL:     server.URL,
					APITokenEnv: "TEST_KEY",
					RateLimit:   "1rps", // 1 request per second - very slow
				},
			},
		}

		router, err := NewRouter(cfg)
		require.NoError(t, err)

		// First request to consume the token
		ctx := context.Background()
		_, err = router.Chat(ctx, ChatRequest{
			Model:       "test-model",
			UserMessage: "Hello",
		})
		require.NoError(t, err)

		// Second request with cancelled context
		ctxWithCancel, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err = router.Chat(ctxWithCancel, ChatRequest{
			Model:       "test-model",
			UserMessage: "Hello",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit wait cancelled")
	})
}

func TestRouter_Helpers(t *testing.T) {
	t.Setenv("TEST_KEY", "test-key")

	cfg := &config.Config{
		DefaultProvider: "provider1",
		Aliases: map[string]string{
			"alias1": "model1",
			"alias2": "model2",
		},
		Providers: []config.Provider{
			{
				Name:        "provider1",
				BaseURL:     "https://p1.com/v1",
				APITokenEnv: "TEST_KEY",
			},
			{
				Name:        "provider2",
				BaseURL:     "https://p2.com/v1",
				APITokenEnv: "TEST_KEY",
			},
		},
	}

	router, err := NewRouter(cfg)
	require.NoError(t, err)

	t.Run("Providers returns list of provider names", func(t *testing.T) {
		providers := router.Providers()
		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "provider1")
		assert.Contains(t, providers, "provider2")
	})

	t.Run("Aliases returns copy of aliases", func(t *testing.T) {
		aliases := router.Aliases()
		assert.Equal(t, map[string]string{"alias1": "model1", "alias2": "model2"}, aliases)

		// Modify the returned map - should not affect router
		aliases["alias3"] = "model3"
		assert.Len(t, router.aliases, 2)
	})

	t.Run("DefaultProvider returns default provider name", func(t *testing.T) {
		assert.Equal(t, "provider1", router.DefaultProvider())
	})
}

func TestRouter_NilAliases(t *testing.T) {
	t.Setenv("TEST_KEY", "test-key")

	cfg := &config.Config{
		DefaultProvider: "test",
		Aliases:         nil, // nil aliases
		Providers: []config.Provider{
			{
				Name:        "test",
				BaseURL:     "https://test.com/v1",
				APITokenEnv: "TEST_KEY",
			},
		},
	}

	router, err := NewRouter(cfg)
	require.NoError(t, err)

	// Should not panic with nil aliases
	fullName, provider := router.ResolveModel("some-model")
	assert.Equal(t, "some-model", fullName)
	assert.Equal(t, "test", provider)
}

func TestRouter_Chat_ReturnsMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some latency
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-123",
			"model":   "gpt-4o-resolved",
			"choices": []map[string]any{{"message": map[string]string{"content": "response"}}},
			"usage":   map[string]int{"prompt_tokens": 100, "completion_tokens": 200},
		})
	}))
	defer server.Close()

	t.Setenv("TEST_KEY", "test-key")

	cfg := &config.Config{
		DefaultProvider: "test",
		Providers: []config.Provider{
			{
				Name:        "test",
				BaseURL:     server.URL,
				APITokenEnv: "TEST_KEY",
			},
		},
	}

	router, err := NewRouter(cfg)
	require.NoError(t, err)

	resp, err := router.Chat(context.Background(), ChatRequest{
		Model:       "gpt-4o",
		UserMessage: "Hello",
	})
	require.NoError(t, err)

	// Verify response content
	assert.Equal(t, "response", resp.Content)
	assert.Equal(t, "gpt-4o-resolved", resp.Model)
	assert.Equal(t, 100, resp.PromptTokens)
	assert.Equal(t, 200, resp.OutputTokens)

	// Verify metadata fields are populated
	assert.Equal(t, server.URL, resp.ProviderURL, "ProviderURL should match server URL")
	assert.GreaterOrEqual(t, resp.Duration, 50*time.Millisecond, "Duration should be at least the simulated latency")
	assert.Less(t, resp.Duration, 1*time.Second, "Duration should be reasonable")
}

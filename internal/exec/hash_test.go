package exec

import "testing"

func TestModelHash(t *testing.T) {
	t.Run("returns 8 character hash", func(t *testing.T) {
		hash := ModelHash("gpt-4")
		if len(hash) != 8 {
			t.Errorf("ModelHash(%q) = %q, want 8 chars, got %d", "gpt-4", hash, len(hash))
		}
	})

	t.Run("is deterministic", func(t *testing.T) {
		h1 := ModelHash("gpt-4")
		h2 := ModelHash("gpt-4")
		if h1 != h2 {
			t.Errorf("ModelHash not deterministic: %q != %q", h1, h2)
		}
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		h1 := ModelHash("gpt-4")
		h2 := ModelHash("gpt-3.5")
		if h1 == h2 {
			t.Errorf("ModelHash collision: %q and %q both produce %q", "gpt-4", "gpt-3.5", h1)
		}
	})

	t.Run("handles various model names", func(t *testing.T) {
		models := []string{
			"gpt-4",
			"claude-3-opus",
			"qwen3-32b",
			"llama-2-70b-chat",
		}

		hashes := make(map[string]string)
		for _, model := range models {
			hash := ModelHash(model)
			if len(hash) != 8 {
				t.Errorf("ModelHash(%q) = %q, want 8 chars", model, hash)
			}
			if existing, ok := hashes[hash]; ok {
				t.Errorf("Hash collision between %q and %q: both produce %q", existing, model, hash)
			}
			hashes[hash] = model
		}
	})
}

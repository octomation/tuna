> # 🐟 Tuna
>
> Dead simple prompt tuner.

**Test and compare LLM prompts across multiple models in seconds.**

Tuna automates prompt engineering workflows. Write your prompts, define test queries,
and instantly see how different models respond — all from the command line.

## Why Tuna?

Iterating on system prompts is tedious: copy-paste into chat interfaces, switch between models,
manually compare outputs. Tuna eliminates this friction.

- **Organize** prompts and test queries in a simple folder structure
- **Execute** the same queries across multiple models with one command
- **Compare** responses side-by-side to find what works best

## Quick Start

```bash
# Initialize a new assistant
tuna init my-assistant

# Edit your system prompt
echo "You are a helpful assistant." > my-assistant/System\ prompt/fragment_001.md

# Add test queries
echo "Explain quantum computing in simple terms." > my-assistant/Input/query_001.md

# Create an execution plan (use aliases or full model names)
tuna plan my-assistant --models sonnet,gpt4

# Run it
tuna exec <plan-id>
```

Results are saved to `my-assistant/Output/<plan-id>/` for easy comparison.

## Project Structure

```
my-assistant/
├── Input/              # Your test queries
│   └── query_001.md
├── Output/             # Generated responses
│   └── <plan-id>/
│       └── <model>/
└── System prompt/      # Prompt fragments (concatenated in order)
    └── fragment_001.md
```

## Configuration

Create `.tuna.toml` in your project root (or `~/.config/tuna.toml` for global config):

```toml
default_provider = "openrouter"

[aliases]
sonnet = "claude-sonnet-4-20250514"
gpt4 = "gpt-4o"

[[providers]]
name = "openrouter"
base_url = "https://openrouter.ai/api/v1"
api_token_env = "OPENROUTER_API_KEY"  # or use api_token = "sk-..." directly
models = ["anthropic/claude-sonnet-4", "openai/gpt-4o"]
```

See [.tuna.toml.example](.tuna.toml.example) for a complete configuration reference.

## Installation

```bash
go install go.octolab.org/toolset/tuna@latest
```

## License

MIT

<p align="right">made with ❤️ for everyone by <a href="https://www.octolab.org/">OctoLab</a></p>

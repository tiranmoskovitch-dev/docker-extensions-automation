# Docker Compose Native Model Serving

Use the `models:` top-level Compose key to serve LLMs without Ollama.

## How It Works

Docker Desktop 4.43+ includes Model Runner, a built-in LLM inference engine. Instead of running a separate Ollama container, you declare models directly in `compose.yaml`:

```yaml
models:
  qwen3:
    model: ai/qwen3:14B-Q6_K
```

Docker automatically:
1. Pulls the model on first `compose up`
2. Serves it via an OpenAI-compatible API
3. Injects `MODEL_RUNNER_URL` and `MODEL_RUNNER_MODEL` into your services
4. Manages GPU allocation

## vs. Ollama

| Feature | Compose Models | Ollama |
|---------|---------------|--------|
| Setup | Declarative in compose.yaml | Separate service |
| API | OpenAI-compatible | Ollama API + OpenAI |
| GPU | Auto-managed by Docker | Manual config |
| Windows GPU | Docker Desktop 4.43+ | Native support |
| Model format | GGUF via Model Runner | GGUF + safetensors |
| Management | `docker compose` commands | `ollama` CLI |

## Quick Start

```bash
docker compose up
```

## When to Use This

- New projects where you want zero LLM infrastructure
- CI/CD pipelines that need reproducible model serving
- When you want one `compose.yaml` to define everything (app + model + tools)

## When to Keep Ollama

- You need models not yet in Docker's registry
- Windows GPU support is critical (Ollama has broader support today)
- You use Ollama-specific features (embeddings API, model customization)

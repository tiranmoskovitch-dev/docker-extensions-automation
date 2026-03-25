"""
Example: Using Docker Compose native model serving.

The `models:` key in compose.yaml declares LLMs that Docker serves automatically.
The MODEL_RUNNER_URL and MODEL_RUNNER_MODEL env vars are injected by Compose.

This gives you an OpenAI-compatible API without running Ollama separately.
"""

import os
from openai import OpenAI

# Docker Compose injects these from the `models:` section
MODEL_URL = os.environ.get("MODEL_RUNNER_URL", "http://model-runner.docker.internal/v1")
MODEL_NAME = os.environ.get("MODEL_RUNNER_MODEL", "ai/qwen3:14B-Q6_K")


def main():
    client = OpenAI(base_url=MODEL_URL, api_key="not-needed")

    print(f"Using model: {MODEL_NAME}")
    print(f"Endpoint: {MODEL_URL}")
    print()

    response = client.chat.completions.create(
        model=MODEL_NAME,
        messages=[
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "Explain Docker Compose models in 2 sentences."},
        ],
        max_tokens=200,
    )

    print(f"Response: {response.choices[0].message.content}")


if __name__ == "__main__":
    main()

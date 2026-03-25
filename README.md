# Docker Extensions for Automation

Build and manage AI agents, MCP servers, LLM models, and infrastructure stacks — all from Docker Desktop.

A suite of **6 Docker Desktop extensions** + **ready-to-use examples** for turning Docker into a full automation control plane.

---

## Extensions

| Extension | What It Does | Backend | Status |
|-----------|-------------|---------|--------|
| **Stack Commander** | Manage all Compose stacks from one UI — start, stop, restart, logs | Go | Scaffolded |
| **Agent Orchestrator** | Visual AI agent lifecycle management — health, tasks, resources | Go | Scaffolded |
| **MCP Gateway Manager** | Configure MCP Gateway, register MCP servers, browse tools | Go | Scaffolded |
| **Model Runner Dashboard** | Native LLM serving via Compose `models:` key — pull, monitor, compare vs Ollama | Go | Scaffolded |
| **Port Dashboard** | Real-time port mapper with conflict detection and one-click open | SDK-only | Scaffolded |
| **Health Monitor** | Container health tracking, auto-restart policies, resource alerts | Go | Scaffolded |

## Examples

| Example | What It Demonstrates |
|---------|---------------------|
| **[MCP Gateway](examples/mcp-gateway/)** | Secure MCP tool access for containerized agents via `docker/mcp-gateway:latest` |
| **[Compose Models](examples/compose-models/)** | Native LLM serving with the `models:` Compose key — zero Ollama needed |

---

## Architecture

```
                        Docker Desktop
  ┌──────────┬──────────┬──────────┬──────────┬─────────┬────────┐
  │  Stack   │  Agent   │   MCP    │  Model   │  Port   │ Health │
  │Commander │Orchestr. │ Gateway  │  Runner  │Dashboard│Monitor │
  ├──────────┴──────────┴──────────┴──────────┴─────────┴────────┤
  │                  Docker Extensions SDK                        │
  │    listContainers | cli.exec | vm.service | events stream     │
  ├───────────────────────────────────────────────────────────────┤
  │                  Docker Engine + Compose                      │
  ├──────────────┬───────────────┬────────────────┬───────────────┤
  │ Compose      │  MCP Gateway  │  Model Runner  │    Agent      │
  │ Stacks       │  (tool broker)│  (LLM serving) │  Containers   │
  └──────────────┴───────────────┴────────────────┴───────────────┘
```

Each extension is **independent** — install one or all. No cross-dependencies.

---

## Quick Start

### Prerequisites

- Docker Desktop 4.43.0+ (for Model Runner and `models:` key)
- Docker Extensions SDK enabled
- Go 1.22+ and Node.js 20+ (for building from source)

### Install an Extension Locally

```bash
# Build and install Stack Commander
cd extensions/stack-commander
docker build -t tiranmoskovitch/stack-commander .
docker extension install tiranmoskovitch/stack-commander
```

### Run an Example

```bash
# Try the MCP Gateway pattern
cd examples/mcp-gateway
docker compose up --build

# Try native model serving
cd examples/compose-models
docker compose up
```

### Enable Extension Development Mode

In Docker Desktop: Settings > Extensions > uncheck "Allow only extensions distributed through the Docker Marketplace"

---

## MCP Gateway Pattern

The [MCP Gateway example](examples/mcp-gateway/) shows how to give AI agents secure access to tools without host access:

```yaml
services:
  mcp-gateway:
    image: docker/mcp-gateway:latest
    use_api_socket: true
    command:
      - --transport=sse
      - --servers=postgres,filesystem
      - --tools=query,read_file

  agent:
    environment:
      - MCP_SERVER_URL=http://mcp-gateway:8811/sse
```

**Why this matters:**
- Agents only see the gateway endpoint — no Docker socket, no host filesystem
- Gateway controls exactly which tools are exposed
- Secrets stay in Docker secrets, not container env vars
- Works with any MCP server that uses stdio transport

---

## Compose Native Models

The [Compose Models example](examples/compose-models/) demonstrates Docker's built-in LLM serving:

```yaml
# Top-level key — Docker handles everything
models:
  qwen3:
    model: ai/qwen3:14B-Q6_K

services:
  agent:
    # MODEL_RUNNER_URL and MODEL_RUNNER_MODEL are auto-injected
    environment:
      - MODEL_RUNNER_URL
      - MODEL_RUNNER_MODEL
```

**What Docker does automatically:**
1. Pulls the model on first `compose up`
2. Serves it via OpenAI-compatible API
3. Injects endpoint URL into your services
4. Manages GPU allocation

**vs. Ollama:** Compose models are declarative and zero-config. Keep Ollama if you need broader Windows GPU support or Ollama-specific features today.

---

## Project Structure

```
docker-extensions-automation/
├── extensions/
│   ├── stack-commander/          # P0 — Compose stack management
│   │   ├── metadata.json
│   │   ├── Dockerfile
│   │   ├── docker-compose.yaml
│   │   ├── backend/              # Go backend
│   │   └── ui/                   # React frontend
│   ├── agent-orchestrator/       # P1 — AI agent lifecycle
│   ├── mcp-gateway-manager/      # P1 — MCP server management
│   ├── model-runner-dashboard/   # P1 — Native LLM management
│   ├── port-dashboard/           # P2 — Port mapping & health
│   └── health-monitor/           # P2 — Container health & alerts
├── examples/
│   ├── mcp-gateway/              # MCP Gateway pattern demo
│   └── compose-models/           # Native model serving demo
├── shared/
│   └── sdk-utils/                # Shared React hooks for SDK
├── docs/
│   ├── PROJECT_PLAN.md           # Roadmap and milestones
│   └── ARCHITECTURE.md           # Technical architecture
├── LICENSE                       # MIT
└── README.md                     # This file
```

## Tech Stack

- **Frontend:** React 18 + TypeScript + Tailwind CSS
- **Backend:** Go 1.22 (where persistent state is needed)
- **SDK:** `@docker/extension-api-client` v0.3.4
- **Build:** Multi-stage Dockerfiles
- **Distribution:** Docker Hub under `tiranmoskovitch/` namespace

## Roadmap

| Phase | What | Status |
|-------|------|--------|
| 1 | Stack Commander MVP | In Progress |
| 2 | Agent Orchestrator + MCP Gateway + Model Runner | Planned |
| 3 | Port Dashboard + Health Monitor | Planned |
| 4 | Docker Hub marketplace publish | Planned |

## Contributing

PRs welcome. Each extension is independent — you can contribute to one without touching others.

## License

[MIT](LICENSE)

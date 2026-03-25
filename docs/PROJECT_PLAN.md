# Project Plan: Docker Extensions for Automation

## Vision

A suite of Docker Desktop extensions that turn Docker into a full automation control plane — managing stacks, AI agents, MCP servers, LLM models, ports, and container health from a single UI.

## Extension Suite (6 Extensions)

### 1. Stack Commander
**Priority:** P0 (build first)
**Purpose:** Manage multiple Docker Compose stacks from Docker Desktop UI.

Features:
- List all stacks with real-time status (running/stopped/partial)
- Start/stop/restart individual stacks or all at once
- View aggregated logs per stack with filtering
- Environment variable editor per stack
- One-click "bring up everything" button
- Stack dependency visualization

Tech: React frontend + Go backend (watches docker events stream)

### 2. Agent Orchestrator
**Priority:** P1
**Purpose:** Visual management of AI agent containers — lifecycle, health, task queues.

Features:
- Agent registry with status indicators
- Start/stop/restart individual agents
- Live log streaming per agent
- Task queue visualization (if agents expose /tasks endpoint)
- Resource usage per agent (CPU/memory/GPU)
- Agent-to-agent communication topology map

Tech: React frontend + Go backend + WebSocket for live updates

### 3. MCP Gateway Manager
**Priority:** P1
**Purpose:** Configure and manage MCP Gateway instances and connected MCP servers.

Features:
- Visual MCP server registry (add/remove/configure)
- Gateway health and connection status
- Tool inventory — list all tools exposed by each MCP server
- Access control configuration (which agents can use which tools)
- Request/response logging and debugging
- One-click deploy of `docker/mcp-gateway:latest`

Tech: React frontend + Go backend + MCP Gateway API

### 4. Model Runner Dashboard
**Priority:** P1
**Purpose:** Manage Docker Compose `models:` key — native LLM serving without Ollama.

Features:
- Browse available models (ai/qwen3, ai/llama3, etc.)
- Pull/remove models from Docker Model Runner
- Monitor inference requests and latency
- Model configuration (quantization, context size)
- Compare with Ollama setup — migration helper
- GPU allocation and memory tracking

Tech: React frontend + Go backend + Docker Model Runner API

### 5. Port Dashboard
**Priority:** P2
**Purpose:** Real-time port/service mapper with health checks.

Features:
- Auto-discover all exposed ports across all containers
- Health check status per port (HTTP probe)
- Service name resolution (container name + port = friendly name)
- Conflict detection (two containers claiming same port)
- Quick-launch: click port to open in browser
- Export port map as markdown/JSON

Tech: React frontend (lightweight, no backend needed — SDK calls only)

### 6. Health Monitor
**Priority:** P2
**Purpose:** Container health tracking with auto-restart and alerting.

Features:
- Health check status for all containers
- Auto-restart policies (configurable per container)
- Health history timeline
- Alert channels (desktop notification, webhook, Slack)
- Resource threshold alerts (CPU > 90%, memory > 80%)
- Downtime tracking and uptime percentages

Tech: React frontend + Go backend (persistent health DB)

## Milestones

| Phase | Target | Deliverables |
|-------|--------|-------------|
| Phase 1 | Stack Commander MVP | Single extension, installable locally, manages compose stacks |
| Phase 2 | Agent + MCP + Models | Three extensions covering AI/automation workflows |
| Phase 3 | Port + Health | Observability extensions |
| Phase 4 | Docker Hub publish | All 6 extensions published to Docker Hub marketplace |

## Architecture Principles

1. **Each extension is independent** — install one or all, no cross-dependencies
2. **SDK-first** — use Docker Extensions SDK APIs, not raw Docker socket
3. **No host binaries unless necessary** — prefer containerized backends
4. **Real-time by default** — WebSocket/SSE for live updates, no polling
5. **Zero config** — extensions auto-discover stacks, containers, ports
6. **Secure** — no host filesystem access beyond what Docker provides

## Tech Stack

- **Frontend:** React 18 + TypeScript + Tailwind CSS + @docker/extension-api-client
- **Backend:** Go 1.22+ (where persistent state is needed)
- **Build:** Multi-stage Dockerfiles
- **Distribution:** Docker Hub under `tiranmoskovitch/` namespace

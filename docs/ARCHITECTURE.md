# Architecture

## How Docker Extensions Work

```
Docker Desktop
├── Extensions Tab (UI host)
│   └── Extension Frontend (React app in iframe)
│       ├── @docker/extension-api-client
│       │   ├── docker.cli.exec()        → Run docker commands
│       │   ├── docker.listContainers()   → Query containers
│       │   ├── docker.listImages()       → Query images
│       │   ├── extension.vm.service      → Call backend APIs
│       │   ├── host.openUrl()            → Open browser
│       │   └── desktopUI.toast.*()       → Notifications
│       └── Custom UI components
├── Extension Backend (container in Docker VM)
│   ├── REST API on Unix socket
│   ├── Docker socket access (/var/run/docker.sock)
│   └── Persistent storage (volumes)
└── Host Binaries (optional, installed on host)
    └── CLI tools executed via host.exec()
```

## Extension Anatomy

Each extension is a Docker image with this structure:

```
extension-image/
├── metadata.json          # Extension manifest
├── docker-compose.yaml    # Backend service definition (optional)
├── Dockerfile             # Multi-stage build
├── ui/                    # Frontend assets (built React app)
│   ├── index.html
│   └── assets/
└── backend/               # Go backend binary (optional)
```

### metadata.json Schema

```json
{
  "icon": "docker.svg",
  "ui": {
    "dashboard-tab": {
      "title": "Extension Name",
      "root": "/ui",
      "src": "index.html"
    }
  },
  "vm": {
    "composefile": "docker-compose.yaml"
  },
  "host": {
    "binaries": []
  }
}
```

## Communication Patterns

### Frontend → Docker Engine
```typescript
// List all running containers
const containers = await ddClient.docker.listContainers();

// Execute docker compose command
const result = await ddClient.docker.cli.exec("compose", [
  "-f", "/path/to/compose.yaml",
  "ps", "--format", "json"
]);
const services = result.parseJsonLines();
```

### Frontend → Backend Service
```typescript
// GET request to backend
const data = await ddClient.extension.vm?.service?.get("/api/stacks");

// POST request to backend
await ddClient.extension.vm?.service?.post("/api/stacks/start", { name: "core" });
```

### Backend → Docker Socket
```go
// Go backend has direct Docker socket access
cli, _ := client.NewClientWithOpts(client.FromEnv)
containers, _ := cli.ContainerList(ctx, container.ListOptions{})
```

### Real-time Updates
```typescript
// Stream docker events
ddClient.docker.cli.exec("events", ["--format", "json"], {
  stream: {
    onOutput(data) { handleEvent(JSON.parse(data.stdout)); },
    onClose(exitCode) { console.log("Stream closed"); },
    onError(err) { console.error(err); }
  }
});
```

## Extension Suite Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Desktop                        │
├──────────┬──────────┬──────────┬──────────┬─────────────┤
│  Stack   │  Agent   │   MCP    │  Model   │   Port  │Health│
│Commander │Orchestr. │ Gateway  │  Runner  │Dashboard│Monitor│
├──────────┴──────────┴──────────┴──────────┴─────────┴──────┤
│              Docker Extensions SDK                         │
├────────────────────────────────────────────────────────────┤
│              Docker Engine + Compose                       │
├────────────────────────────────────────────────────────────┤
│  Compose Stacks  │  MCP Gateway  │  Model Runner  │ Agents │
└──────────────────┴───────────────┴────────────────┴────────┘
```

Each extension operates independently but they complement each other:
- **Stack Commander** manages the infrastructure layer
- **Agent Orchestrator** manages the AI agent layer
- **MCP Gateway Manager** manages tool access for agents
- **Model Runner Dashboard** manages LLM inference
- **Port Dashboard** provides observability
- **Health Monitor** provides reliability

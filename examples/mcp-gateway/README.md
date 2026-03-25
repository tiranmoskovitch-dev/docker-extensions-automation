# MCP Gateway Pattern

Secure MCP tool access for containerized AI agents.

## What This Solves

When running AI agents in containers, they need access to tools (databases, filesystems, APIs). Giving containers direct host access is a security risk. The MCP Gateway acts as a secure broker:

```
Agent Container  ──SSE──>  MCP Gateway  ──stdio──>  MCP Server  ──>  Database
                           (broker)                 (postgres)
```

- Agents only see the gateway endpoint
- Gateway controls which tools are exposed
- No host filesystem or socket access needed
- Secrets stay in Docker secrets, not env vars

## Quick Start

```bash
docker compose up --build
```

## Key Files

| File | Purpose |
|------|---------|
| `compose.yaml` | Full stack: gateway + agent + database |
| `mcp-config.json` | MCP server registry for the gateway |
| `agent.py` | Example Python agent using gateway tools |

## Adapting for Your Stack

Replace the MCP servers in `mcp-config.json` with your own. The gateway supports any MCP server that uses stdio transport.

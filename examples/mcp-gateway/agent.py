"""
Example agent that connects to MCP Gateway via SSE transport.

The agent discovers available tools from the gateway and can invoke them
without any direct access to the host system or MCP server processes.
"""

import os
import json
import httpx

MCP_SERVER_URL = os.environ.get("MCP_SERVER_URL", "http://mcp-gateway:8811/sse")
AGENT_NAME = os.environ.get("AGENT_NAME", "example-agent")


def discover_tools() -> list[dict]:
    """List all tools available through the MCP Gateway."""
    response = httpx.get(f"{MCP_SERVER_URL.replace('/sse', '')}/tools")
    response.raise_for_status()
    return response.json()


def call_tool(tool_name: str, arguments: dict) -> dict:
    """Invoke a tool through the MCP Gateway."""
    response = httpx.post(
        f"{MCP_SERVER_URL.replace('/sse', '')}/tools/{tool_name}",
        json={"arguments": arguments},
        timeout=30.0,
    )
    response.raise_for_status()
    return response.json()


def main():
    print(f"[{AGENT_NAME}] Starting up...")
    print(f"[{AGENT_NAME}] MCP Gateway: {MCP_SERVER_URL}")

    # Discover available tools
    tools = discover_tools()
    print(f"[{AGENT_NAME}] Available tools: {json.dumps([t['name'] for t in tools], indent=2)}")

    # Example: Query the database via MCP
    result = call_tool("query", {
        "sql": "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'"
    })
    print(f"[{AGENT_NAME}] Database tables: {json.dumps(result, indent=2)}")


if __name__ == "__main__":
    main()

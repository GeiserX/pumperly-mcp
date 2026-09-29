<p align="center">
  <img src="https://raw.githubusercontent.com/GeiserX/pumperly-mcp/main/docs/images/banner.svg" alt="pumperly-mcp banner" width="900"/>
</p>

<h1 align="center">pumperly-mcp</h1>

<p align="center">
  <a href="https://www.npmjs.com/package/pumperly-mcp"><img src="https://img.shields.io/npm/v/pumperly-mcp?style=flat-square&logo=npm" alt="npm"/></a>
  <a href="https://github.com/GeiserX/pumperly-mcp/actions/workflows/ci.yml"><img src="https://github.com/GeiserX/pumperly-mcp/actions/workflows/ci.yml/badge.svg" alt="CI"/></a>
  <a href="https://github.com/GeiserX/pumperly-mcp/blob/main/LICENSE"><img src="https://img.shields.io/github/license/GeiserX/pumperly-mcp?style=flat-square" alt="License"/></a>
  <a href="https://hub.docker.com/r/drumsergio/pumperly-mcp"><img src="https://img.shields.io/docker/pulls/drumsergio/pumperly-mcp?style=flat-square&logo=docker" alt="Docker Pulls"/></a>
  <a href="https://glama.ai/mcp/servers/GeiserX/pumperly-mcp"><img src="https://glama.ai/mcp/servers/GeiserX/pumperly-mcp/badges/score.svg" alt="Glama MCP Server"/></a>
</p>

A small bridge that exposes any [Pumperly](https://github.com/GeiserX/Pumperly) instance as an MCP server, so an AI agent can query fuel prices, find stations, plan routes and geocode places.

## Features

- Five tools: `find_nearest_stations`, `get_stations_in_area`, `calculate_route`, `find_route_stations`, `geocode`.
- Three read-only resources: `pumperly://config`, `pumperly://stats`, `pumperly://exchange-rates`.
- Two transports: stdio, and HTTP on a single `/mcp` endpoint.
- An npm package that downloads the release binary for your platform and runs it over stdio.
- A Docker image for amd64 and arm64.
- Works with the public instance at pumperly.com or your own: set `PUMPERLY_URL`.
- No API key: it reads the Pumperly API anonymously.

## Quick start

Add this to your MCP client's configuration (Claude Desktop, Claude Code, Cursor and others read the `mcpServers` format):

```json
{
  "mcpServers": {
    "pumperly": {
      "command": "npx",
      "args": ["-y", "pumperly-mcp"],
      "env": { "PUMPERLY_URL": "https://pumperly.com" }
    }
  }
}
```

For the HTTP transport (`/mcp` on port 8080) run the Docker image; see [Getting started](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/getting-started.md).

## Documentation

- [Getting started](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/getting-started.md): npm, Docker Compose, building from source, connecting a client
- [Configuration](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/configuration.md): environment variables
- [Usage](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/usage.md): tools, resources, the request flow, testing with the Inspector
- [Related projects](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/related.md): the Pumperly family, other MCP servers, registry listings

## License

[GPL-3.0-or-later](https://github.com/GeiserX/pumperly-mcp/blob/main/LICENSE)

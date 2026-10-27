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

A small bridge that exposes any [Pumperly](https://github.com/GeiserX/Pumperly) instance as an MCP server, so an AI agent can find the cheapest fuel or charging near a place or along a drive, in about 40 countries, mostly in Europe.

## Features

- Six read-only tools:

  | Tool | What it answers |
  |---|---|
  | `cheapest_fuel_on_route` | Where to fill up between two places, in one call |
  | `find_nearest_stations` | Stations near a place name or coordinates, by distance or price |
  | `get_stations_in_area` | Stations inside a bounding box, cheapest first |
  | `calculate_route` | Driving distance and time between places, with up to 5 stops |
  | `find_route_stations` | Stations along a route geometry you already have |
  | `geocode` | Coordinates for a place name |

- Four resources: `pumperly://fuel-types`, `pumperly://config`, `pumperly://stats`, `pumperly://exchange-rates`.
- Short answers: station lists are sorted, cut to a limit and summarised (min, median and max price), so a search costs about 1 KB of context.
- MCP 2026-07-28 and every earlier version back to 2024-11-05, on the same binary: new clients use `server/discover`, older ones `initialize`.
- Two transports: stdio, and streamable HTTP on `/mcp` with a bearer token and origin checks.
- Works with the public instance at pumperly.com or your own: set `PUMPERLY_URL`. No API key.

Every tool only reads, so there is no read-only switch to set.

## Install

The npm package downloads the release binary for your platform, checks its SHA-256 against the checksums shipped in the package, and runs it over stdio. It needs Node.js 18 or later.

Claude Code:

```sh
claude mcp add pumperly -- npx -y pumperly-mcp
```

Claude Desktop (`claude_desktop_config.json`) and Cursor (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "pumperly": { "command": "npx", "args": ["-y", "pumperly-mcp"] }
  }
}
```

VS Code (`.vscode/mcp.json`):

```json
{
  "servers": {
    "pumperly": { "type": "stdio", "command": "npx", "args": ["-y", "pumperly-mcp"] }
  }
}
```

Codex (`~/.codex/config.toml`):

```toml
[mcp_servers.pumperly]
command = "npx"
args = ["-y", "pumperly-mcp"]
```

Docker, over stdio:

```sh
docker run -i --rm -e TRANSPORT=stdio drumsergio/pumperly-mcp:v0.2.0
```

To use your own Pumperly instance, add `PUMPERLY_URL` to the environment (`"env": { "PUMPERLY_URL": "https://..." }` in the JSON configs). For the HTTP transport, see [Getting started](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/getting-started.md).

## Documentation

- [Getting started](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/getting-started.md): npm, Docker Compose, building from source, connecting a client
- [Configuration](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/configuration.md): environment variables
- [Usage](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/usage.md): tools, resources, the request flow, testing with the Inspector
- [Related projects](https://github.com/GeiserX/pumperly-mcp/blob/main/docs/related.md): the Pumperly family, other MCP servers, registry listings

## License

[GPL-3.0-or-later](https://github.com/GeiserX/pumperly-mcp/blob/main/LICENSE)

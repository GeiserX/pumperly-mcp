# Getting started

pumperly-mcp talks to a Pumperly instance over its public API. The default is `https://pumperly.com`; point `PUMPERLY_URL` at your own instance to use that instead. No API key is needed.

## npm (stdio)

```sh
npx -y pumperly-mcp
```

Or install it globally:

```sh
npm install -g pumperly-mcp
pumperly-mcp
```

The package downloads the prebuilt Go binary for your platform from [GitHub Releases](https://github.com/GeiserX/pumperly-mcp/releases), checks its SHA-256 against the `checksums.txt` shipped inside the package, and runs it with the stdio transport. If npm skipped the install script, the first run downloads it. It needs Node.js 18 or later.

Client configuration (the `mcpServers` format that Claude Desktop, Cursor and others read; the [README](https://github.com/GeiserX/pumperly-mcp#install) has Claude Code, VS Code and Codex):

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

## Docker Compose (HTTP)

```yaml
services:
  pumperly-mcp:
    image: drumsergio/pumperly-mcp:v0.1.2
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - TRANSPORT=http
      - PUMPERLY_URL=https://pumperly.com
      - MCP_AUTH_TOKEN=${MCP_AUTH_TOKEN}
```

The image listens on `0.0.0.0:8080` inside the container, so it needs a token: generate one with `openssl rand -hex 32`, put it in `.env` as `MCP_AUTH_TOKEN`, and the server refuses to start without it. The compose file above also publishes the port on host loopback only. The server then answers on `http://localhost:8080/mcp`. A client that supports remote servers connects with:

```json
{
  "mcpServers": {
    "pumperly": {
      "url": "http://localhost:8080/mcp",
      "headers": { "Authorization": "Bearer <your token>" }
    }
  }
}
```

A browser app on another origin also needs its origin in `MCP_ALLOWED_ORIGINS`. To expose the server beyond the host, put it behind a reverse proxy with TLS and keep the token.

The image runs over stdio by default: `docker run -i --rm -e TRANSPORT=stdio drumsergio/pumperly-mcp:v0.1.2` (images before the next release default to HTTP, hence the explicit variable).

## Build from source

```sh
git clone https://github.com/GeiserX/pumperly-mcp
cd pumperly-mcp

# (optional) create .env from the sample
cp .env.example .env && $EDITOR .env

go run ./cmd/server
```

The settings are in [Configuration](configuration.md).

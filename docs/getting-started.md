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

The package downloads the prebuilt Go binary for your platform from [GitHub Releases](https://github.com/GeiserX/pumperly-mcp/releases) and runs it with the stdio transport. It needs Node.js 18 or later.

Client configuration (the `mcpServers` format that Claude Desktop, Claude Code, Cursor and others read):

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
    image: drumsergio/pumperly-mcp:v0.1.0
    ports:
      - "127.0.0.1:8080:8080"
    environment:
      - PUMPERLY_URL=https://pumperly.com
```

The server then answers on `http://localhost:8080/mcp`. A client that supports remote servers connects with:

```json
{
  "mcpServers": {
    "pumperly": { "url": "http://localhost:8080/mcp" }
  }
}
```

> **Security note:** the HTTP transport has no authentication. Outside Docker it listens on `127.0.0.1:8080` by default. The image sets `LISTEN_ADDR=0.0.0.0:8080`, so the container listens on all its interfaces, and the compose file above publishes the port on host loopback only. If you need to expose it on a network, put it behind a reverse proxy with authentication.

To run the image over stdio instead: `docker run -i --rm -e TRANSPORT=stdio drumsergio/pumperly-mcp:v0.1.0`.

## Build from source

```sh
git clone https://github.com/GeiserX/pumperly-mcp
cd pumperly-mcp

# (optional) create .env from the sample
cp .env.example .env && $EDITOR .env

go run ./cmd/server
```

The settings are in [Configuration](configuration.md).

# Configuration

| Variable              | Default                | Description |
|-----------------------|------------------------|-------------|
| `PUMPERLY_URL`        | `https://pumperly.com` | Pumperly instance URL, absolute `http(s)`, no trailing `/`. The server refuses to start on anything else. |
| `TRANSPORT`           | `http` (`stdio` in the Docker image) | `stdio` or `http`. The npm package always sets `stdio`. |
| `LISTEN_ADDR`         | `127.0.0.1:8080`       | HTTP listen address. The Docker image sets `0.0.0.0:8080`. |
| `MCP_AUTH_TOKEN`      | _(empty)_              | When set, every HTTP request needs `Authorization: Bearer <token>`. Required whenever `LISTEN_ADDR` is not a loopback address: the server refuses to start without it. |
| `MCP_ALLOWED_ORIGINS` | _(empty)_              | Comma-separated browser origins (`https://app.example.com`) allowed to call the HTTP endpoint cross-origin. Same-origin and non-browser clients need nothing. |

Set them in the environment, or put them in a `.env` file (copy [`.env.example`](https://github.com/GeiserX/pumperly-mcp/blob/main/.env.example)). The server reads `.env` at start unless `TRANSPORT=stdio` is already set in the environment (the npm package sets it), so a `.env` file can also set `TRANSPORT`.

## HTTP limits

The HTTP transport is stateless: no session id, any replica can answer any request. A request body over 1 MB gets `413`. Header read times out after 10 s, the whole request after 30 s, a response after 2 minutes. Answers from the Pumperly API over 8 MB are refused, and a tool result is capped at 512 KB.

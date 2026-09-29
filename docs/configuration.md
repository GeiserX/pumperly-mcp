# Configuration

| Variable       | Default                  | Description                                      |
|----------------|--------------------------|--------------------------------------------------|
| `PUMPERLY_URL` | `https://pumperly.com`   | Pumperly instance URL (without trailing /)       |
| `LISTEN_ADDR`  | `127.0.0.1:8080`         | HTTP listen address (the Docker image sets `0.0.0.0:8080`) |
| `TRANSPORT`    | _(empty = HTTP)_         | Set to `stdio` for the stdio transport (the npm package sets it for you) |

Set them in the environment, or put them in a `.env` file (copy [`.env.example`](https://github.com/GeiserX/pumperly-mcp/blob/main/.env.example)). The `.env` file is read only with the HTTP transport.

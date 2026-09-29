# Usage

## Tools and resources

| Type          | What for                                                       | MCP URI / Tool id                |
|---------------|----------------------------------------------------------------|----------------------------------|
| **Resources** | Browse configuration, statistics, and exchange rates read-only | `pumperly://config`<br>`pumperly://stats`<br>`pumperly://exchange-rates` |
| **Tools**     | Find stations, calculate routes, and geocode locations          | `find_nearest_stations`<br>`get_stations_in_area`<br>`calculate_route`<br>`find_route_stations`<br>`geocode` |

Over HTTP everything is exposed on a single JSON-RPC endpoint, `/mcp`. A client calls `initialize`, then reuses the returned session id in the `Mcp-Session-Id` header for every other call: `resources/read` for the `pumperly://` URIs, `tools/list` to discover the tools and `tools/call` to run one.

## Testing

The server is tested with the [MCP Inspector](https://modelcontextprotocol.io/docs/tools/inspector). Before opening a pull request, check that your change behaves there too.

## Contributing

[Open an issue](https://github.com/GeiserX/pumperly-mcp/issues/new) or send a pull request. pumperly-mcp follows the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/) Code of Conduct.

Built with [mcp-go](https://github.com/mark3labs/mcp-go) and released with [GoReleaser](https://goreleaser.com/).

# Usage

## Tools and resources

| Tool | Main arguments | Returns |
|---|---|---|
| `cheapest_fuel_on_route` | `origin`, `destination` (place names, `"lat,lon"` text or coordinates), `fuel`, `corridor_km` (default 2), `limit`, `sort_by` `price` or `route` | Route distance and time, plus the stations along it with `route_km` |
| `find_nearest_stations` | `near` (place name) or `lat`/`lon`, `fuel`, `radius_km` (default 10), `limit` (default 5), `sort_by` `distance` or `price` | Stations with distance in km |
| `get_stations_in_area` | `bbox` as `minLon,minLat,maxLon,maxLat`, `fuel`, `limit` (default 20), `sort_by` `price` or `distance` (from the box centre) | Stations, cheapest first |
| `calculate_route` | `origin`, `destination` (or the `_lat`/`_lon` pairs), `waypoints` as `[lon, lat]` (max 5), `include_geometry` | Distance, duration, and the geometry only when asked |
| `find_route_stations` | `geometry` (GeoJSON LineString), `fuel`, `corridor_km` (default 5), `limit`, `sort_by` | Stations along the route |
| `geocode` | `query`, optional `lat`/`lon` to bias, `limit` | Places with coordinates |

Every station list carries the min, median and max price per currency, how many stations matched and how many were returned. Every tool is read-only, returns `structuredContent` that matches its `outputSchema`, and the same JSON as text for older clients. `fuel` is one of the 16 codes in `pumperly://fuel-types`.

| Resource | What it holds |
|---|---|
| `pumperly://fuel-types` | Fuel codes, labels and price units (litre, kg, kWh) |
| `pumperly://config` | Default country and fuel, map centre, enabled countries |
| `pumperly://stats` | Station and price counts per country, with the last update time |
| `pumperly://exchange-rates` | Daily ECB rates against EUR, for comparing prices across currencies |

Over HTTP everything is on one endpoint, `/mcp`. MCP 2026-07-28 clients call `server/discover` and then any method; older clients call `initialize` first. The server keeps no session, so there is no `Mcp-Session-Id` to carry.

## Testing

The server is tested with the [MCP Inspector](https://modelcontextprotocol.io/docs/tools/inspector). Before opening a pull request, check that your change behaves there too.

## Contributing

[Open an issue](https://github.com/GeiserX/pumperly-mcp/issues/new) or send a pull request. pumperly-mcp follows the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/) Code of Conduct.

Built with [mcp-go](https://github.com/mark3labs/mcp-go) v1.4 and released with [GoReleaser](https://goreleaser.com/).

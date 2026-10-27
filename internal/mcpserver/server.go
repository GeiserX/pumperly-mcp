// Package mcpserver builds the Pumperly MCP server: its identity, resources
// and tools. cmd/server only picks the transport.
package mcpserver

import (
	"github.com/geiserx/pumperly-mcp/client"
	"github.com/geiserx/pumperly-mcp/internal/resources"
	"github.com/geiserx/pumperly-mcp/internal/tools"
	"github.com/geiserx/pumperly-mcp/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Instructions is what server/discover and initialize tell the model about
// this server.
const Instructions = `Pumperly fuel and EV charging prices, stations and driving routes in about 40 countries, mostly in Europe (pumperly://stats lists them). Every tool is read-only.

- Cheapest fuel near a place: find_nearest_stations with near="<place>" (or lat/lon) and sort_by=price.
- Cheapest in an area: get_stations_in_area with bbox "minLon,minLat,maxLon,maxLat".
- Where to fill up on a trip: cheapest_fuel_on_route with origin and destination names. One call; do not chain calculate_route and find_route_stations for this.
- fuel is a code: B7 diesel, E5 petrol 95, E10, E5_98 petrol 98, LPG, CNG, EV and more in pumperly://fuel-types.

Prices are per litre (per kg for CNG, LNG and H2, per kWh for EV) in each station's local currency; convert with pumperly://exchange-rates (EUR base) before comparing across borders. Coordinates in GeoJSON, bbox and waypoints are [lon, lat], longitude first; lat and lon parameters are named. Route searches are limited to 30 per minute.`

// New returns a server with every resource and tool registered. It serves
// both protocol eras: legacy clients that call initialize and 2026-07-28
// clients that call server/discover and send per-request _meta.
func New(c *client.Client) *server.MCPServer {
	s := server.NewMCPServer(
		"Pumperly MCP Bridge",
		version.Version,
		server.WithTitle("Pumperly"),
		server.WithDescription("Fuel and EV charging prices, stations and routes from Pumperly"),
		server.WithWebsiteURL("https://github.com/GeiserX/pumperly-mcp"),
		server.WithIcons(mcp.Icon{Src: "https://pumperly.com/icon.svg", MIMEType: "image/svg+xml", Sizes: []string{"any"}}),
		server.WithInstructions(Instructions),
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithRecovery(),
		server.WithInputSchemaValidation(),
		server.WithStrictInputSchemaDefault(),
		server.WithOutputSchemaValidation(),
		// Station results are compact; this only stops a runaway result.
		server.WithToolResultSizeLimit(512<<10),
		// The tool and resource lists only change with a release.
		server.WithCacheHints(60*60*1000, mcp.CacheScopePublic),
		// Resource contents are live data: prices update every few minutes.
		server.WithMethodCacheHints(mcp.MethodResourcesRead, 60*1000, mcp.CacheScopePublic),
	)

	resources.RegisterConfig(s, c)
	resources.RegisterStats(s, c)
	resources.RegisterExchangeRates(s, c)
	resources.RegisterFuelTypes(s)

	for _, newTool := range []func(*client.Client) (mcp.Tool, server.ToolHandlerFunc){
		tools.NewFindNearestStations,
		tools.NewGetStationsInArea,
		tools.NewCheapestFuelOnRoute,
		tools.NewCalculateRoute,
		tools.NewFindRouteStations,
		tools.NewGeocode,
	} {
		tool, handler := newTool(c)
		s.AddTool(tool, handler)
	}
	return s
}

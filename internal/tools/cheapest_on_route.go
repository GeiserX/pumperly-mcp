package tools

import (
	"context"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RouteFuelResult is the structured result of cheapest_fuel_on_route.
type RouteFuelResult struct {
	Origin      Place       `json:"origin"`
	Destination Place       `json:"destination"`
	DistanceKm  float64     `json:"distance_km"`
	DurationMin float64     `json:"duration_min"`
	Stations    StationList `json:"stations"`
}

// NewCheapestFuelOnRoute builds the Tool definition plus its handler.
func NewCheapestFuelOnRoute(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {
	opts := append(readOnlyTool("Cheapest fuel along a trip"), routeEndParams()...)
	opts = append(opts,
		mcp.WithDescription("Plan a drive and find where to fill up: routes from origin to destination, then lists the stations within corridor_km of the route, cheapest first, with each one's km from the start. "+
			"Ends can be place names (\"Madrid\", \"Valencia\"), \"lat,lon\" text or coordinates. "+
			"Prices are per litre (per kg for CNG, LNG and H2, per kWh for EV) in local currency; cross-border trips mix currencies, see pumperly://exchange-rates. "+
			"Uses two of the API's 30 route searches per minute."),
		fuelParam(),
		mcp.WithNumber("corridor_km", mcp.Min(0.5), mcp.Max(50), mcp.DefaultNumber(2.0),
			mcp.Description("How far off the route a station may be, in km, 0.5 to 50. Default 2 (a short detour).")),
		mcp.WithInteger("limit", mcp.Min(1), mcp.Max(50), mcp.DefaultNumber(10),
			mcp.Description("How many stations to return, 1 to 50. Default 10.")),
		mcp.WithString("sort_by", mcp.Enum("price", "route"), mcp.DefaultString("price"),
			mcp.Description("price (cheapest first, default) or route (in driving order).")),
		mcp.WithOutputSchema[RouteFuelResult](),
	)
	tool := mcp.NewTool("cheapest_fuel_on_route", opts...)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		origin, err := routeEnd(ctx, c, req, "origin", "origin_lat", "origin_lon")
		if err != nil {
			return toolError(err)
		}
		dest, err := routeEnd(ctx, c, req, "destination", "dest_lat", "dest_lon")
		if err != nil {
			return toolError(err)
		}
		fuel, err := requireFuel(req)
		if err != nil {
			return toolError(err)
		}
		corridorKm, err := optionalFloat(req, "corridor_km", 2, 0.5, 50)
		if err != nil {
			return toolError(err)
		}
		limit, err := optionalInt(req, "limit", 10, 1, 50)
		if err != nil {
			return toolError(err)
		}
		sortBy, err := optionalEnum(req, "sort_by", "price", "price", "route")
		if err != nil {
			return toolError(err)
		}

		resp, err := c.CalculateRoute(ctx, routeRequest{
			Origin:      [2]float64{origin.Lon, origin.Lat},
			Destination: [2]float64{dest.Lon, dest.Lat},
		})
		if err != nil {
			return toolError(err)
		}
		routes, err := decodeRoutes(resp)
		if err != nil {
			return toolError(err)
		}
		r := routes[0]
		line := LineString{Type: "LineString", Coordinates: downsample(r.Geometry.Coordinates, maxRoutePoints)}
		st, err := stationsAlongRoute(ctx, c, line, fuel, corridorKm, r.Distance)
		if err != nil {
			return toolError(err)
		}
		return structured(RouteFuelResult{
			Origin:      origin,
			Destination: dest,
			DistanceKm:  round(r.Distance, 1),
			DurationMin: round(r.Duration/60, 0),
			Stations:    newStationList(st, fuel, sortBy, limit),
		})
	}

	return tool, handler
}

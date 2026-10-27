package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// maxRoutePoints is the API's limit (route-stations.ts MAX_COORDINATES).
const maxRoutePoints = 2000

// NewFindRouteStations builds the Tool definition plus its handler.
func NewFindRouteStations(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {
	opts := append(readOnlyTool("Fuel or EV stations along a route line"),
		mcp.WithDescription("List the fuel stations or EV chargers within corridor_km of a route line you already have, cheapest first by default, with each station's km along the route. "+
			"Most callers want cheapest_fuel_on_route, which takes place names and does the routing too. "+
			"Prices are per litre (per kg for CNG, LNG and H2, per kWh for EV) in local currency. The API allows 30 route searches per minute."),
		mcp.WithObject("geometry",
			mcp.Required(),
			mcp.Properties(map[string]any{
				"type": map[string]any{"type": "string", "enum": []string{"LineString"}},
				"coordinates": map[string]any{
					"type":     "array",
					"minItems": 2,
					"maxItems": maxRoutePoints,
					"items": map[string]any{
						"type":     "array",
						"items":    map[string]any{"type": "number"},
						"minItems": 2,
						"maxItems": 2,
					},
				},
			}),
			mcp.Description("GeoJSON LineString of the route: {\"type\":\"LineString\",\"coordinates\":[[lon,lat],...]}, longitude first, 2 to 2000 points. calculate_route returns one with include_geometry=true."),
		),
		fuelParam(),
		mcp.WithNumber("corridor_km", mcp.Min(0.5), mcp.Max(50), mcp.DefaultNumber(5.0),
			mcp.Description("How far from the route a station may be, in km, 0.5 to 50. Default 5.")),
		mcp.WithInteger("limit", mcp.Min(1), mcp.Max(50), mcp.DefaultNumber(10),
			mcp.Description("How many stations to return, 1 to 50. Default 10.")),
		mcp.WithString("sort_by", mcp.Enum("price", "route"), mcp.DefaultString("price"),
			mcp.Description("price (cheapest first, default) or route (in driving order from the start).")),
		mcp.WithOutputSchema[StationList](),
	)
	tool := mcp.NewTool("find_route_stations", opts...)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		line, err := parseLineString(req.GetArguments()["geometry"])
		if err != nil {
			return toolError(err)
		}
		fuel, err := requireFuel(req)
		if err != nil {
			return toolError(err)
		}
		corridorKm, err := optionalFloat(req, "corridor_km", 5, 0.5, 50)
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
		st, err := stationsAlongRoute(ctx, c, line, fuel, corridorKm, lineLengthKm(line))
		if err != nil {
			return toolError(err)
		}
		return structured(newStationList(st, fuel, sortBy, limit))
	}

	return tool, handler
}

type routeStationsRequest struct {
	Geometry   LineString `json:"geometry"`
	Fuel       string     `json:"fuel"`
	CorridorKm float64    `json:"corridorKm"`
}

func stationsAlongRoute(ctx context.Context, c *client.Client, line LineString, fuel string, corridorKm, routeKm float64) ([]Station, error) {
	body, err := c.FindRouteStations(ctx, routeStationsRequest{Geometry: line, Fuel: fuel, CorridorKm: corridorKm})
	if err != nil {
		return nil, err
	}
	return decodeStations(body, fuel, routeKm)
}

// parseLineString reads a GeoJSON LineString object (or the same object as a
// JSON string) and checks it against the API's limits.
func parseLineString(v any) (LineString, error) {
	var line LineString
	if v == nil {
		return line, fmt.Errorf("missing required parameter: geometry")
	}
	raw, ok := v.(string)
	if !ok {
		b, err := json.Marshal(v)
		if err != nil {
			return line, fmt.Errorf("geometry must be a GeoJSON LineString object")
		}
		raw = string(b)
	}
	if err := json.Unmarshal([]byte(raw), &line); err != nil {
		return line, fmt.Errorf("geometry must be a GeoJSON LineString object: {\"type\":\"LineString\",\"coordinates\":[[lon,lat],...]}")
	}
	if line.Type != "LineString" {
		return line, fmt.Errorf("geometry type must be LineString, got %q", line.Type)
	}
	if n := len(line.Coordinates); n < 2 || n > maxRoutePoints {
		return line, fmt.Errorf("geometry needs 2 to %d points, got %d", maxRoutePoints, n)
	}
	for i, p := range line.Coordinates {
		if len(p) != 2 {
			return line, fmt.Errorf("geometry point %d must be [lon, lat]", i+1)
		}
		if err := checkLatLon(p[1], p[0]); err != nil {
			return line, fmt.Errorf("geometry point %d: %v (points are [lon, lat], longitude first)", i+1, err)
		}
	}
	return line, nil
}

func lineLengthKm(line LineString) float64 {
	total := 0.0
	for i := 1; i < len(line.Coordinates); i++ {
		a, b := line.Coordinates[i-1], line.Coordinates[i]
		total += haversineKm(a[1], a[0], b[1], b[0])
	}
	return total
}

// downsample keeps at most max points, evenly spaced, always keeping the
// first and the last, so a long route still fits the API's point limit.
func downsample(pts [][]float64, max int) [][]float64 {
	if len(pts) <= max || max < 2 {
		return pts
	}
	out := make([][]float64, 0, max)
	step := float64(len(pts)-1) / float64(max-1)
	for i := 0; i < max; i++ {
		out = append(out, pts[int(float64(i)*step+0.5)])
	}
	return out
}

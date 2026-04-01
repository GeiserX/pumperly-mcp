package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewFindRouteStations builds the Tool definition plus its handler.
func NewFindRouteStations(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("find_route_stations",
		mcp.WithDescription("Find fuel stations along a route corridor"),
		mcp.WithString("geometry",
			mcp.Required(),
			mcp.Description("GeoJSON LineString as a JSON string, e.g. {\"type\":\"LineString\",\"coordinates\":[[lon,lat],...]}"),
		),
		mcp.WithString("fuel",
			mcp.Required(),
			mcp.Description("Fuel type code (e.g. B7, E5, E10, E85, LPG, CNG, LNG, EV)"),
		),
		mcp.WithNumber("corridor_km",
			mcp.Description("Corridor width in kilometres (default 5)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		geometryStr, err := req.RequireString("geometry")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fuel, err := req.RequireString("fuel")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		corridorKm := optionalFloat(req, "corridor_km", 5)

		// Parse the geometry JSON string into a map
		var geometry map[string]any
		if err := json.Unmarshal([]byte(geometryStr), &geometry); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid geometry JSON: %v", err)), nil
		}

		body := map[string]any{
			"geometry":   geometry,
			"fuel":       fuel,
			"corridorKm": corridorKm,
		}

		resp, err := c.FindRouteStations(body)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%s", string(resp))), nil
	}

	return tool, handler
}

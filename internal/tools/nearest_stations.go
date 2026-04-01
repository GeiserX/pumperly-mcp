package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewFindNearestStations builds the Tool definition plus its handler.
func NewFindNearestStations(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("find_nearest_stations",
		mcp.WithDescription("Find fuel stations nearest to a coordinate"),
		mcp.WithNumber("lat",
			mcp.Required(),
			mcp.Description("Latitude of the search centre"),
		),
		mcp.WithNumber("lon",
			mcp.Required(),
			mcp.Description("Longitude of the search centre"),
		),
		mcp.WithString("fuel",
			mcp.Required(),
			mcp.Description("Fuel type code (e.g. B7, E5, E10, E85, LPG, CNG, LNG, EV)"),
		),
		mcp.WithNumber("radius_km",
			mcp.Description("Search radius in kilometres (default 10)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (default 5)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lat, err := requiredFloat(req, "lat")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		lon, err := requiredFloat(req, "lon")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fuel, err := req.RequireString("fuel")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		radiusKm := optionalFloat(req, "radius_km", 10)
		limit := optionalFloat(req, "limit", 5)

		resp, err := c.FindNearestStations(ctx, lat, lon, fuel, radiusKm, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%s", string(resp))), nil
	}

	return tool, handler
}

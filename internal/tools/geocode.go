package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewGeocode builds the Tool definition plus its handler.
func NewGeocode(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("geocode",
		mcp.WithDescription("Search for a location by name (geocoding)"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query (e.g. city name, address, POI)"),
		),
		mcp.WithNumber("lat",
			mcp.Description("Optional latitude for geographic biasing"),
		),
		mcp.WithNumber("lon",
			mcp.Description("Optional longitude for geographic biasing"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var lat, lon *float64
		if v, ok := req.GetArguments()["lat"].(float64); ok {
			lat = &v
		}
		if v, ok := req.GetArguments()["lon"].(float64); ok {
			lon = &v
		}

		resp, err := c.Geocode(ctx, query, lat, lon)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%s", string(resp))), nil
	}

	return tool, handler
}

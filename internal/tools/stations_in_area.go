package tools

import (
	"context"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewGetStationsInArea builds the Tool definition plus its handler.
func NewGetStationsInArea(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("get_stations_in_area",
		mcp.WithDescription("Get fuel stations within a bounding box"),
		mcp.WithString("bbox",
			mcp.Required(),
			mcp.Description("Bounding box as \"minLon,minLat,maxLon,maxLat\""),
		),
		mcp.WithString("fuel",
			mcp.Required(),
			mcp.Description("Fuel type code (e.g. B7, E5, E10, E85, LPG, CNG, LNG, EV)"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		bbox, err := req.RequireString("bbox")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		fuel, err := req.RequireString("fuel")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := c.GetStationsInArea(ctx, bbox, fuel)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%s", string(resp))), nil
	}

	return tool, handler
}

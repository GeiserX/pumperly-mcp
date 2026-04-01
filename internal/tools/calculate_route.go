package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewCalculateRoute builds the Tool definition plus its handler.
func NewCalculateRoute(c *client.Client) (mcp.Tool, server.ToolHandlerFunc) {

	tool := mcp.NewTool("calculate_route",
		mcp.WithDescription("Calculate a driving route between two points"),
		mcp.WithNumber("origin_lon",
			mcp.Required(),
			mcp.Description("Origin longitude"),
		),
		mcp.WithNumber("origin_lat",
			mcp.Required(),
			mcp.Description("Origin latitude"),
		),
		mcp.WithNumber("dest_lon",
			mcp.Required(),
			mcp.Description("Destination longitude"),
		),
		mcp.WithNumber("dest_lat",
			mcp.Required(),
			mcp.Description("Destination latitude"),
		),
		mcp.WithString("waypoints",
			mcp.Description("Optional JSON array of waypoints as [[lon,lat],...]"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		originLon, err := requiredFloat(req, "origin_lon")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		originLat, err := requiredFloat(req, "origin_lat")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		destLon, err := requiredFloat(req, "dest_lon")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		destLat, err := requiredFloat(req, "dest_lat")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		body := map[string]any{
			"origin":      []float64{originLon, originLat},
			"destination": []float64{destLon, destLat},
		}

		// Parse optional waypoints JSON string
		if wp, ok := req.GetArguments()["waypoints"].(string); ok && wp != "" {
			var waypoints [][]float64
			if err := json.Unmarshal([]byte(wp), &waypoints); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid waypoints JSON: %v", err)), nil
			}
			body["waypoints"] = waypoints
		}

		resp, err := c.CalculateRoute(body)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%s", string(resp))), nil
	}

	return tool, handler
}

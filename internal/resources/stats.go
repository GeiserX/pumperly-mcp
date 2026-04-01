package resources

import (
	"context"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterStats wires pumperly://stats into the server.
func RegisterStats(s *server.MCPServer, c *client.Client) {
	res := mcp.NewResource(
		"pumperly://stats",
		"Platform statistics",
		mcp.WithResourceDescription("Station and price counts per country"),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(res, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := c.GetStats(ctx)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "pumperly://stats",
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}

package resources

import (
	"context"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterConfig wires pumperly://config into the server.
func RegisterConfig(s *server.MCPServer, c *client.Client) {
	res := mcp.NewResource(
		"pumperly://config",
		"App configuration",
		mcp.WithResourceDescription("Countries, defaults, and fuel types supported by Pumperly"),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(res, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := c.GetConfig(ctx)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "pumperly://config",
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}

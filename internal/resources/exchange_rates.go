package resources

import (
	"context"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterExchangeRates wires pumperly://exchange-rates into the server.
func RegisterExchangeRates(s *server.MCPServer, c *client.Client) {
	res := mcp.NewResource(
		"pumperly://exchange-rates",
		"ECB exchange rates",
		mcp.WithResourceDescription("Daily exchange rates from the European Central Bank"),
		mcp.WithMIMEType("application/json"),
	)

	s.AddResource(res, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		body, err := c.GetExchangeRates(ctx)
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "pumperly://exchange-rates",
				MIMEType: "application/json",
				Text:     string(body),
			},
		}, nil
	})
}

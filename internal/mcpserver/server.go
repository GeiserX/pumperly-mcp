// Package mcpserver builds the Pumperly MCP server: its identity, resources
// and tools. cmd/server only picks the transport.
package mcpserver

import (
	"github.com/geiserx/pumperly-mcp/client"
	"github.com/geiserx/pumperly-mcp/internal/resources"
	"github.com/geiserx/pumperly-mcp/internal/tools"
	"github.com/geiserx/pumperly-mcp/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// New returns a server with every resource and tool registered. It serves
// both protocol eras: legacy clients that call initialize and 2026-07-28
// clients that call server/discover and send per-request _meta.
func New(c *client.Client) *server.MCPServer {
	s := server.NewMCPServer(
		"Pumperly MCP Bridge",
		version.Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	resources.RegisterConfig(s, c)
	resources.RegisterStats(s, c)
	resources.RegisterExchangeRates(s, c)

	for _, newTool := range []func(*client.Client) (mcp.Tool, server.ToolHandlerFunc){
		tools.NewFindNearestStations,
		tools.NewGetStationsInArea,
		tools.NewCalculateRoute,
		tools.NewFindRouteStations,
		tools.NewGeocode,
	} {
		tool, handler := newTool(c)
		s.AddTool(tool, handler)
	}
	return s
}

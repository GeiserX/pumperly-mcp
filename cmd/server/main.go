package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/geiserx/pumperly-mcp/config"
	"github.com/geiserx/pumperly-mcp/internal/resources"
	"github.com/geiserx/pumperly-mcp/internal/tools"
	"github.com/geiserx/pumperly-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	log.Printf("Pumperly MCP %s starting...", version.String())

	// Load config & initialise Pumperly client
	cfg := config.LoadPumperlyConfig()
	c := client.New(cfg.BaseURL)

	// Create MCP server
	s := server.NewMCPServer(
		"Pumperly MCP Bridge",
		"0.0.1",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// ── Resources ────────────────────────────────────────────────────────
	resources.RegisterConfig(s, c)
	resources.RegisterStats(s, c)
	resources.RegisterExchangeRates(s, c)

	// ── Tools ────────────────────────────────────────────────────────────

	// TOOL: find_nearest_stations
	tool, handler := tools.NewFindNearestStations(c)
	s.AddTool(tool, handler)

	// TOOL: get_stations_in_area
	tool, handler = tools.NewGetStationsInArea(c)
	s.AddTool(tool, handler)

	// TOOL: calculate_route
	tool, handler = tools.NewCalculateRoute(c)
	s.AddTool(tool, handler)

	// TOOL: find_route_stations
	tool, handler = tools.NewFindRouteStations(c)
	s.AddTool(tool, handler)

	// TOOL: geocode
	tool, handler = tools.NewGeocode(c)
	s.AddTool(tool, handler)

	// ── Transport ────────────────────────────────────────────────────────
	transport := strings.ToLower(os.Getenv("TRANSPORT"))
	if transport == "stdio" {
		stdioSrv := server.NewStdioServer(s)
		log.Println("Pumperly MCP bridge running on stdio")
		if err := stdioSrv.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
	} else {
		httpSrv := server.NewStreamableHTTPServer(s)
		log.Println("Pumperly MCP bridge listening on :8080")
		if err := httpSrv.Start(":8080"); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}

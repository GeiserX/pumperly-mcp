package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/geiserx/pumperly-mcp/config"
	"github.com/geiserx/pumperly-mcp/internal/mcpserver"
	"github.com/geiserx/pumperly-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	log.Printf("Pumperly MCP %s starting...", version.String())

	// Load config & initialise Pumperly client
	cfg := config.LoadPumperlyConfig()
	c := client.New(cfg.BaseURL)

	s := mcpserver.New(c)

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
		addr := os.Getenv("LISTEN_ADDR")
		if addr == "" {
			addr = "127.0.0.1:8080"
		}
		log.Printf("Pumperly MCP bridge listening on %s", addr)
		if err := httpSrv.Start(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/geiserx/pumperly-mcp/config"
	"github.com/geiserx/pumperly-mcp/internal/mcpserver"
	"github.com/geiserx/pumperly-mcp/version"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Logs go to stderr; on stdio, stdout carries only MCP messages.
	log.Printf("Pumperly MCP %s starting...", version.String())

	cfg := config.LoadPumperlyConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	srvCfg := config.LoadServerConfig()
	s := mcpserver.New(client.New(cfg.BaseURL))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if srvCfg.Transport == "stdio" {
		log.Println("Pumperly MCP bridge running on stdio")
		if err := server.NewStdioServer(s).Listen(ctx, os.Stdin, os.Stdout); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("stdio server error: %v", err)
		}
		return
	}

	httpSrv, err := mcpserver.NewHTTPServer(s, mcpserver.HTTPOptions{
		Addr:           srvCfg.ListenAddr,
		AuthToken:      srvCfg.AuthToken,
		AllowedOrigins: srvCfg.AllowedOrigins,
	})
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()
	auth := "no auth (loopback only)"
	if srvCfg.AuthToken != "" {
		auth = "bearer token required"
	}
	log.Printf("Pumperly MCP bridge listening on http://%s/mcp, %s", srvCfg.ListenAddr, auth)
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	if strings.ToLower(os.Getenv("TRANSPORT")) != "stdio" {
		_ = godotenv.Load()
	}
}

type PumperlyConfig struct {
	BaseURL string
}

func LoadPumperlyConfig() PumperlyConfig {
	return PumperlyConfig{
		BaseURL: getEnv("PUMPERLY_URL", "https://pumperly.com"),
	}
}

// Validate rejects a PUMPERLY_URL that is not an absolute http(s) URL, so a
// typo fails at start instead of on the first tool call.
func (c PumperlyConfig) Validate() error {
	u, err := url.Parse(c.BaseURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return fmt.Errorf("PUMPERLY_URL must be an absolute http(s) URL such as https://pumperly.com")
	}
	return nil
}

// ServerConfig holds the transport settings.
type ServerConfig struct {
	// Transport is "stdio" or "http" (the default when TRANSPORT is unset).
	Transport string
	// ListenAddr is the HTTP listen address.
	ListenAddr string
	// AuthToken, when set, is required as "Authorization: Bearer <token>"
	// on every HTTP request. It is mandatory on a non-loopback ListenAddr.
	AuthToken string
	// AllowedOrigins are extra browser origins (scheme://host[:port]) that may
	// call the HTTP endpoint cross-origin.
	AllowedOrigins []string
}

func LoadServerConfig() ServerConfig {
	transport := "http"
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TRANSPORT")), "stdio") {
		transport = "stdio"
	}
	var origins []string
	for _, o := range strings.Split(os.Getenv("MCP_ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return ServerConfig{
		Transport:      transport,
		ListenAddr:     getEnv("LISTEN_ADDR", "127.0.0.1:8080"),
		AuthToken:      strings.TrimSpace(os.Getenv("MCP_AUTH_TOKEN")),
		AllowedOrigins: origins,
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

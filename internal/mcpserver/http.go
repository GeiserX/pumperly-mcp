package mcpserver

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// MaxRequestBytes caps an HTTP request body. The largest legitimate request,
// find_route_stations with a 2000-point LineString, is well under 100 KB.
const MaxRequestBytes = 1 << 20

// HTTPOptions configures the Streamable HTTP transport.
type HTTPOptions struct {
	Addr string
	// AuthToken, when set, must arrive as "Authorization: Bearer <token>".
	AuthToken string
	// AllowedOrigins are trusted browser origins for cross-origin requests.
	AllowedOrigins []string
}

// NewHTTPServer returns an http.Server serving MCP on /mcp. The SDK handler
// runs stateless and sits behind, in order: a bearer-token check, a request
// body cap and cross-origin protection. The SDK itself adds a Host check
// against DNS rebinding on loopback connections.
//
// A listen address that is not loopback needs a token: binding 0.0.0.0
// without one would let anyone who can reach the port use the server.
func NewHTTPServer(s *server.MCPServer, opts HTTPOptions) (*http.Server, error) {
	if opts.AuthToken == "" && !IsLoopback(opts.Addr) {
		return nil, fmt.Errorf("LISTEN_ADDR %q is not a loopback address: set MCP_AUTH_TOKEN so clients must authenticate, or listen on 127.0.0.1", opts.Addr)
	}
	h, err := NewHTTPHandler(s, opts)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:              opts.Addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		// A tool call can chain several upstream requests of up to 30 s each.
		WriteTimeout:   2 * time.Minute,
		IdleTimeout:    2 * time.Minute,
		MaxHeaderBytes: 64 << 10,
	}, nil
}

// NewHTTPHandler builds the /mcp handler chain without listening.
func NewHTTPHandler(s *server.MCPServer, opts HTTPOptions) (http.Handler, error) {
	cop := http.NewCrossOriginProtection()
	for _, o := range opts.AllowedOrigins {
		if err := cop.AddTrustedOrigin(o); err != nil {
			return nil, fmt.Errorf("MCP_ALLOWED_ORIGINS: %w", err)
		}
	}
	mcpHandler := server.NewStreamableHTTPServer(s, server.WithStateLess(true))

	var h http.Handler = cop.Handler(mcpHandler)
	h = limitBody(h, MaxRequestBytes)
	if opts.AuthToken != "" {
		h = requireBearer(h, opts.AuthToken)
	}
	mux := http.NewServeMux()
	mux.Handle("/mcp", h)
	return mux, nil
}

// IsLoopback reports whether a listen address only accepts local
// connections. An empty host (":8080") listens on every interface.
func IsLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func requireBearer(next http.Handler, token string) http.Handler {
	want := sha256.Sum256([]byte(token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, got, ok := strings.Cut(r.Header.Get("Authorization"), " ")
		if ok && strings.EqualFold(scheme, "Bearer") {
			sum := sha256.Sum256([]byte(strings.TrimSpace(got)))
			if subtle.ConstantTimeCompare(sum[:], want[:]) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Bearer realm="pumperly-mcp"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

// limitBody reads at most max bytes of the body and answers 413 past that.
// The SDK reads the whole body with io.ReadAll and has no limit of its own.
func limitBody(next http.Handler, max int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > max {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		if r.Body != nil && r.Body != http.NoBody {
			b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, max))
			if err != nil {
				var tooBig *http.MaxBytesError
				if errors.As(err, &tooBig) {
					http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
					return
				}
				http.Error(w, "could not read request body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(b))
		}
		next.ServeHTTP(w, r)
	})
}

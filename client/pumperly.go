// Package client talks to the Pumperly HTTP API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/geiserx/pumperly-mcp/version"
)

const (
	// MaxResponseBytes caps a successful upstream body. A bounding box over a
	// whole country can return several MB of GeoJSON; anything past this is
	// refused rather than held in memory.
	MaxResponseBytes = 8 << 20
	// maxErrorBodyBytes caps how much of an error body is read to build the
	// message. The body itself is never relayed.
	maxErrorBodyBytes = 4 << 10
	maxDetailLen      = 200
)

// ErrResponseTooLarge is returned when the API answers with more than
// MaxResponseBytes.
var ErrResponseTooLarge = fmt.Errorf("the Pumperly API answer is larger than %d MB; narrow the search (a smaller bbox, radius or route)", MaxResponseBytes>>20)

type Client struct {
	base string
	hc   *http.Client
	ua   string
}

func New(base string) *Client {
	return &Client{
		base: strings.TrimRight(base, "/"),
		hc:   &http.Client{Timeout: 30 * time.Second},
		ua:   "pumperly-mcp/" + version.Version + " (+https://github.com/GeiserX/pumperly-mcp)",
	}
}

// APIError is an HTTP error from the Pumperly API, reduced to a short message
// that is safe to show to a model: it never carries the raw body or the URL.
type APIError struct {
	Status     int
	Message    string
	RetryAfter int // seconds, set on 429 when the API sends Retry-After
}

func (e *APIError) Error() string {
	if e.Status == http.StatusTooManyRequests {
		if e.RetryAfter > 0 {
			return fmt.Sprintf("Pumperly API rate limit reached (HTTP 429); retry after %d s", e.RetryAfter)
		}
		return "Pumperly API rate limit reached (HTTP 429); retry in a minute"
	}
	if e.Message == "" {
		return fmt.Sprintf("Pumperly API error (HTTP %d %s)", e.Status, http.StatusText(e.Status))
	}
	return fmt.Sprintf("Pumperly API error (HTTP %d): %s", e.Status, e.Message)
}

func (c *Client) buildURL(path string, q url.Values) string {
	u := c.base + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

func (c *Client) doGet(ctx context.Context, path string, q url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path, q), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	return c.do(req)
}

func (c *Client) doPost(ctx context.Context, path string, body any) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(path, nil), bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, transportError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes))
		return nil, newAPIError(resp, b)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBytes+1))
	if err != nil {
		return nil, transportError(err)
	}
	if len(b) > MaxResponseBytes {
		return nil, ErrResponseTooLarge
	}
	return b, nil
}

// transportError keeps the full cause in the server log (stderr) and returns
// a short message without the request URL or address, so the configured
// instance URL never reaches the model.
func transportError(err error) error {
	log.Printf("pumperly api request failed: %v", err)
	if errors.Is(err, context.Canceled) {
		return errors.New("the request was cancelled")
	}
	var ue *url.Error
	if errors.As(err, &ue) && ue.Timeout() {
		return errors.New("the Pumperly API did not answer in time; try again")
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return errors.New("could not reach the Pumperly API (DNS lookup failed)")
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Err != nil {
		var sysErr *os.SyscallError
		if errors.As(opErr.Err, &sysErr) {
			return fmt.Errorf("could not reach the Pumperly API (%v)", sysErr.Err)
		}
	}
	return errors.New("could not reach the Pumperly API")
}

// newAPIError turns Pumperly's {"error": "...", "details": [zod issues]}
// shape into one short line. Anything else becomes the status alone.
func newAPIError(resp *http.Response, body []byte) *APIError {
	e := &APIError{Status: resp.StatusCode}
	if resp.StatusCode == http.StatusTooManyRequests {
		if s, err := strconv.Atoi(strings.TrimSpace(resp.Header.Get("Retry-After"))); err == nil && s > 0 {
			e.RetryAfter = s
		}
		return e
	}
	var parsed struct {
		Error   string `json:"error"`
		Details []struct {
			Path    []any  `json:"path"`
			Message string `json:"message"`
		} `json:"details"`
	}
	if json.Unmarshal(body, &parsed) != nil || parsed.Error == "" {
		return e
	}
	msg := clean(parsed.Error)
	var parts []string
	for i, d := range parsed.Details {
		if i == 3 {
			parts = append(parts, fmt.Sprintf("and %d more", len(parsed.Details)-3))
			break
		}
		var path []string
		for _, p := range d.Path {
			path = append(path, fmt.Sprint(p))
		}
		part := clean(d.Message)
		if len(path) > 0 {
			part = clean(strings.Join(path, ".")) + ": " + part
		}
		parts = append(parts, part)
	}
	if len(parts) > 0 {
		msg += " (" + strings.Join(parts, "; ") + ")"
	}
	e.Message = msg
	return e
}

// clean strips control characters and caps the length of upstream text.
func clean(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
	if r := []rune(s); len(r) > maxDetailLen {
		s = string(r[:maxDetailLen]) + "..."
	}
	return strings.TrimSpace(s)
}

// ── Resources ────────────────────────────────────────────────────────────────

// GetConfig returns the instance configuration: default country and fuel,
// map centre and the enabled countries.
func (c *Client) GetConfig(ctx context.Context) ([]byte, error) {
	return c.doGet(ctx, "/api/config", nil)
}

// GetStats returns platform statistics (station/price counts per country).
func (c *Client) GetStats(ctx context.Context) ([]byte, error) {
	return c.doGet(ctx, "/api/stats", nil)
}

// GetExchangeRates returns ECB daily exchange rates.
func (c *Client) GetExchangeRates(ctx context.Context) ([]byte, error) {
	return c.doGet(ctx, "/api/exchange-rates", nil)
}

// ── Tools ────────────────────────────────────────────────────────────────────

// FindNearestStations queries for fuel stations near a coordinate, nearest
// first.
func (c *Client) FindNearestStations(ctx context.Context, lat, lon float64, fuel string, radiusKm float64, limit int) ([]byte, error) {
	q := url.Values{}
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	q.Set("fuel", fuel)
	q.Set("radius_km", strconv.FormatFloat(radiusKm, 'f', -1, 64))
	q.Set("limit", strconv.Itoa(limit))
	return c.doGet(ctx, "/api/stations/nearest", q)
}

// GetStationsInArea queries for fuel stations within a bounding box given as
// "minLon,minLat,maxLon,maxLat".
func (c *Client) GetStationsInArea(ctx context.Context, bbox, fuel string) ([]byte, error) {
	q := url.Values{}
	q.Set("bbox", bbox)
	q.Set("fuel", fuel)
	return c.doGet(ctx, "/api/stations", q)
}

// CalculateRoute requests a driving route; body is the /api/route request.
func (c *Client) CalculateRoute(ctx context.Context, body any) ([]byte, error) {
	return c.doPost(ctx, "/api/route", body)
}

// FindRouteStations finds fuel stations along a route corridor; body is the
// /api/route-stations request.
func (c *Client) FindRouteStations(ctx context.Context, body any) ([]byte, error) {
	return c.doPost(ctx, "/api/route-stations", body)
}

// Geocode searches for a location by name, optionally biased towards a point.
func (c *Client) Geocode(ctx context.Context, query string, lat, lon *float64) ([]byte, error) {
	q := url.Values{}
	q.Set("q", query)
	if lat != nil {
		q.Set("lat", strconv.FormatFloat(*lat, 'f', -1, 64))
	}
	if lon != nil {
		q.Set("lon", strconv.FormatFloat(*lon, 'f', -1, 64))
	}
	return c.doGet(ctx, "/api/geocode", q)
}

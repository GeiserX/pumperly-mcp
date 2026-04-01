package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base string
	hc   *http.Client
}

func New(base string) *Client {
	return &Client{
		base: strings.TrimRight(base, "/"),
		hc:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) buildURL(path string, q url.Values) string {
	u := c.base + path
	if q != nil && len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

func (c *Client) doGet(path string, q url.Values) ([]byte, error) {
	endpoint := c.buildURL(path, q)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) doPost(path string, body any) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	endpoint := c.buildURL(path, nil)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Pumperly API error %d: %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(resp.Body)
}

// ── Resources ────────────────────────────────────────────────────────────────

// GetConfig returns the app configuration (countries, defaults, fuel types).
func (c *Client) GetConfig() ([]byte, error) {
	return c.doGet("/api/config", nil)
}

// GetStats returns platform statistics (station/price counts per country).
func (c *Client) GetStats() ([]byte, error) {
	return c.doGet("/api/stats", nil)
}

// GetExchangeRates returns ECB daily exchange rates.
func (c *Client) GetExchangeRates() ([]byte, error) {
	return c.doGet("/api/exchange-rates", nil)
}

// ── Tools ────────────────────────────────────────────────────────────────────

// FindNearestStations queries for fuel stations near a coordinate.
func (c *Client) FindNearestStations(lat, lon float64, fuel string, radiusKm, limit float64) ([]byte, error) {
	q := url.Values{}
	q.Set("lat", fmt.Sprintf("%f", lat))
	q.Set("lon", fmt.Sprintf("%f", lon))
	q.Set("fuel", fuel)
	q.Set("radius_km", fmt.Sprintf("%g", radiusKm))
	q.Set("limit", fmt.Sprintf("%g", limit))
	return c.doGet("/api/stations/nearest", q)
}

// GetStationsInArea queries for fuel stations within a bounding box.
func (c *Client) GetStationsInArea(bbox, fuel string) ([]byte, error) {
	q := url.Values{}
	q.Set("bbox", bbox)
	q.Set("fuel", fuel)
	return c.doGet("/api/stations", q)
}

// CalculateRoute requests a driving route between origin and destination.
func (c *Client) CalculateRoute(body map[string]any) ([]byte, error) {
	return c.doPost("/api/route", body)
}

// FindRouteStations finds fuel stations along a route corridor.
func (c *Client) FindRouteStations(body map[string]any) ([]byte, error) {
	return c.doPost("/api/route-stations", body)
}

// Geocode searches for a location by name.
func (c *Client) Geocode(query string, lat, lon *float64) ([]byte, error) {
	q := url.Values{}
	q.Set("q", query)
	if lat != nil {
		q.Set("lat", fmt.Sprintf("%f", *lat))
	}
	if lon != nil {
		q.Set("lon", fmt.Sprintf("%f", *lon))
	}
	return c.doGet("/api/geocode", q)
}

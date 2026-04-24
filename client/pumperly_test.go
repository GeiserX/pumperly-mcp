package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ── buildURL ────────────────────────────────────────────────────────────────

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		path   string
		query  url.Values
		expect string
	}{
		{
			name:   "simple path no query",
			base:   "https://pumperly.com",
			path:   "/api/config",
			query:  nil,
			expect: "https://pumperly.com/api/config",
		},
		{
			name:   "path with query params",
			base:   "https://pumperly.com",
			path:   "/api/stations/nearest",
			query:  url.Values{"lat": {"40.0"}, "lon": {"-3.5"}, "fuel": {"gasoline_95"}},
			expect: "https://pumperly.com/api/stations/nearest?fuel=gasoline_95&lat=40.0&lon=-3.5",
		},
		{
			name:   "base with trailing slash",
			base:   "http://localhost:8080/",
			path:   "/api/stats",
			query:  nil,
			expect: "http://localhost:8080/api/stats",
		},
		{
			name:   "empty query values",
			base:   "https://pumperly.com",
			path:   "/api/geocode",
			query:  url.Values{},
			expect: "https://pumperly.com/api/geocode",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tc.base)
			got := c.buildURL(tc.path, tc.query)
			if got != tc.expect {
				t.Errorf("got %q, want %q", got, tc.expect)
			}
		})
	}
}

// ── New ──────────────────────────────────────────────────────────────────────

func TestNew_trims_trailing_slash(t *testing.T) {
	c := New("https://example.com/")
	if c.base != "https://example.com" {
		t.Errorf("base = %q, want trailing slash stripped", c.base)
	}
}

func TestNew_preserves_base_without_slash(t *testing.T) {
	c := New("https://example.com")
	if c.base != "https://example.com" {
		t.Errorf("base = %q, want unchanged", c.base)
	}
}

func TestNew_sets_http_client(t *testing.T) {
	c := New("https://example.com")
	if c.hc == nil {
		t.Fatal("http client must not be nil")
	}
}

// ── do / doGet / doPost (via httptest) ───────────────────────────────────────

func TestDoGet_returns_body_on_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	body, err := c.doGet(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("body = %q", string(body))
	}
}

func TestDoGet_passes_query_params(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("fuel"); got != "B7" {
			t.Errorf("fuel param = %q, want B7", got)
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := New(srv.URL)
	q := url.Values{"fuel": {"B7"}}
	_, err := c.doGet(context.Background(), "/api/stations", q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoPost_sends_json_body(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type = %q, want application/json", ct)
		}
		w.Write([]byte(`{"routed":true}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	body, err := c.doPost(context.Background(), "/api/route", map[string]any{"origin": []float64{1, 2}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), "routed") {
		t.Errorf("body = %q", string(body))
	}
}

func TestDo_returns_error_on_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.doGet(context.Background(), "/fail", nil)
	if err == nil {
		t.Fatal("expected error for 4xx response")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error = %q, want to contain 400", err.Error())
	}
}

func TestDo_returns_error_on_5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.doGet(context.Background(), "/fail", nil)
	if err == nil {
		t.Fatal("expected error for 5xx response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want to contain 500", err.Error())
	}
}

func TestDo_returns_error_on_unreachable_server(t *testing.T) {
	c := New("http://127.0.0.1:1") // port 1 is almost certainly closed
	_, err := c.doGet(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestDoPost_returns_error_for_unmarshalable_body(t *testing.T) {
	c := New("http://localhost")
	// channels cannot be marshaled to JSON
	_, err := c.doPost(context.Background(), "/test", make(chan int))
	if err == nil {
		t.Fatal("expected error for unmarshalable body")
	}
	if !strings.Contains(err.Error(), "marshal body") {
		t.Errorf("error = %q, want marshal body", err.Error())
	}
}

// ── Resource methods ─────────────────────────────────────────────────────────

func TestGetConfig_hits_correct_path(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/config" {
			t.Errorf("path = %q, want /api/config", r.URL.Path)
		}
		w.Write([]byte(`{"countries":[]}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	body, err := c.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), "countries") {
		t.Errorf("body = %q", string(body))
	}
}

func TestGetStats_hits_correct_path(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stats" {
			t.Errorf("path = %q, want /api/stats", r.URL.Path)
		}
		w.Write([]byte(`{"stations":100}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	body, err := c.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), "stations") {
		t.Errorf("body = %q", string(body))
	}
}

func TestGetExchangeRates_hits_correct_path(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/exchange-rates" {
			t.Errorf("path = %q, want /api/exchange-rates", r.URL.Path)
		}
		w.Write([]byte(`{"rates":{}}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	body, err := c.GetExchangeRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), "rates") {
		t.Errorf("body = %q", string(body))
	}
}

// ── Tool methods ─────────────────────────────────────────────────────────────

func TestFindNearestStations_sends_correct_params(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stations/nearest" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("fuel") != "B7" {
			t.Errorf("fuel = %q", q.Get("fuel"))
		}
		if q.Get("lat") == "" || q.Get("lon") == "" {
			t.Error("lat/lon missing")
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.FindNearestStations(context.Background(), 40.4, -3.7, "B7", 10, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetStationsInArea_sends_correct_params(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stations" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("bbox") != "-4,39,-3,41" {
			t.Errorf("bbox = %q", q.Get("bbox"))
		}
		if q.Get("fuel") != "E5" {
			t.Errorf("fuel = %q", q.Get("fuel"))
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.GetStationsInArea(context.Background(), "-4,39,-3,41", "E5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCalculateRoute_sends_post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/route" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Write([]byte(`{"distance":100}`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.CalculateRoute(context.Background(), map[string]any{
		"origin":      []float64{-3.7, 40.4},
		"destination": []float64{-0.37, 39.47},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindRouteStations_sends_post(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/route-stations" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.FindRouteStations(context.Background(), map[string]any{"geometry": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGeocode_sends_query_with_optional_bias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/geocode" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("q") != "Madrid" {
			t.Errorf("q = %q", q.Get("q"))
		}
		if q.Get("lat") == "" {
			t.Error("expected lat bias param")
		}
		w.Write([]byte(`[{"name":"Madrid"}]`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	lat := 40.4
	lon := -3.7
	_, err := c.Geocode(context.Background(), "Madrid", &lat, &lon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGeocode_without_bias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("lat") != "" {
			t.Error("lat should be absent when nil")
		}
		if q.Get("lon") != "" {
			t.Error("lon should be absent when nil")
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.Geocode(context.Background(), "Barcelona", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetConfig_returns_error_on_server_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.GetConfig(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetStats_returns_error_on_server_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.GetStats(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetExchangeRates_returns_error_on_server_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.GetExchangeRates(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

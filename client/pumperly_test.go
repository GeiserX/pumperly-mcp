package client

import (
	"net/url"
	"testing"
)

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

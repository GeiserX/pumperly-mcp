package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/geiserx/pumperly-mcp/client"
)

// Place is a geocoded location.
type Place struct {
	Name    string  `json:"name"`
	City    string  `json:"city,omitempty"`
	State   string  `json:"state,omitempty"`
	Country string  `json:"country,omitempty"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

func decodePlaces(body []byte) ([]Place, error) {
	var raw []struct {
		Name        string    `json:"name"`
		City        *string   `json:"city"`
		State       *string   `json:"state"`
		Country     *string   `json:"country"`
		Coordinates []float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unexpected answer from the Pumperly API (not a geocoding result)")
	}
	out := make([]Place, 0, len(raw))
	for _, r := range raw {
		if len(r.Coordinates) < 2 {
			continue
		}
		p := Place{Name: r.Name, Lon: r.Coordinates[0], Lat: r.Coordinates[1]}
		if r.City != nil {
			p.City = *r.City
		}
		if r.State != nil {
			p.State = *r.State
		}
		if r.Country != nil {
			p.Country = *r.Country
		}
		out = append(out, p)
	}
	return out, nil
}

// latLonText matches "40.4168, -3.7038": latitude first, the way people
// write coordinates.
var latLonText = regexp.MustCompile(`^\s*(-?\d{1,3}(?:\.\d+)?)\s*[,;\s]\s*(-?\d{1,3}(?:\.\d+)?)\s*$`)

// resolvePlace turns "lat,lon" text or a place name into a point, geocoding
// names through Pumperly and taking the best match.
func resolvePlace(ctx context.Context, c *client.Client, text, param string) (Place, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Place{}, fmt.Errorf("parameter %s is empty", param)
	}
	if m := latLonText.FindStringSubmatch(text); m != nil {
		lat, _ := strconv.ParseFloat(m[1], 64)
		lon, _ := strconv.ParseFloat(m[2], 64)
		if err := checkLatLon(lat, lon); err != nil {
			return Place{}, fmt.Errorf("parameter %s: %v", param, err)
		}
		return Place{Name: text, Lat: lat, Lon: lon}, nil
	}
	if len([]rune(text)) > 200 {
		return Place{}, fmt.Errorf("parameter %s is longer than 200 characters", param)
	}
	body, err := c.Geocode(ctx, text, nil, nil)
	if err != nil {
		return Place{}, err
	}
	places, err := decodePlaces(body)
	if err != nil {
		return Place{}, err
	}
	if len(places) == 0 {
		return Place{}, fmt.Errorf("no place found for %s %q; try a more specific name or give coordinates", param, text)
	}
	return places[0], nil
}

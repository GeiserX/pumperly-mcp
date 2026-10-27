package tools

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Station is one station in a compact tool result.
type Station struct {
	Name    string  `json:"name"`
	Brand   string  `json:"brand,omitempty"`
	Address string  `json:"address,omitempty"`
	City    string  `json:"city,omitempty"`
	Country string  `json:"country,omitempty" jsonschema:"ISO 3166-1 alpha-2 country code"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	// Price is absent when the station has no current price for this fuel
	// (common for EV chargers).
	Price      *float64 `json:"price,omitempty" jsonschema:"Price per unit in currency"`
	Currency   string   `json:"currency,omitempty" jsonschema:"ISO 4217 currency of price"`
	Unit       string   `json:"unit,omitempty" jsonschema:"What the price is per: L, kg or kWh"`
	ReportedAt string   `json:"reported_at,omitempty" jsonschema:"When the price was last reported (RFC 3339)"`
	DistanceKm *float64 `json:"distance_km,omitempty" jsonschema:"Straight-line distance from the search point, km"`
	RouteKm    *float64 `json:"route_km,omitempty" jsonschema:"Distance along the route from its start, km"`
	PowerKw    *float64 `json:"power_kw,omitempty" jsonschema:"EV only: maximum charging power, kW"`
	// EVFees are fees on top of the per-kWh price, as Pumperly reports them.
	EVFees map[string]any `json:"ev_fees,omitempty"`
}

// PriceSummary describes the prices of one currency in a result.
type PriceSummary struct {
	Currency string  `json:"currency"`
	Count    int     `json:"count"`
	Min      float64 `json:"min"`
	Median   float64 `json:"median"`
	Max      float64 `json:"max"`
}

// StationList is the structured result of every station tool.
type StationList struct {
	// Center is the search point (find_nearest_stations only).
	Center *Place `json:"center,omitempty"`
	Fuel   string `json:"fuel"`
	Unit   string `json:"unit" jsonschema:"What prices are per: L, kg or kWh"`
	// SortBy is how Stations are ordered: price (cheapest first, unpriced
	// last), distance or route (along the route).
	SortBy string `json:"sort_by" jsonschema:"enum=price,enum=distance,enum=route"`
	// Found is how many stations matched before the limit was applied.
	Found    int            `json:"found"`
	Returned int            `json:"returned"`
	Prices   []PriceSummary `json:"prices"`
	Stations []Station      `json:"stations"`
}

type geoFeature struct {
	Geometry struct {
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties struct {
		Name          string         `json:"name"`
		Brand         *string        `json:"brand"`
		Address       string         `json:"address"`
		City          string         `json:"city"`
		Country       string         `json:"country"`
		Price         *float64       `json:"price"`
		Currency      string         `json:"currency"`
		ReportedAt    *string        `json:"reportedAt"`
		DistanceKm    *float64       `json:"distanceKm"`
		RouteFraction *float64       `json:"routeFraction"`
		PowerKw       *float64       `json:"powerKw"`
		EVFees        map[string]any `json:"evFees"`
	} `json:"properties"`
}

// decodeStations parses a Pumperly station FeatureCollection. routeKm, when
// positive, turns each station's routeFraction into km along the route.
func decodeStations(body []byte, fuel string, routeKm float64) ([]Station, error) {
	var fc struct {
		Type     string       `json:"type"`
		Features []geoFeature `json:"features"`
	}
	if err := json.Unmarshal(body, &fc); err != nil || fc.Type != "FeatureCollection" {
		return nil, fmt.Errorf("unexpected answer from the Pumperly API (not a station list)")
	}
	unit := fuelUnit(fuel)
	out := make([]Station, 0, len(fc.Features))
	for _, f := range fc.Features {
		p := f.Properties
		s := Station{
			Name:     p.Name,
			Address:  p.Address,
			City:     p.City,
			Country:  p.Country,
			Price:    p.Price,
			Currency: p.Currency,
			Unit:     unit,
			PowerKw:  p.PowerKw,
			EVFees:   p.EVFees,
		}
		if p.Brand != nil {
			s.Brand = *p.Brand
		}
		if p.ReportedAt != nil {
			s.ReportedAt = *p.ReportedAt
		}
		if len(f.Geometry.Coordinates) >= 2 {
			s.Lon, s.Lat = f.Geometry.Coordinates[0], f.Geometry.Coordinates[1]
		}
		if p.DistanceKm != nil {
			d := round(*p.DistanceKm, 2)
			s.DistanceKm = &d
		}
		if p.RouteFraction != nil && routeKm > 0 {
			d := round(*p.RouteFraction*routeKm, 1)
			s.RouteKm = &d
		}
		if s.Price == nil {
			s.Currency = ""
		}
		out = append(out, s)
	}
	return out, nil
}

// sortStations orders stations in place. Stations without the sort key go
// last; ties keep the API's order.
func sortStations(st []Station, by string) {
	key := func(s Station) *float64 {
		switch by {
		case "price":
			return s.Price
		case "route":
			return s.RouteKm
		default:
			return s.DistanceKm
		}
	}
	sort.SliceStable(st, func(i, j int) bool {
		a, b := key(st[i]), key(st[j])
		switch {
		case a == nil:
			return false
		case b == nil:
			return true
		default:
			return *a < *b
		}
	})
}

func summarisePrices(st []Station) []PriceSummary {
	byCur := map[string][]float64{}
	var order []string
	for _, s := range st {
		if s.Price == nil {
			continue
		}
		if _, ok := byCur[s.Currency]; !ok {
			order = append(order, s.Currency)
		}
		byCur[s.Currency] = append(byCur[s.Currency], *s.Price)
	}
	sort.Strings(order)
	out := make([]PriceSummary, 0, len(order))
	for _, c := range order {
		p := byCur[c]
		sort.Float64s(p)
		med := p[len(p)/2]
		if len(p)%2 == 0 {
			med = (p[len(p)/2-1] + p[len(p)/2]) / 2
		}
		out = append(out, PriceSummary{Currency: c, Count: len(p), Min: p[0], Median: round(med, 4), Max: p[len(p)-1]})
	}
	return out
}

// newStationList sorts, summarises and cuts a station slice to limit.
func newStationList(st []Station, fuel, sortBy string, limit int) StationList {
	sortStations(st, sortBy)
	l := StationList{
		Fuel:   fuel,
		Unit:   fuelUnit(fuel),
		SortBy: sortBy,
		Found:  len(st),
		Prices: summarisePrices(st),
	}
	if limit > 0 && len(st) > limit {
		st = st[:limit]
	}
	l.Stations = st
	l.Returned = len(st)
	return l
}

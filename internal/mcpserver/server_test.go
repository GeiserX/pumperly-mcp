package mcpserver

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/geiserx/pumperly-mcp/client"
)

const legacyInit = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"t","version":"0"}}}`

func legacyPeer(t *testing.T, c *client.Client) *stdioPeer {
	t.Helper()
	p := newStdioPeerWith(t, c)
	result(t, p.call(legacyInit))
	p.notify(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	return p
}

func listTools(t *testing.T, p *stdioPeer) []map[string]any {
	t.Helper()
	r := result(t, p.call(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`))
	raw, _ := r["tools"].([]any)
	var out []map[string]any
	for _, x := range raw {
		out = append(out, x.(map[string]any))
	}
	return out
}

func TestTools_every_tool_is_annotated_read_only_with_schemas(t *testing.T) {
	tools := listTools(t, legacyPeer(t, client.New("http://127.0.0.1:1")))

	var names []string
	for _, tool := range tools {
		name, _ := tool["name"].(string)
		names = append(names, name)
		if title, _ := tool["title"].(string); title == "" {
			t.Errorf("%s: no title", name)
		}
		a, _ := tool["annotations"].(map[string]any)
		for hint, want := range map[string]bool{"readOnlyHint": true, "destructiveHint": false, "idempotentHint": true, "openWorldHint": true} {
			if got, ok := a[hint].(bool); !ok || got != want {
				t.Errorf("%s: %s = %v, want %v", name, hint, a[hint], want)
			}
		}
		if os, _ := tool["outputSchema"].(map[string]any); len(os) == 0 {
			t.Errorf("%s: no outputSchema", name)
		}
		in, _ := tool["inputSchema"].(map[string]any)
		if in["additionalProperties"] != false {
			t.Errorf("%s: inputSchema accepts unknown arguments", name)
		}
		if d, _ := tool["description"].(string); strings.Contains(d, "E85") {
			t.Errorf("%s: description still offers E85, which the API rejects", name)
		}
	}
	want := []string{"calculate_route", "cheapest_fuel_on_route", "find_nearest_stations", "find_route_stations", "geocode", "get_stations_in_area"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("tools = %v, want %v in this order", names, want)
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("tools/list is not in a deterministic order: %v", names)
	}
}

func TestTools_limit_parameters_are_integers_and_fuel_is_an_enum(t *testing.T) {
	for _, tool := range listTools(t, legacyPeer(t, client.New("http://127.0.0.1:1"))) {
		props, _ := tool["inputSchema"].(map[string]any)["properties"].(map[string]any)
		if l, ok := props["limit"].(map[string]any); ok && l["type"] != "integer" {
			t.Errorf("%s: limit type = %v", tool["name"], l["type"])
		}
		if f, ok := props["fuel"].(map[string]any); ok {
			if enum, _ := f["enum"].([]any); len(enum) != 16 {
				t.Errorf("%s: fuel enum has %d codes, want 16", tool["name"], len(enum))
			}
		}
	}
}

func TestTools_invalid_arguments_never_reach_the_api(t *testing.T) {
	var hits atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	}))
	defer api.Close()
	p := legacyPeer(t, client.New(api.URL))

	for _, args := range []string{
		`{"lat":40,"lon":-3,"fuel":"E85"}`,
		`{"lat":40,"lon":-3,"fuel":"B7","limit":500}`,
		`{"lat":40,"lon":-3,"fuel":"B7","colour":"red"}`,
	} {
		r := result(t, p.call(`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"find_nearest_stations","arguments":`+args+`}}`))
		if r["isError"] != true {
			t.Errorf("%s: not rejected: %v", args, r)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("%d upstream calls for invalid arguments", n)
	}

	// Positive control: valid arguments do reach the API and pass output validation.
	r := result(t, p.call(`{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"find_nearest_stations","arguments":{"lat":40,"lon":-3,"fuel":"B7"}}}`))
	if r["isError"] == true || hits.Load() != 1 {
		t.Errorf("valid call failed: %v (hits %d)", r, hits.Load())
	}
	if sc, _ := r["structuredContent"].(map[string]any); sc["fuel"] != "B7" {
		t.Errorf("structuredContent = %v", r["structuredContent"])
	}
}

func TestServer_identity_and_instructions(t *testing.T) {
	p := newStdioPeer(t)
	init := result(t, p.call(legacyInit))
	if ins, _ := init["instructions"].(string); !strings.Contains(ins, "cheapest_fuel_on_route") || !strings.Contains(ins, "longitude first") {
		t.Errorf("instructions = %q", ins)
	}
	info, _ := init["serverInfo"].(map[string]any)
	if info["title"] != "Pumperly" || info["websiteUrl"] == nil || info["icons"] == nil {
		t.Errorf("serverInfo = %v", info)
	}
}

func TestServer_modern_lists_are_cacheable_for_an_hour(t *testing.T) {
	p := newStdioPeer(t)
	for _, method := range []string{"tools/list", "resources/list"} {
		r := result(t, p.call(`{"jsonrpc":"2.0","id":5,"method":"`+method+`","params":{`+modernMeta+`}}`))
		if r["ttlMs"] != float64(3600000) || r["cacheScope"] != "public" {
			t.Errorf("%s: ttlMs %v cacheScope %v", method, r["ttlMs"], r["cacheScope"])
		}
	}
}

func TestResources_fuel_types_lists_the_16_codes(t *testing.T) {
	p := legacyPeer(t, client.New("http://127.0.0.1:1"))
	r := result(t, p.call(`{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"pumperly://fuel-types"}}`))
	contents, _ := r["contents"].([]any)
	text, _ := contents[0].(map[string]any)["text"].(string)
	for _, code := range []string{`"B7"`, `"E5_98"`, `"ADBLUE"`, `"H2"`, `"EV"`, `"kWh"`} {
		if !strings.Contains(text, code) {
			t.Errorf("fuel-types lacks %s: %s", code, text)
		}
	}
	if strings.Contains(text, "E85") {
		t.Error("fuel-types lists E85")
	}
}

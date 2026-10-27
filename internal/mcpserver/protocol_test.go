package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/geiserx/pumperly-mcp/client"
	"github.com/mark3labs/mcp-go/server"
)

// stdioPeer drives the real stdio transport in-process, line by line, the
// way a client process would.
type stdioPeer struct {
	t   *testing.T
	in  *io.PipeWriter
	out *bufio.Reader
}

func newStdioPeer(t *testing.T) *stdioPeer {
	t.Helper()
	return newStdioPeerWith(t, client.New("http://127.0.0.1:1"))
}

func newStdioPeerWith(t *testing.T, c *client.Client) *stdioPeer {
	t.Helper()
	s := New(c)
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = server.NewStdioServer(s).Listen(ctx, inR, outW)
		_ = outW.Close()
	}()
	t.Cleanup(func() {
		_ = inW.Close()
		cancel()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("stdio server did not stop on stdin EOF")
		}
	})
	return &stdioPeer{t: t, in: inW, out: bufio.NewReader(outR)}
}

// call writes one JSON-RPC line and returns the next response line, decoded.
func (p *stdioPeer) call(msg string) map[string]any {
	p.t.Helper()
	if _, err := io.WriteString(p.in, msg+"\n"); err != nil {
		p.t.Fatalf("write: %v", err)
	}
	type lineOrErr struct {
		line []byte
		err  error
	}
	ch := make(chan lineOrErr, 1)
	go func() {
		l, err := p.out.ReadBytes('\n')
		ch <- lineOrErr{l, err}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			p.t.Fatalf("read: %v", r.err)
		}
		var m map[string]any
		if err := json.Unmarshal(r.line, &m); err != nil {
			p.t.Fatalf("decode %q: %v", r.line, err)
		}
		return m
	case <-time.After(5 * time.Second):
		p.t.Fatalf("no answer to %s", msg)
		return nil
	}
}

// notify writes a notification, which gets no answer.
func (p *stdioPeer) notify(msg string) {
	p.t.Helper()
	if _, err := io.WriteString(p.in, msg+"\n"); err != nil {
		p.t.Fatalf("write: %v", err)
	}
}

const modernMeta = `"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28","io.modelcontextprotocol/clientCapabilities":{},"io.modelcontextprotocol/clientInfo":{"name":"smoke","version":"0"}}`

func result(t *testing.T, resp map[string]any) map[string]any {
	t.Helper()
	r, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("no result in %v", resp)
	}
	return r
}

func errorCode(t *testing.T, resp map[string]any) float64 {
	t.Helper()
	e, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected an error, got %v", resp)
	}
	code, _ := e["code"].(float64)
	return code
}

func TestProtocol_modern_discover_lists_2026_07_28_and_legacy_versions(t *testing.T) {
	p := newStdioPeer(t)
	r := result(t, p.call(`{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{`+modernMeta+`}}`))
	got := map[string]bool{}
	versions, _ := r["supportedVersions"].([]any)
	for _, v := range versions {
		s, _ := v.(string)
		got[s] = true
	}
	for _, want := range []string{"2026-07-28", "2025-11-25", "2024-11-05"} {
		if !got[want] {
			t.Errorf("supportedVersions = %v, missing %s", versions, want)
		}
	}
}

func TestProtocol_modern_tools_list_is_a_cacheable_complete_result(t *testing.T) {
	p := newStdioPeer(t)
	r := result(t, p.call(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{`+modernMeta+`}}`))
	if r["resultType"] != "complete" {
		t.Errorf("resultType = %v, want complete", r["resultType"])
	}
	if _, ok := r["ttlMs"]; !ok {
		t.Errorf("ttlMs missing from %v", r)
	}
	if _, ok := r["cacheScope"]; !ok {
		t.Errorf("cacheScope missing from %v", r)
	}
	if tools, _ := r["tools"].([]any); len(tools) == 0 {
		t.Errorf("no tools listed")
	}
}

func TestProtocol_modern_ping_is_method_not_found(t *testing.T) {
	p := newStdioPeer(t)
	if code := errorCode(t, p.call(`{"jsonrpc":"2.0","id":3,"method":"ping","params":{`+modernMeta+`}}`)); code != -32601 {
		t.Errorf("ping error code = %v, want -32601", code)
	}
}

func TestProtocol_unknown_version_is_rejected_with_32022(t *testing.T) {
	p := newStdioPeer(t)
	msg := `{"jsonrpc":"2.0","id":4,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2099-01-01","io.modelcontextprotocol/clientCapabilities":{}}}}`
	if code := errorCode(t, p.call(msg)); code != -32022 {
		t.Errorf("error code = %v, want -32022", code)
	}
}

func TestProtocol_legacy_initialize_then_tools_list_still_works(t *testing.T) {
	p := newStdioPeer(t)
	init := result(t, p.call(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"legacy","version":"0"}}}`))
	if init["protocolVersion"] != "2025-11-25" {
		t.Errorf("protocolVersion = %v, want 2025-11-25", init["protocolVersion"])
	}
	p.notify(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	r := result(t, p.call(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`))
	if tools, _ := r["tools"].([]any); len(tools) == 0 {
		t.Errorf("legacy tools/list returned no tools: %v", r)
	}
	if _, ok := r["resultType"]; ok {
		t.Errorf("legacy result carries the modern resultType field: %v", r)
	}
}

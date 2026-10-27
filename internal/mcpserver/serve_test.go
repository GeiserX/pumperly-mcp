package mcpserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// A cancelled context must not cut off a request that is already running:
// Serve returns only after the handler has finished and the client got its
// answer.
func TestServeDrainsInFlightRequests(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	var finished atomic.Bool
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(300 * time.Millisecond)
		finished.Store(true)
		_, _ = io.WriteString(w, "done")
	})}

	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- Serve(ctx, srv, ln, 5*time.Second) }()

	type reply struct {
		body string
		err  error
	}
	got := make(chan reply, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err != nil {
			got <- reply{err: err}
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		got <- reply{string(b), err}
	}()

	<-started
	cancel()
	if err := <-served; err != nil {
		t.Fatalf("Serve returned %v", err)
	}
	if !finished.Load() {
		t.Fatal("Serve returned while a request was still running")
	}
	if r := <-got; r.err != nil || r.body != "done" {
		t.Fatalf("in-flight request got %q, %v", r.body, r.err)
	}
}

// A listener error is returned, not swallowed.
func TestServeReturnsListenerError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ln.Close()
	if err := Serve(context.Background(), &http.Server{}, ln, time.Second); err == nil {
		t.Fatal("Serve on a closed listener returned nil")
	}
}

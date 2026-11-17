package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// An in-flight request finishes before Serve returns.
func TestServeWaitsForInFlightRequests(t *testing.T) {
	started := make(chan struct{})
	var finished atomic.Bool
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte("done"))
		finished.Store(true)
	})}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- Serve(ctx, srv, ln) }()

	body := make(chan string, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			body <- "error: " + err.Error()
			return
		}
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(resp.Body)
		body <- string(b)
	}()

	<-started
	cancel()

	if err := <-served; err != nil {
		t.Fatalf("Serve returned %v", err)
	}
	if !finished.Load() {
		t.Fatal("Serve returned before the in-flight request finished")
	}
	select {
	case got := <-body:
		if got != "done" {
			t.Fatalf("in-flight request got %q, want done", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("client never got a response")
	}
}

func TestServeReturnsListenerErrors(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_ = ln.Close()
	if err := Serve(context.Background(), &http.Server{}, ln); err == nil {
		t.Fatal("Serve on a closed listener returned nil")
	}
}

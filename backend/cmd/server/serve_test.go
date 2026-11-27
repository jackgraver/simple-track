package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// SIGTERM (ctx cancelled) lets a request that is already running finish, then serve returns.
func TestServeDrainsInFlightRequests(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		io.WriteString(w, "done")
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := newHTTPServer(ln.Addr().String(), mux)
	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- serveListener(ctx, srv, ln, nil, 5*time.Second) }()

	type result struct {
		body string
		err  error
	}
	got := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/slow")
		if err != nil {
			got <- result{err: err}
			return
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		got <- result{string(b), err}
	}()

	<-started
	cancel() // the signal arrives mid-request
	select {
	case err := <-served:
		t.Fatalf("serve returned before the in-flight request finished: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	// New connections are refused while draining.
	if c, err := net.DialTimeout("tcp", ln.Addr().String(), time.Second); err == nil {
		c.Close()
		t.Error("listener still accepting during shutdown")
	}
	close(release)

	if r := <-got; r.err != nil || r.body != "done" {
		t.Fatalf("in-flight request: body=%q err=%v", r.body, r.err)
	}
	if err := <-served; err != nil {
		t.Fatalf("serve: %v", err)
	}
}

// Requests still running after the grace period are cut off and serve reports it.
func TestServeGivesUpAfterGrace(t *testing.T) {
	started := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/stuck", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() {
		served <- serveListener(ctx, newHTTPServer(ln.Addr().String(), mux), ln, nil, 200*time.Millisecond)
	}()
	go http.Get("http://" + ln.Addr().String() + "/stuck")
	<-started
	cancel()
	select {
	case err := <-served:
		if err == nil {
			t.Fatal("want a shutdown error after the grace period")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not give up after the grace period")
	}
}

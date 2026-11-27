package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// HTTP server timeouts. Requests are small JSON calls; the slowest handlers are the media lookups
// (10 s outbound client timeout) and queries bounded by the app role's 15 s statement_timeout,
// so writeTimeout leaves room for both. Reasoning: docs/multi-user/DECISIONS.md.
const (
	readHeaderTimeout = 5 * time.Second   // slow-loris guard: headers must arrive promptly
	readTimeout       = 30 * time.Second  // whole request incl. body (imports are small CSVs)
	writeTimeout      = 45 * time.Second  // handler + response
	idleTimeout       = 120 * time.Second // keep-alive between requests
	// shutdownTimeout is how long in-flight requests get to finish after SIGTERM. Keep it under
	// the container's stop grace period (stop_grace_period in docker-compose.vps.yml, 30 s),
	// after which Docker sends SIGKILL.
	shutdownTimeout = 25 * time.Second
)

func newHTTPServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// serve listens on srv.Addr until ctx is cancelled (SIGTERM/SIGINT), then stops accepting
// connections, waits up to shutdownTimeout for in-flight requests and closes the database pool.
// A deploy therefore drains requests instead of cutting them off mid-transaction.
func serve(ctx context.Context, srv *http.Server, db *gorm.DB) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	log.Printf("server listening on %s", ln.Addr())
	return serveListener(ctx, srv, ln, db, shutdownTimeout)
}

func serveListener(ctx context.Context, srv *http.Server, ln net.Listener, db *gorm.DB, grace time.Duration) error {
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		// Serve only returns early on a listener failure.
		closeDB(db)
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	log.Printf("shutting down: draining requests (up to %s)", grace)
	sctx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()
	shutdownErr := srv.Shutdown(sctx)
	if shutdownErr != nil {
		// Requests still running after the grace period: cut them off.
		log.Printf("shutdown: %v; closing remaining connections", shutdownErr)
		srv.Close()
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("http server: %v", err)
	}
	closeDB(db)
	log.Print("shutdown complete")
	if shutdownErr != nil {
		return fmt.Errorf("shutdown: %w", shutdownErr)
	}
	return nil
}

func closeDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}
}

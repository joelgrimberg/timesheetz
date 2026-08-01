package main

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"
)

// maybeStartPprof starts net/http/pprof on localhost:6060 when
// TIMESHEETZ_PPROF=1. The endpoint is bound to loopback only so it is not
// exposed on the network. When the env var is unset, this is a no-op with
// zero cost.
func maybeStartPprof() {
	if os.Getenv("TIMESHEETZ_PPROF") != "1" {
		return
	}

	srv := &http.Server{
		Addr:              "127.0.0.1:6060",
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("pprof listening on http://%s/debug/pprof/", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("pprof server error: %v", err)
		}
	}()
}

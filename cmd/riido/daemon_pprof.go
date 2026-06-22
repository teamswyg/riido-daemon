package main

import (
	"errors"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

const daemonPprofShutdownTimeout = 2 * time.Second

func startDaemonPprofServer(ctx lifecycle.Context, addr string, log logging.Logger) (func(), string, error) {
	if addr == "" {
		return func() {}, "", nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	listener := net.ListenConfig{}
	ln, err := listener.Listen(ctx.Context(), "tcp", addr)
	if err != nil {
		return nil, "", daemonWrapf(ErrDaemonSocket, "pprof.listen", err, "listen pprof %s", addr)
	}
	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	actualAddr := ln.Addr().String()
	done := make(chan struct{})
	shutdownCh := make(chan struct{})
	var requestShutdown sync.Once
	var shutdownServer sync.Once
	stopServer := func() {
		shutdownServer.Do(func() {
			shutdownDaemonPprofServer(server)
		})
	}
	go func() {
		select {
		case <-ctx.Done():
		case <-shutdownCh:
		}
		stopServer()
	}()
	go func() {
		defer close(done)
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("pprof server error: %v", err)
		}
	}()
	log.Printf("pprof listening addr=%s path=/debug/pprof/", actualAddr)
	return func() {
		requestShutdown.Do(func() { close(shutdownCh) })
		stopServer()
		select {
		case <-done:
		case <-time.After(daemonPprofShutdownTimeout):
		}
	}, actualAddr, nil
}

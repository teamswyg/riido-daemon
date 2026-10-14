package main

import (
	"context"
	"net/http"
)

func shutdownDaemonPprofServer(server *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), daemonPprofShutdownTimeout)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

package riidoapi

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func shortRiidoAPISocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "riido-api-timeout-")
	if err != nil {
		t.Fatalf("create socket dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "api.sock")
}

func serveStalledRiidoAPI(t *testing.T, socketPath string) func() {
	t.Helper()
	listener, cleanup, err := listenLocalEndpoint(LocalTransportUnixSocket, socketPath)
	if err != nil {
		t.Fatalf("listen riido API socket: %v", err)
	}
	release := make(chan struct{})
	done := make(chan struct{})
	go serveStalledRiidoAPILoop(listener, release, done)
	return func() {
		close(release)
		cleanup()
		<-done
	}
}

func serveStalledRiidoAPILoop(listener net.Listener, release <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go stallRiidoAPIConn(conn, release)
	}
}

func stallRiidoAPIConn(conn net.Conn, release <-chan struct{}) {
	defer conn.Close()
	var req requestEnvelope
	_ = json.NewDecoder(conn).Decode(&req)
	<-release
}

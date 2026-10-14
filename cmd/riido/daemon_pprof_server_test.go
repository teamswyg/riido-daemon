package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestStartDaemonPprofServerServesIndex(t *testing.T) {
	ctx, cancel := lifecycle.WithCancel(lifecycle.Background())
	defer cancel()
	stop, addr, err := startDaemonPprofServer(ctx, "127.0.0.1:0", logging.NewWriterLogger(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer stop()

	body := readDaemonPprofIndex(t, addr)
	if !strings.Contains(body, "profile") || !strings.Contains(body, "goroutine") {
		t.Fatalf("pprof index body does not look like pprof: %q", body)
	}
}

func TestStartDaemonPprofServerStopIsIdempotent(t *testing.T) {
	ctx, cancel := lifecycle.WithCancel(lifecycle.Background())
	defer cancel()
	stop, _, err := startDaemonPprofServer(ctx, "127.0.0.1:0", logging.NewWriterLogger(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	stop()
	stop()
}

func readDaemonPprofIndex(t *testing.T, addr string) string {
	t.Helper()
	client := http.Client{Timeout: time.Second}
	res, err := client.Get("http://" + addr + "/debug/pprof/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

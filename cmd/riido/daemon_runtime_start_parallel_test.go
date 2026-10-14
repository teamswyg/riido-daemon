package main

import (
	"io"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func TestStartDaemonRuntimesDetectsProvidersConcurrently(t *testing.T) {
	adapters := make([]agentbridge.Adapter, 4)
	for i := range adapters {
		adapters[i] = daemonTestAdapter{name: string(rune('a' + i)), delay: 150 * time.Millisecond}
	}
	runtimes, err := newDaemonRuntimeActors(dynamicSaaSRuntimeSettings(), adapters, nil)
	if err != nil {
		t.Fatal(err)
	}
	startedAt := time.Now()
	log := logging.NewWriterLogger(io.Discard)
	if err := startDaemonRuntimes(lifecycle.Background(), runtimes, log); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stopDaemonRuntimes(lifecycle.ShutdownGraceful, runtimes, log) })
	if elapsed := time.Since(startedAt); elapsed >= 450*time.Millisecond {
		t.Fatalf("runtime startup took %s; provider detection appears serial", elapsed)
	}
}

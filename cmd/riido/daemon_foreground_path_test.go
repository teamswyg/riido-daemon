package main

import (
	"context"
	"testing"
	"time"
)

func TestForegroundPathStillWorks(t *testing.T) {
	paths := newDaemonBackgroundPaths(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- runDaemonWithContext(
			ctx,
			[]string{"start", "--foreground", "--socket", paths.socket, "--lock-file", paths.lock},
		)
	}()
	if !waitForSocket(paths.socket, 2*time.Second) {
		t.Fatal("foreground daemon did not bind socket")
	}
	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("foreground daemon did not exit on ctx cancel")
	}
}

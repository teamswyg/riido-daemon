package main

import (
	"context"
	"testing"
	"time"
)

func TestDaemonStartHoldsSingletonLock(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	t.Setenv(envTaskQueueDir, "")
	t.Setenv(envTaskReportDir, "")

	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runDaemonWithContext(ctx, []string{
			"start", "--foreground",
			"--socket", sock,
			"--lock-file", lockPath,
		})
	}()
	dialDaemon(t, sock, 2*time.Second)

	secondErr := make(chan error, 1)
	go func() {
		secondErr <- runDaemonWithContext(context.Background(), []string{
			"start", "--foreground",
			"--socket", daemonSocketPath(t),
			"--lock-file", lockPath,
		})
	}()
	select {
	case err := <-secondErr:
		if err == nil {
			t.Fatal("expected second daemon start to fail while singleton lock is held")
		}
	case <-time.After(time.Second):
		t.Fatal("second foreground daemon remained blocked on singleton lock")
	}

	assertForegroundDaemonExits(t, cancel, errCh)
}

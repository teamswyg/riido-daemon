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

	secondCtx, secondCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer secondCancel()
	err := runDaemonWithContext(secondCtx, []string{
		"start", "--foreground",
		"--socket", daemonSocketPath(t),
		"--lock-file", lockPath,
	})
	if err == nil {
		t.Fatal("expected second daemon start to fail while singleton lock is held")
	}

	assertForegroundDaemonExits(t, cancel, errCh)
}

package main

import (
	"context"
	"testing"
	"time"
)

type foregroundDaemonRun struct {
	socket string
	cancel context.CancelFunc
	errCh  <-chan error
}

func startForegroundDaemonForStatus(t *testing.T) foregroundDaemonRun {
	t.Helper()
	configureForegroundStatusEnv(t)
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runDaemonWithContext(ctx, []string{
			"start", "--foreground",
			"--socket", sock,
			"--lock-file", lockPath,
		})
	}()
	dialDaemon(t, sock, 5*time.Second)
	return foregroundDaemonRun{socket: sock, cancel: cancel, errCh: errCh}
}

func configureForegroundStatusEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envDaemonID, "daemon-test-1")
	t.Setenv(envDaemonVersion, "riido-agentd v1.2.3")
	t.Setenv(envDaemonProfile, "desktop-api.riido.ai")
	t.Setenv(envServerURL, "https://api.riido.ai")
	t.Setenv(envDeviceName, "MacBook-Pro-SK.local")
	t.Setenv(envRuntimeOwner, "kim")
	t.Setenv(envRuntimeAgents, "Riido, Orion")
	t.Setenv(envWorkspaceCount, "2")
	t.Setenv(envTaskQueueDir, "")
	t.Setenv(envTaskReportDir, "")
}

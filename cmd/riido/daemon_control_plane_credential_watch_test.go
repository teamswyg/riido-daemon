package main

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

type credentialRejectionProbe struct {
	rejected chan error
}

func (p credentialRejectionProbe) DeviceCredentialRejected() <-chan error {
	return p.rejected
}

func TestCredentialRejectionCancelsManagedDaemon(t *testing.T) {
	daemonCtx, cancelDaemon := lifecycle.WithCancel(lifecycle.Background())
	defer cancelDaemon()
	probe := credentialRejectionProbe{rejected: make(chan error, 1)}
	stop := watchDaemonDeviceCredentialRejection(
		daemonCtx,
		probe,
		cancelDaemon,
		logging.NewWriterLogger(io.Discard),
	)
	defer stop()

	probe.rejected <- errors.New("401 Unauthorized")
	select {
	case <-daemonCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for managed daemon cancellation")
	}
}

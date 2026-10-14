package codex

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func startedProtocolDriver(t *testing.T) (agentbridge.ProtocolDriver, *recordingIO) {
	t.Helper()
	d, err := NewProtocolDriver(agentbridge.StartRequest{})
	if err != nil {
		t.Fatalf("NewProtocolDriver: %v", err)
	}
	io := newRecordingIO()
	if err := d.OnStart(context.Background(), io); err != nil {
		t.Fatalf("OnStart: %v", err)
	}
	_ = io.next(t, time.Second)
	return d, io
}

package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverImplementsSessionInterface(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	if d == nil {
		t.Fatal("driver is nil")
	}
	// Compile-time check: assignable to agentbridge.ProtocolDriver.
	_ = d
}

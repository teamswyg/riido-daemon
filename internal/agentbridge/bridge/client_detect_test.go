package bridge

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNewRequiresAdapter(t *testing.T) {
	_, err := New(Config{})
	if err == nil {
		t.Fatal("expected error without adapters")
	}
}

func TestDetectReturnsCapabilities(t *testing.T) {
	a := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{Available: true, Version: "1.0"}}
	b := &stubAdapter{name: "codex", detected: agentbridge.DetectResult{Available: false, Reason: "not in path"}}
	c, err := New(Config{Adapters: []agentbridge.Adapter{a, b}})
	if err != nil {
		t.Fatal(err)
	}
	caps, err := c.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(caps) != 2 {
		t.Fatalf("want 2 caps, got %d", len(caps))
	}
	if caps[0].Provider != "claude" || !caps[0].Result.Available {
		t.Fatalf("claude detect: %+v", caps[0])
	}
	if caps[1].Provider != "codex" || caps[1].Result.Available {
		t.Fatalf("codex detect: %+v", caps[1])
	}
}

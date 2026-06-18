package bridge

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRunCancel(t *testing.T) {
	a := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{Available: true}}
	fake := process.NewFake()
	fake.NextRunning = process.NewFakeRunning()
	c, _ := New(Config{Adapters: []agentbridge.Adapter{a}, Process: fake})

	sess, _ := c.Run(context.Background(), TaskRequest{ID: "t-2", Provider: "claude"})
	sess.Cancel(nil)

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCancelled {
			t.Fatalf("expected cancelled, got %s", res.Status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel timeout")
	}
}

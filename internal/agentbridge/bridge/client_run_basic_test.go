package bridge

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRunUnknownProvider(t *testing.T) {
	a := &stubAdapter{name: "claude"}
	c, _ := New(Config{Adapters: []agentbridge.Adapter{a}})
	_, err := c.Run(context.Background(), TaskRequest{Provider: "ghost"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestRunReachesCompletion(t *testing.T) {
	a := &stubAdapter{name: "claude", detected: agentbridge.DetectResult{Available: true}}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running

	c, _ := New(Config{
		Adapters: []agentbridge.Adapter{a},
		Process:  fake,
	})

	sess, err := c.Run(context.Background(), TaskRequest{
		ID:       "t-1",
		Provider: "claude",
		Prompt:   "hello",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	go func() {
		running.EmitStdout([]byte("hello"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "hello" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for result")
	}
}

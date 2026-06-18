package bridge

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRunInstallsProtocolDriverProvider(t *testing.T) {
	driver := &driverSpy{started: make(chan struct{})}
	a := &protocolAdapter{
		stubAdapter: stubAdapter{name: "codex", detected: agentbridge.DetectResult{Available: true}},
		driver:      driver,
	}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running
	c, _ := New(Config{
		Adapters: []agentbridge.Adapter{a},
		Process:  fake,
	})

	sess, err := c.Run(context.Background(), TaskRequest{ID: "t-driver", Provider: "codex", Prompt: "hello"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	select {
	case <-driver.started:
	case <-time.After(2 * time.Second):
		t.Fatal("protocol driver OnStart was not called")
	}

	go func() {
		running.EmitStdout([]byte("ok"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted || res.Output != "ok" {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for result")
	}
}

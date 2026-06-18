package bridge

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestRunPassesAdapterTempFilesToSessionCleanup(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "mcp-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	a := &stubAdapter{
		name:     "claude",
		detected: agentbridge.DetectResult{Available: true},
		startCommand: agentbridge.StartCommand{
			Executable: "claude",
			TempFiles:  []string{tempFile.Name()},
		},
	}
	fake := process.NewFake()
	running := process.NewFakeRunning()
	fake.NextRunning = running

	c, _ := New(Config{
		Adapters: []agentbridge.Adapter{a},
		Process:  fake,
	})
	sess, err := c.Run(context.Background(), TaskRequest{ID: "t-tempfile", Provider: "claude"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	go func() {
		running.EmitStdout([]byte("ok"))
		running.EmitExit(0, nil)
	}()

	select {
	case res := <-sess.Result():
		if res.Status != agentbridge.ResultCompleted {
			t.Fatalf("result: %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for result")
	}
	if _, err := os.Stat(tempFile.Name()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed by bridge-run session, stat err=%v", err)
	}
}

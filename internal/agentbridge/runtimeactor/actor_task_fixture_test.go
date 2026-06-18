package runtimeactor

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func submitFakeTask(t *testing.T, a *Actor, id string) *SessionHandle {
	t.Helper()
	h, err := a.Submit(context.Background(), bridge.TaskRequest{ID: id, Provider: "fake"})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func emitCompletedOutput(running *process.FakeRunning) {
	go func() {
		running.EmitStdout([]byte("done"))
		running.EmitExit(0, nil)
	}()
}

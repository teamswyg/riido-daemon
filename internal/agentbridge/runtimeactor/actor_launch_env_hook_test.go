package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorLaunchEnvHookReachesStartRequest(t *testing.T) {
	startReqCh := make(chan agentbridge.StartRequest, 1)
	var hookRuntimeID, hookTaskID string
	a, p := startActor(t, Config{
		RuntimeID:            "rt-test",
		Adapters:             []agentbridge.Adapter{detectedOpenClawAdapter("/opt/riido/bin/openclaw", startReqCh)},
		MaxConcurrent:        1,
		ContextCLIExecutable: "/daemon/owned/riido",
		LaunchEnv: func(req bridge.TaskRequest, runtimeID string, env map[string]string) {
			hookRuntimeID, hookTaskID = runtimeID, req.ID
			env["RIIDO_RUN_TOKEN"] = "rrc_test"
		},
	})
	h, err := a.Submit(context.Background(), bridge.TaskRequest{
		ID: "t-hook", Provider: "openclaw", Prompt: "hi", Env: map[string]string{"PATH": "/riido/test/bin", "KEEP": "1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := waitForRunning(t, p, 0, time.Second)
	select {
	case req := <-startReqCh:
		if req.ContextCLIExecutable != "/daemon/owned/riido" {
			t.Fatal("daemon-owned executable did not reach StartRequest")
		}
		if req.Env["RIIDO_RUN_TOKEN"] != "rrc_test" || req.Env["KEEP"] != "1" || req.Env["PATH"] != "/riido/test/bin" {
			t.Fatalf("launch env = %v", req.Env)
		}
	case <-time.After(time.Second):
		t.Fatal("BuildStart was not called")
	}
	if hookRuntimeID != "rt-test" || hookTaskID != "t-hook" {
		t.Fatalf("hook got runtime=%q task=%q", hookRuntimeID, hookTaskID)
	}
	emitCompletedOutput(r)
	expectTaskStatus(t, h.Result(), agentbridge.ResultCompleted, "no result")
}

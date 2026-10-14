package runtimeactor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestRuntimeActorPassesDetectedExecutableToBuildStartAndSpawn(t *testing.T) {
	selected := "/opt/riido/bin/openclaw-supported"
	launchPath := "/riido/test/bin"
	startReqCh := make(chan agentbridge.StartRequest, 1)
	a, p := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			detectedOpenClawAdapter(selected, startReqCh),
		},
		MaxConcurrent: 1,
	})
	h, err := a.Submit(context.Background(), bridge.TaskRequest{
		ID: "t-openclaw", Provider: "openclaw", Prompt: "hi", Env: map[string]string{"PATH": launchPath},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := waitForRunning(t, p, 0, time.Second)

	expectStartRequest(t, startReqCh, selected, launchPath)
	expectSpawnCommand(t, p, selected, launchPath)
	emitCompletedOutput(r)
	expectTaskStatus(t, h.Result(), agentbridge.ResultCompleted, "no result")
}

func detectedOpenClawAdapter(selected string, startReqCh chan agentbridge.StartRequest) agentbridge.Adapter {
	return &stubAdapter{
		name: "openclaw",
		detected: agentbridge.DetectResult{
			Available:  true,
			Executable: selected,
		},
		startReqCh: startReqCh,
	}
}

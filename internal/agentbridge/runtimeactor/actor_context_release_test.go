package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
)

func TestRuntimeActorReleasesContextOnCompletion(t *testing.T) {
	released := make(chan string, 2)
	registry := contextbroker.NewRegistry()
	var issued string
	a, p := startActor(t, Config{
		RuntimeID:            "rt-context-release",
		Adapters:             []agentbridge.Adapter{detectedOpenClawAdapter("/opt/riido/bin/openclaw", make(chan agentbridge.StartRequest, 1))},
		ContextCLIExecutable: "/daemon/riido",
		LaunchEnv: func(_ bridge.TaskRequest, _ string, env map[string]string) {
			var err error
			issued, err = registry.Issue(contextbroker.Binding{RunKey: "release-task"})
			if err != nil {
				t.Error(err)
			}
			env["RIIDO_RUN_TOKEN"] = issued
		},
		ReleaseContextCapability: func(token string) { registry.RevokeCapability(token); released <- token },
	})
	h, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "release-task", Provider: "openclaw", Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	emitCompletedOutput(waitForRunning(t, p, 0, time.Second))
	expectTaskStatus(t, h.Result(), agentbridge.ResultCompleted, "no result")
	select {
	case token := <-released:
		if token != issued {
			t.Fatal("wrong capability released")
		}
	case <-time.After(time.Second):
		t.Fatal("finished session retained context capability")
	}
	if registry.Len() != 0 {
		t.Fatal("finished runtime retained capability without a followup read")
	}
	select {
	case <-released:
		t.Fatal("capability released twice")
	default:
	}
}

type contextFailedAdapter struct{ agentbridge.Adapter }

func (contextFailedAdapter) BuildStart(agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{}, errors.New("synthetic launch failure")
}

func TestRuntimeActorReleasesContextOnFailedLaunchAndStop(t *testing.T) {
	for _, fail := range []bool{true, false} {
		t.Run(map[bool]string{true: "build", false: "stop"}[fail], func(t *testing.T) {
			released := make(chan string, 2)
			adapter := detectedOpenClawAdapter("/opt/riido/bin/openclaw", nil)
			if fail {
				adapter = contextFailedAdapter{adapter}
			}
			a, p := startActor(t, Config{Adapters: []agentbridge.Adapter{adapter}, ContextCLIExecutable: "/daemon/riido", LaunchEnv: func(_ bridge.TaskRequest, _ string, env map[string]string) { env["RIIDO_RUN_TOKEN"] = "issued" }, ReleaseContextCapability: func(token string) { released <- token }})
			_, err := a.Submit(context.Background(), bridge.TaskRequest{ID: "release", Provider: "openclaw", Prompt: "hello"})
			if fail && err == nil {
				t.Fatal("launch failure missing")
			}
			if !fail {
				if err != nil {
					t.Fatal(err)
				}
				waitForRunning(t, p, 0, time.Second)
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				if err := a.Stop(ctx); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case token := <-released:
				if token != "issued" {
					t.Fatal("wrong capability")
				}
			case <-time.After(time.Second):
				t.Fatal("launch/stop retained capability")
			}
			select {
			case <-released:
				t.Fatal("duplicate release")
			default:
			}
		})
	}
}

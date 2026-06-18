package cursor

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func runIntegrationSession(t *testing.T, ctx context.Context, profile Profile, workdir string) (agentbridge.Result, []agentbridge.Event) {
	t.Helper()
	req := integrationStartRequest(workdir)
	spawn, err := BuildStart(req, StartOptions{Profile: profile})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}
	sess, err := session.Start(ctx, session.Config{
		TaskID:      "integration-cursor",
		RuntimeID:   "cursor",
		Adapter:     integrationAdapter{},
		Process:     processexec.New(),
		Spawn:       process.Command{Executable: spawn.Executable, Args: spawn.Args, Env: spawn.Env, Dir: spawn.Dir},
		Request:     req,
		HardTimeout: 45 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	eventsDone := collectIntegrationEvents(sess)
	return <-sess.Result(), <-eventsDone
}

func collectIntegrationEvents(sess *session.Session) <-chan []agentbridge.Event {
	eventsDone := make(chan []agentbridge.Event, 1)
	go func() {
		var events []agentbridge.Event
		for ev := range sess.Events() {
			events = append(events, ev)
		}
		eventsDone <- events
	}()
	return eventsDone
}

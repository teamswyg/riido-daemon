package openclaw

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func runOpenClawIntegrationSession(
	t *testing.T,
	ctx context.Context,
	req agentbridge.StartRequest,
	sessionID string,
) agentbridge.Result {
	t.Helper()
	spawn, err := BuildStart(req, StartOptions{SessionID: sessionID})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}
	sess, err := startOpenClawIntegrationSession(ctx, spawn, sessionID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	go drainOpenClawIntegrationEvents(sess)
	return <-sess.Result()
}

func startOpenClawIntegrationSession(
	ctx context.Context,
	spawn agentbridge.StartCommand,
	sessionID string,
) (*session.Session, error) {
	return session.Start(ctx, session.Config{
		TaskID:      "integration-openclaw",
		RuntimeID:   "openclaw",
		Adapter:     integrationAdapter{sessionID: sessionID},
		Process:     processexec.New(),
		Spawn:       processCommandFromStart(spawn),
		HardTimeout: 180 * time.Second,
	})
}

func processCommandFromStart(spawn agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: spawn.Executable,
		Args:       spawn.Args,
		Env:        spawn.Env,
		Dir:        spawn.Dir,
	}
}

func drainOpenClawIntegrationEvents(sess *session.Session) {
	for range sess.Events() {
	}
}

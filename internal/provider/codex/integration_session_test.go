package codex

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func runCodexIntegrationSession(
	t *testing.T,
	ctx context.Context,
	req agentbridge.StartRequest,
) agentbridge.Result {
	t.Helper()
	spawn, err := BuildStart(req, StartOptions{})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}
	driver, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatalf("NewProtocolDriver: %v", err)
	}
	sess, err := startCodexIntegrationSession(ctx, req, spawn, driver)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	go drainCodexIntegrationEvents(sess)
	return <-sess.Result()
}

func startCodexIntegrationSession(
	ctx context.Context,
	req agentbridge.StartRequest,
	spawn agentbridge.StartCommand,
	driver agentbridge.ProtocolDriver,
) (*session.Session, error) {
	return session.Start(ctx, session.Config{
		TaskID:         "integration-codex",
		RuntimeID:      "codex",
		Adapter:        integrationAdapter{},
		Process:        processexec.New(),
		Spawn:          processCommandFromStart(spawn),
		Request:        req,
		HardTimeout:    codexIntegrationHardTimeout,
		ProtocolDriver: driver,
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

func drainCodexIntegrationEvents(sess *session.Session) {
	for range sess.Events() {
	}
}

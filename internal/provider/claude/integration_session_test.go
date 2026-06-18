package claude

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func startClaudeIntegrationSession(
	t *testing.T,
	ctx context.Context,
	req agentbridge.StartRequest,
) *session.Session {
	t.Helper()

	spawn, err := BuildStart(req, StartOptions{PermissionMode: PermissionModeAcceptEdits})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}
	driver, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatalf("NewProtocolDriver: %v", err)
	}

	sess, err := session.Start(ctx, session.Config{
		TaskID:         "integration-claude",
		RuntimeID:      "claude",
		Adapter:        integrationAdapter{},
		Process:        processexec.New(),
		Spawn:          claudeIntegrationSpawn(spawn),
		HardTimeout:    45 * time.Second,
		ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	return sess
}

func claudeIntegrationSpawn(spawn agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: spawn.Executable,
		Args:       spawn.Args,
		Env:        spawn.Env,
		Dir:        spawn.Dir,
	}
}

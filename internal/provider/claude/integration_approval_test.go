package claude

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

func TestIntegrationApproval(t *testing.T) {
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("AGENTBRIDGE_INTEGRATION not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	det, err := Detect(ctx, agentbridge.DetectEnv{})
	if err != nil {
		t.Fatal(err)
	}
	if !det.Available {
		t.Skip("Claude executable unavailable")
	}
	path := filepath.Join(t.TempDir(), "approved.txt")
	req := agentbridge.StartRequest{Executable: det.Executable, Cwd: filepath.Dir(path), Prompt: "Use only the Write tool to create " + path + " with exactly the content RIIDO_APPROVAL_OK. Do not use Bash or other tools. Then respond ok."}
	spawn, err := BuildStart(req, StartOptions{PermissionMode: PermissionModeApproval})
	if err != nil {
		t.Fatal(err)
	}
	driver, _ := NewProtocolDriver(req)
	sess, err := session.Start(ctx, session.Config{
		TaskID: "integration-claude-approval", RuntimeID: Name, Adapter: integrationAdapter{},
		Process: processexec.New(), Spawn: process.Command{Executable: spawn.Executable, Args: spawn.Args, Env: spawn.Env, Dir: spawn.Dir},
		Request: req, ProtocolDriver: driver, HardTimeout: 45 * time.Second,
		AutoApprove: func(tool agentbridge.ToolRef) bool { return tool.Name == "Write" && tool.Args["file_path"] == path },
	})
	if err != nil {
		t.Fatal(err)
	}
	approvals := 0
	for event := range sess.Events() {
		if event.Kind == agentbridge.EventToolApprovalNeeded {
			approvals++
		}
	}
	result := <-sess.Result()
	if result.Status != agentbridge.ResultCompleted || approvals == 0 {
		t.Fatalf("approval roundtrip failed: approvals=%d result=%+v", approvals, result)
	}
	body, err := os.ReadFile(path)
	if err != nil || string(body) != "RIIDO_APPROVAL_OK" {
		t.Fatalf("approved artifact=%q err=%v", body, err)
	}
}

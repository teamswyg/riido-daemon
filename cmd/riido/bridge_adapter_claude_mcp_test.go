package main

import (
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
)

func TestDaemonClaudeAdapterAddsPermissionPromptMCPWithoutSecrets(t *testing.T) {
	adapter := bridgeClaudeAdapter{approvalSocket: "/tmp/agentd.sock"}
	cmd, err := adapter.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{
			"RIIDO_DEVICE_SECRET": "must-not-enter-mcp-config",
			"RIIDO_SAAS_URL":      "https://example.invalid",
		},
		Metadata: map[string]string{
			saasplane.MetadataAssignmentID: "asn-1",
			controlplane.MetadataTaskID:    "task-1",
			controlplane.MetadataRuntimeID: "rt-1",
		},
	})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}
	if len(cmd.TempFiles) != 1 {
		t.Fatalf("temp files = %+v, want one MCP config", cmd.TempFiles)
	}
	defer os.Remove(cmd.TempFiles[0])
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--permission-prompt-tool "+claudePermissionPromptToolName) {
		t.Fatalf("missing permission prompt tool: %q", args)
	}
	config, err := os.ReadFile(cmd.TempFiles[0])
	if err != nil {
		t.Fatalf("read MCP config: %v", err)
	}
	text := string(config)
	for _, forbidden := range []string{"must-not-enter-mcp-config", "RIIDO_DEVICE_SECRET", "RIIDO_SAAS_URL"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("MCP config leaked forbidden value %q:\n%s", forbidden, text)
		}
	}
	for _, want := range []string{"claude-permission-mcp", "/tmp/agentd.sock", "asn-1", "task-1", "rt-1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("MCP config missing %q:\n%s", want, text)
		}
	}
}

func TestDaemonClaudeAdapterRejectsIncompletePermissionMCPIdentity(t *testing.T) {
	adapter := bridgeClaudeAdapter{approvalSocket: "/tmp/agentd.sock"}
	_, err := adapter.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Metadata: map[string]string{
			saasplane.MetadataAssignmentID: "asn-1",
			controlplane.MetadataTaskID:    "task-1",
		},
	})
	if err == nil {
		t.Fatal("BuildStart succeeded with missing runtime id")
	}
	if !strings.Contains(err.Error(), "--runtime-id") {
		t.Fatalf("error = %v", err)
	}
}

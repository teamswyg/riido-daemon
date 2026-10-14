package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
)

func TestDaemonClaudeAdapterUsesBetaFullAccessWithoutMCP(t *testing.T) {
	adapter := bridgeClaudeAdapter{approvalSocket: "/tmp/agentd.sock", betaFullAccess: true}
	cmd, err := adapter.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Metadata: map[string]string{
			saasplane.MetadataAssignmentID: "asn-1",
			controlplane.MetadataTaskID:    "task-1",
			controlplane.MetadataRuntimeID: "rt-1",
		},
	})
	if err != nil {
		t.Fatalf("BuildStart: %v", err)
	}
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, "--permission-mode bypassPermissions") {
		t.Fatalf("missing beta full-access mode: %q", args)
	}
	if strings.Contains(args, "--permission-prompt-tool") ||
		strings.Contains(args, "--mcp-config") ||
		len(cmd.TempFiles) != 0 {
		t.Fatalf("beta full-access must not wait on approval MCP: args=%q temp=%v", args, cmd.TempFiles)
	}
}

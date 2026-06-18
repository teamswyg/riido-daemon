package claude

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartMCPGuard(t *testing.T) {
	bare, _ := BuildStart(agentbridge.StartRequest{}, safeStartOptions())
	if strings.Contains(strings.Join(bare.Args, " "), "--strict-mcp-config") {
		t.Fatalf("--strict-mcp-config must not be set without --mcp-config: %v", bare.Args)
	}

	with, _ := BuildStart(agentbridge.StartRequest{}, StartOptions{
		PermissionMode: PermissionModeApproval,
		MCPConfigPath:  "/tmp/mcp.json",
	})
	args := strings.Join(with.Args, " ")
	if !strings.Contains(args, "--strict-mcp-config") ||
		!strings.Contains(args, "--mcp-config /tmp/mcp.json") {
		t.Fatalf("missing strict-mcp-config or --mcp-config when path provided: %q", args)
	}
	if !slices.Contains(with.TempFiles, "/tmp/mcp.json") {
		t.Fatalf("MCP config path should be registered as a temp file, got %v", with.TempFiles)
	}
}

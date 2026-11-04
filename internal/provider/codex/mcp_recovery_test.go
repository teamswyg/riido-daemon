package codex

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestMCPRecoveryLive(t *testing.T) {
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" {
		t.Skip("requires installed Codex CLI")
	}
	for _, tc := range []struct {
		name, want string
		offline    bool
	}{{"expired", "codex mcp login docs", false}, {"offline", "서버의 실행 상태", true}} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("WWW-Authenticate", "Bearer")
				w.WriteHeader(http.StatusUnauthorized)
			}))
			defer server.Close()
			if tc.offline {
				server.Close()
			}
			home := t.TempDir()
			config := fmt.Sprintf("[mcp_servers.docs]\nrequired = true\nurl = %q\nstartup_timeout_sec = 2\n", server.URL+"/mcp")
			if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			res := runCodexIntegrationSession(t, ctx, agentbridge.StartRequest{Prompt: "Read docs", Cwd: t.TempDir(), Env: map[string]string{"CODEX_HOME": home, "PATH": os.Getenv("PATH"), "HOME": os.Getenv("HOME")}})
			if !res.RequiresUserAction || res.Status != agentbridge.ResultBlocked || !strings.Contains(res.Output, tc.want) || strings.Contains(res.Output, server.URL) {
				t.Fatalf("unexpected recovery: %+v", res)
			}
		})
	}
}

func TestMCPRecoveryScope(t *testing.T) {
	d := protocolDriver{threadID: "mine"}
	p := map[string]any{"name": "docs", "failureReason": "reauthenticationRequired", "status": "failed"}
	ev := d.onMCPStartup(p)
	if len(ev) != 1 || ev[0].Kind != agentbridge.EventProgress || !strings.Contains(ev[0].Text, "codex mcp login docs") || len(d.onMCPStartup(p)) != 0 {
		t.Fatal(ev)
	}
	p = map[string]any{"threadId": "mine", "item": map[string]any{"type": "mcpToolCall", "server": "docs", "status": "failed", "error": map[string]any{"message": "connection refused"}}}
	if e := d.mcpToolRecovery(p); len(e) != 1 || !e[0].Result.RequiresUserAction {
		t.Fatal(e)
	}
	p["threadId"] = "another"
	if e := d.mcpToolRecovery(p); len(e) != 0 {
		t.Fatal(e)
	}
	if m := recoveryAdvice("docs;bad", "reauthenticationRequired", ""); m != "" {
		t.Fatal(m)
	}
}

func TestMCPRecoveryMultipleServers(t *testing.T) {
	reply := requiredMCPRecovery("required MCP servers failed to initialize: local: connection refused; docs: Auth required")
	if !strings.Contains(reply, "local MCP에 연결") || !strings.Contains(reply, "codex mcp login docs") || strings.Contains(reply, "login local") {
		t.Fatal(reply)
	}
}

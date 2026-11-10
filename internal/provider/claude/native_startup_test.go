package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process/processexec"
)

// Opt-in native CLI transport test. No user configuration or paid API is used.
// On Windows processexec applies the same CREATE_NO_WINDOW flags as the daemon.
func TestClaudeStartupNativeCLI(t *testing.T) {
	executable := os.Getenv("RIIDO_TEST_CLAUDE_EXECUTABLE")
	if executable == "" {
		t.Skip("set RIIDO_TEST_CLAUDE_EXECUTABLE to a native Claude binary")
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			http.NotFound(w, r)
			return
		}
		calls.Add(1)
		var request struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		body, err := os.ReadFile("testdata/native-startup.sse")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Fprint(w, strings.ReplaceAll(string(body), "MODEL", request.Model))
	}))
	defer server.Close()
	dir := t.TempDir()
	config := filepath.Join(dir, "config")
	if err := os.Mkdir(config, 0700); err != nil {
		t.Fatal(err)
	}
	mcp := filepath.Join(dir, "mcp.json")
	if err := os.WriteFile(mcp, []byte(`{"mcpServers":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	req := agentbridge.StartRequest{
		Executable: executable, Cwd: dir, Prompt: "Reply with OK only.",
		Env: map[string]string{
			"CLAUDE_CONFIG_DIR":                        config,
			"ANTHROPIC_API_KEY":                        "riido-local-test-only",
			"ANTHROPIC_BASE_URL":                       server.URL,
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
			"DISABLE_AUTOUPDATER":                      "1",
			"CLAUDECODE":                               "", "CLAUDE_CODE_OAUTH_TOKEN": "",
		},
	}
	spawn, err := BuildStart(req, StartOptions{PermissionMode: PermissionModeApproval, MCPConfigPath: mcp})
	if err != nil {
		t.Fatal(err)
	}
	driver, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	sess, err := session.Start(ctx, session.Config{
		Adapter: integrationAdapter{}, Process: processexec.New(),
		Spawn: claudeIntegrationSpawn(spawn), ProtocolDriver: driver,
	})
	if err != nil {
		t.Fatal(err)
	}
	var logs []string
	for event := range sess.Events() {
		if event.Kind == agentbridge.EventLog || event.Kind == agentbridge.EventError {
			logs = append(logs, event.Text+event.Err)
		}
	}
	result := <-sess.Result()
	if result.Status != agentbridge.ResultCompleted || strings.TrimSpace(result.Output) != "OK" || calls.Load() == 0 {
		t.Fatalf("native CLI status=%s output=%q error=%q calls=%d logs=%v", result.Status, result.Output, result.Error, calls.Load(), logs)
	}
}

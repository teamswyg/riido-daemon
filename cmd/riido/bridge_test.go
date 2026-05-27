package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/supervisor"
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

// captureStdout redirects os.Stdout for the duration of fn and returns
// what was written.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	_ = w.Close()
	<-done
	os.Stdout = orig
	return buf.String()
}

func TestBridgeProvidersListsAllFour(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run([]string{"bridge", "providers"}); err != nil {
			t.Fatalf("run: %v", err)
		}
	})

	var listing struct {
		SchemaVersion string `json:"schema_version"`
		Providers     []struct {
			Name        string   `json:"name"`
			BlockedArgs []string `json:"blocked_args"`
		} `json:"providers"`
	}
	if err := json.Unmarshal([]byte(out), &listing); err != nil {
		t.Fatalf("parse JSON %q: %v", out, err)
	}
	if listing.SchemaVersion == "" {
		t.Fatalf("schema version missing: %s", out)
	}
	want := map[string]bool{"claude": false, "codex": false, "openclaw": false, "cursor": false}
	for _, p := range listing.Providers {
		if _, ok := want[p.Name]; ok {
			want[p.Name] = true
		}
		if len(p.BlockedArgs) == 0 {
			t.Fatalf("provider %s has no blocked args", p.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("provider %s not listed in %v", name, listing.Providers)
		}
	}
}

func TestBridgeDetectIncludesEachProvider(t *testing.T) {
	out := captureStdout(t, func() {
		if err := run([]string{"bridge", "detect"}); err != nil {
			t.Fatalf("run: %v", err)
		}
	})
	for _, want := range []string{"claude", "codex", "openclaw", "cursor"} {
		if !strings.Contains(out, `"`+want+`"`) {
			t.Fatalf("detect output missing %s: %s", want, out)
		}
	}
}

func TestRegisteredAdaptersBuildStartForDaemonRuntime(t *testing.T) {
	for _, adapter := range registeredAdapters() {
		cmd, err := adapter.BuildStart(agentbridge.StartRequest{
			TaskID: "task-" + adapter.Name(),
			Prompt: "do the thing",
			Cwd:    "/tmp/work",
		})
		if err != nil {
			t.Fatalf("%s BuildStart: %v", adapter.Name(), err)
		}
		if cmd.Executable == "" {
			t.Fatalf("%s executable empty", adapter.Name())
		}
		switch adapter.Name() {
		case claude.Name:
			args := strings.Join(cmd.Args, " ")
			if !strings.Contains(args, "--permission-mode default") {
				t.Fatalf("claude daemon adapter must use approval mode, got %q", args)
			}
			if strings.Contains(args, "bypassPermissions") {
				t.Fatalf("claude daemon adapter must not default to bypassPermissions: %q", args)
			}
		case codex.Name:
			for _, env := range cmd.Env {
				if strings.HasPrefix(env, "CODEX_HOME=") {
					t.Fatalf("codex adapter must not invent CODEX_HOME without supervisor workdir metadata: %v", cmd.Env)
				}
			}
		case openclaw.Name:
			assertBridgeArgPair(t, cmd.Args, "--session-id", "task-openclaw")
		case cursor.Name:
			if strings.Contains(strings.Join(cmd.Args, " "), "--yolo") {
				t.Fatalf("cursor daemon adapter must not default to --yolo: %v", cmd.Args)
			}
		}
	}
}

func TestCodexDaemonAdapterUsesTaskScopedCodexHome(t *testing.T) {
	cmd, err := bridgeCodexAdapter{}.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Metadata: map[string]string{
			supervisor.MetadataNativeConfigHome: "/tmp/work/.codex",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsEnv(cmd.Env, "CODEX_HOME=/tmp/work/.codex") {
		t.Fatalf("codex adapter did not set task-scoped CODEX_HOME: %v", cmd.Env)
	}
}

func TestCodexDaemonAdapterDoesNotInferTaskScopedCodexHome(t *testing.T) {
	cmd, err := bridgeCodexAdapter{}.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Metadata: map[string]string{
			supervisor.MetadataWorkdir: "/tmp/work",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "CODEX_HOME=") {
			t.Fatalf("codex adapter must wait for explicit native config home metadata: %v", cmd.Env)
		}
	}
}

func assertBridgeArgPair(t *testing.T, args []string, key string, value string) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}
	t.Fatalf("missing arg pair %s %s in %v", key, value, args)
}

func containsEnv(env []string, want string) bool {
	for _, value := range env {
		if value == want {
			return true
		}
	}
	return false
}

func TestBridgeUsageOnUnknownSubcommand(t *testing.T) {
	err := run([]string{"bridge", "nonsense"})
	if err == nil {
		t.Fatal("expected error for unknown bridge subcommand")
	}
}

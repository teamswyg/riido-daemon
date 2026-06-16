package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
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
	for _, adapter := range builtinAgentAdapters() {
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
			args := strings.Join(cmd.Args, " ")
			assertBridgeArgPair(t, cmd.Args, "--sandbox", codex.FullAccessSandboxMode)
			if strings.Contains(args, "default_permissions") || strings.Contains(args, "permissions.riido-task") {
				t.Fatalf("codex adapter must not generate a task-scoped permission profile: %q", args)
			}
			for _, env := range cmd.Env {
				if strings.HasPrefix(env, "CODEX_HOME=") {
					t.Fatalf("codex adapter must not invent CODEX_HOME: %v", cmd.Env)
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

func TestCodexDaemonAdapterPreservesConfiguredCodexHomeWithoutPermissionProfile(t *testing.T) {
	cmd, err := bridgeCodexAdapter{}.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"CODEX_HOME": "/Users/example/.codex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	assertBridgeArgPair(t, cmd.Args, "--sandbox", codex.FullAccessSandboxMode)
	for _, bad := range []string{`"/Users/example/.codex"="none"`, "default_permissions", "permissions.riido-task"} {
		if strings.Contains(args, bad) {
			t.Fatalf("codex adapter must not generate permission profile token %q: %q", bad, args)
		}
	}
	if !containsEnv(cmd.Env, "CODEX_HOME=/Users/example/.codex") {
		t.Fatalf("codex adapter should preserve caller CODEX_HOME for app-server auth: %v", cmd.Env)
	}
}

func TestCodexDaemonAdapterDoesNotDeriveDefaultCodexHomeFromHome(t *testing.T) {
	cmd, err := bridgeCodexAdapter{}.BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/work",
		Env: map[string]string{"HOME": "/Users/example"},
	})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	if strings.Contains(args, "/Users/example/.codex") || strings.Contains(args, "default_permissions") {
		t.Fatalf("codex adapter must not derive auth home permission profile: %q", args)
	}
	for _, env := range cmd.Env {
		if strings.HasPrefix(env, "CODEX_HOME=") {
			t.Fatalf("codex adapter must not invent CODEX_HOME: %v", cmd.Env)
		}
	}
}

func assertBridgeArgPair(t *testing.T, args []string, key, value string) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}
	t.Fatalf("missing arg pair %s %s in %v", key, value, args)
}

func containsEnv(env []string, want string) bool {
	return slices.Contains(env, want)
}

func TestBridgeUsageOnUnknownSubcommand(t *testing.T) {
	err := run([]string{"bridge", "nonsense"})
	if err == nil {
		t.Fatal("expected error for unknown bridge subcommand")
	}
}

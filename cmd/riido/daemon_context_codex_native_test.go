package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// Exercise the real Codex command environment, compiled CLI and local broker.
// Only upstream context data is synthetic; no model or user account is used.
func TestIntegrationCodexContextCLIReadBinding(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this native shell fixture uses POSIX commands")
	}
	exe := os.Getenv("RIIDO_CODEX_NATIVE_TEST_EXECUTABLE")
	if exe == "" {
		t.Skip("set RIIDO_CODEX_NATIVE_TEST_EXECUTABLE to a Codex binary")
	}
	f := newNativeContextFixture(t)
	home := t.TempDir()
	config := "[shell_environment_policy]\ninherit = 'core'\nignore_default_excludes = false\n[shell_environment_policy.set]\nUNRELATED_EXPLICIT = 'preserved'\n"
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	req := agentbridge.StartRequest{Executable: exe, Cwd: home, ContextCLIEnabled: true, Env: map[string]string{
		contextcli.EnvRunToken: f.token, contextcli.EnvDaemonSocket: f.socket,
		contextcli.EnvTaskID: "RIID-5900", contextcli.EnvWorkspaceID: "native-ws",
	}}
	var secrets []string
	for _, resume := range []bool{false, true} {
		if resume {
			f.broker.Registry.RevokeCapability(f.token)
			var err error
			f.token, err = f.broker.Registry.Issue(contextbroker.Binding{RunKey: "native-run", TaskID: "RIID-5900", WorkspaceID: "native-ws", RuntimeID: "synthetic-runtime", AgentID: "synthetic-agent", AssignmentID: "synthetic-assignment", LeaseToken: "synthetic-lease-private"})
			if err != nil {
				t.Fatal("rotate fixture capability")
			}
			req.Env[contextcli.EnvRunToken] = f.token
		}
		secrets = append(secrets, f.token)
		client := newCodexContextClient(t, exe, home)
		method, params := codexContextThreadRequest(t, req)
		result := client.call(t, method, params)
		thread := result["thread"].(map[string]any)["id"].(string)
		if resume && thread != req.ResumeSessionID {
			t.Fatal("resume changed thread identity")
		}
		for _, args := range []string{"context get", "work get RIID-5900", "work comments RIID-5900", "work relations RIID-5900"} {
			out := client.shell(t, thread, quoteCodexContextCLI(f.binary)+" "+args+" --json --no-input")
			f.assertNoSecrets(t, out)
			var envelope contextcli.Envelope
			if json.Unmarshal([]byte(out), &envelope) != nil || envelope.Error != nil || len(envelope.Sources) != 1 {
				t.Fatalf("resume=%v command=%s did not read the assigned context", resume, args)
			}
			data, _ := json.Marshal(envelope.Data)
			if !strings.Contains(string(data), "NATIVE_CONTEXT_FIXTURE") {
				t.Fatal("CLI did not receive captured source data")
			}
		}
		if out := client.shell(t, thread, "test \"$UNRELATED_EXPLICIT\" = preserved && test -z \"${UNRELATED_INHERITED:-}\" && echo policy-preserved"); !strings.Contains(out, "policy-preserved") {
			t.Fatal("run binding changed unrelated user environment policy")
		}
		req.ResumeSessionID = thread
		client.stop()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.queries) != 8 {
		t.Fatal("expected context, body, comments and relations reads on both start and resume")
	}
	for _, query := range f.queries {
		if query.AssignmentID != "synthetic-assignment" || query.LeaseToken != "synthetic-lease-private" {
			t.Fatal("read escaped the current assignment binding")
		}
	}
	assertCodexContextSecretsNotPersisted(t, home, secrets)
}

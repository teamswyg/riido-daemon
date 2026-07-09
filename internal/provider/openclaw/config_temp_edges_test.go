package openclaw

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartSkipsTaskScopedConfigWithoutWorkspace(t *testing.T) {
	source := writeOpenClawConfigFixture(t, "/old/workspace", "ollama/slow")
	cmd, err := BuildStart(agentbridge.StartRequest{
		Env: map[string]string{openClawConfigPathEnv: source},
	}, StartOptions{SessionID: "sess-no-workspace"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd.TempFiles) != 0 {
		t.Fatalf("TempFiles=%v, want none", cmd.TempFiles)
	}
	if got := envValueFromList(cmd.Env, openClawConfigPathEnv); got != source {
		t.Fatalf("config path=%q, want source %q", got, source)
	}
}

func TestBuildStartSkipsMissingTaskScopedConfig(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-openclaw.json")
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/riido-work",
		Env: map[string]string{openClawConfigPathEnv: missing},
	}, StartOptions{SessionID: "sess-missing-config"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cmd.TempFiles) != 0 {
		t.Fatalf("TempFiles=%v, want none", cmd.TempFiles)
	}
}

func TestBuildStartRejectsInvalidTaskScopedConfig(t *testing.T) {
	source := filepath.Join(t.TempDir(), "openclaw.json")
	if err := os.WriteFile(source, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := BuildStart(agentbridge.StartRequest{
		Cwd: "/tmp/riido-work",
		Env: map[string]string{openClawConfigPathEnv: source},
	}, StartOptions{SessionID: "sess-bad-config"})
	if err == nil || !strings.Contains(err.Error(), "parse config") {
		t.Fatalf("expected parse config error, got %v", err)
	}
}

func TestSourceConfigPathUsesStateDir(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "state")
	got, ok := sourceConfigPath(map[string]string{openClawStateDirEnv: stateDir})
	if !ok {
		t.Fatalf("expected config path")
	}
	if got != filepath.Join(stateDir, "openclaw.json") {
		t.Fatalf("source config path=%q", got)
	}
}

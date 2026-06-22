package openclaw

import (
	"os"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestBuildStartWritesTaskScopedConfig(t *testing.T) {
	source := writeOpenClawConfigFixture(t, "/old/workspace", "ollama/slow")
	cmd, err := BuildStart(agentbridge.StartRequest{
		Cwd:   "/tmp/riido-work",
		Model: "ollama/fast",
		Env:   map[string]string{openClawConfigPathEnv: source},
	}, StartOptions{SessionID: "sess-config"})
	if err != nil {
		t.Fatal(err)
	}
	configPath := envValueFromList(cmd.Env, openClawConfigPathEnv)
	if configPath == "" || configPath == source {
		t.Fatalf("task-scoped config path = %q source=%q", configPath, source)
	}
	t.Cleanup(func() { _ = os.Remove(configPath) })
	assertTaskScopedConfig(t, configPath, "/tmp/riido-work", "ollama/fast")
	if len(cmd.TempFiles) != 1 || cmd.TempFiles[0] != configPath {
		t.Fatalf("TempFiles=%v, want [%s]", cmd.TempFiles, configPath)
	}
}

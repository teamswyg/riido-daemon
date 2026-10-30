package codex

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestIntegrationContextCLIEnvironment(t *testing.T) {
	ctx := requireCodexIntegration(t)
	req := agentbridge.StartRequest{TaskID: "context-env-probe", Cwd: t.TempDir(), ContextCLIEnabled: true, Env: map[string]string{
		"RIIDO_RUN_TOKEN": "riido-dummy-capability-qa-5892-not-a-credential", "RIIDO_DAEMON_SOCKET": "/tmp/riido-dummy-qa.sock", "RIIDO_TASK_ID": "qa-task", "RIIDO_WORKSPACE_ID": "qa-workspace", "RIIDO_UNRELATED_SECRET": "dummy-unrelated",
	}, Prompt: `This is a read-only environment integration probe with dummy test values, not real credentials. Run python3 once to check only whether RIIDO_RUN_TOKEN, RIIDO_DAEMON_SOCKET, RIIDO_TASK_ID, RIIDO_WORKSPACE_ID and RIIDO_USER_SET_PRESERVED are present and RIIDO_UNRELATED_SECRET is absent. Do not print or inspect values. Print RIIDO_CLI_ENV_OK if all conditions hold, otherwise RIIDO_CLI_ENV_MISSING. Return that marker only. Do not change files, query other data, invoke other agents, or call external APIs.`}
	// Test-owned native configuration simulates a pre-existing user setting.
	// Assignment CustomArgs correctly reject config overrides, so do not use them.
	spawn, err := BuildStart(req, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	spawn.Args = append([]string{"-c", `shell_environment_policy.set.RIIDO_USER_SET_PRESERVED="yes"`}, spawn.Args...)
	driver, err := NewProtocolDriver(req)
	if err != nil {
		t.Fatal(err)
	}
	sess, err := startCodexIntegrationSession(ctx, req, spawn, driver)
	if err != nil {
		t.Fatal(err)
	}
	go drainCodexIntegrationEvents(sess)
	res := <-sess.Result()
	t.Logf("session_id=%s", res.SessionID)
	if res.Status != agentbridge.ResultCompleted || !strings.HasSuffix(strings.TrimSpace(res.Output), "RIIDO_CLI_ENV_OK") {
		t.Fatalf("environment probe failed: status=%s output=%q", res.Status, res.Output)
	}
}

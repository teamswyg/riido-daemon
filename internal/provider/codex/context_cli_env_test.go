package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestContextCLIEnvironmentReachesStartAndResumeWithoutBroadInheritance(t *testing.T) {
	for _, resume := range []string{"", "thread-existing"} {
		req := agentbridge.StartRequest{ContextCLIEnabled: true, ResumeSessionID: resume, Env: map[string]string{"RIIDO_RUN_TOKEN": "test-run-capability", "RIIDO_DAEMON_SOCKET": "/tmp/test.sock", "RIIDO_TASK_ID": "task", "RIIDO_WORKSPACE_ID": "workspace", "AWS_SECRET_ACCESS_KEY": "must-not-forward", "PATH": "must-not-override"}}
		d := &protocolDriver{req: req}
		_, params := d.threadStartRequest()
		config, ok := params["config"].(map[string]any)
		if !ok {
			t.Fatal("missing per-run CLI environment configuration")
		}
		if len(config) != 4 || config["shell_environment_policy.set.RIIDO_RUN_TOKEN"] != req.Env["RIIDO_RUN_TOKEN"] || config["shell_environment_policy.set.RIIDO_TASK_ID"] != "task" {
			t.Fatal("CLI variables missing or unrelated settings overridden")
		}
		req.ContextCLIEnabled = false
		d.req = req
		_, params = d.threadStartRequest()
		if _, ok := params["config"]; ok {
			t.Fatal("disabled CLI changed shell environment")
		}
	}
}

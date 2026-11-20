package codex

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestThreadContextCLIEnvironmentSurvivesShellPolicy(t *testing.T) {
	for _, resume := range []string{"", "existing-thread"} {
		t.Run("resume="+resume, func(t *testing.T) {
			req := agentbridge.StartRequest{ContextCLIEnabled: true, ResumeSessionID: resume, Env: map[string]string{
				"RIIDO_RUN_TOKEN": "synthetic-current-capability", "RIIDO_DAEMON_SOCKET": "/tmp/fixture.sock",
				"RIIDO_TASK_ID": "task-current", "RIIDO_WORKSPACE_ID": "workspace-current",
				"RIIDO_DEVICE_SECRET": "must-not-forward", "UNRELATED": "must-not-forward",
			}}
			d := &protocolDriver{req: req}
			_, params := d.threadStartRequest()
			wire, err := json.Marshal(params["config"])
			if err != nil {
				t.Fatal(err)
			}
			var config map[string]map[string]string
			if json.Unmarshal(wire, &config) != nil || len(config) != 1 {
				t.Fatal("run binding is missing from the thread-scoped shell override")
			}
			values := config["shell_environment_policy.set"]
			if len(values) != 4 {
				t.Fatal("must forward exactly the four run binding fields")
			}
			for _, key := range []string{"RIIDO_RUN_TOKEN", "RIIDO_DAEMON_SOCKET", "RIIDO_TASK_ID", "RIIDO_WORKSPACE_ID"} {
				if values[key] != req.Env[key] {
					t.Fatal("current run binding changed")
				}
			}
			spawn, _ := BuildStart(req, StartOptions{})
			if strings.Contains(strings.Join(spawn.Args, " "), req.Env["RIIDO_RUN_TOKEN"]) {
				t.Fatal("capability must not be a process argument")
			}
		})
	}
}

func TestThreadContextCLIEnvironmentClearsAbsentWorkspace(t *testing.T) {
	req := agentbridge.StartRequest{ContextCLIEnabled: true, ResumeSessionID: "existing-thread", Env: map[string]string{
		"RIIDO_RUN_TOKEN": "current", "RIIDO_DAEMON_SOCKET": "/tmp/current.sock", "RIIDO_TASK_ID": "current-task",
	}}
	_, params := (&protocolDriver{req: req}).threadStartRequest()
	values := params["config"].(map[string]any)["shell_environment_policy.set"].(map[string]string)
	if value, ok := values["RIIDO_WORKSPACE_ID"]; !ok || value != "" {
		t.Fatal("resume would preserve a previous workspace override")
	}
}

func TestThreadContextCLIEnvironmentRequiresCompleteEnabledBinding(t *testing.T) {
	for _, missing := range []string{"disabled", "RIIDO_RUN_TOKEN", "RIIDO_DAEMON_SOCKET", "RIIDO_TASK_ID"} {
		t.Run(missing, func(t *testing.T) {
			req := agentbridge.StartRequest{ContextCLIEnabled: missing != "disabled", Env: map[string]string{
				"RIIDO_RUN_TOKEN": "fixture", "RIIDO_DAEMON_SOCKET": "/tmp/fixture", "RIIDO_TASK_ID": "task",
			}}
			delete(req.Env, missing)
			_, params := (&protocolDriver{req: req}).threadStartRequest()
			if _, present := params["config"]; present {
				t.Fatal("inactive CLI wiring changed the user's command environment")
			}
		})
	}
}

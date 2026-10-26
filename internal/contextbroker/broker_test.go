package contextbroker

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func issueForTest(t *testing.T, b *Broker, binding Binding) string {
	t.Helper()
	env := map[string]string{"PATH": "/usr/bin"}
	if err := b.LaunchEnv(binding, env); err != nil {
		t.Fatalf("LaunchEnv: %v", err)
	}
	token := env[contextcli.EnvRunToken]
	if !strings.HasPrefix(token, tokenPrefix) || len(token) < 40 {
		t.Fatalf("weak or missing token: %q", token)
	}
	return token
}

func request(token, command string) contextcli.ContextReadRequest {
	return contextcli.ContextReadRequest{
		Method: contextcli.IPCMethodContextRead, SchemaVersion: contextcli.ContextIPCSchemaVersion,
		Command: command, RunCapability: token, RequestID: "cli_test",
	}
}

func activeSet(keys ...string) RunActive {
	return func(runKey string) bool {
		for _, key := range keys {
			if key == runKey {
				return true
			}
		}
		return false
	}
}

func errorCode(t *testing.T, env contextcli.Envelope) contextcli.ErrorCode {
	t.Helper()
	if env.Error == nil {
		return ""
	}
	return env.Error.Code
}

func TestLaunchEnvInjectsCLIWiring(t *testing.T) {
	b := New("/tmp/agentd.sock", "/opt/riido/bin")
	env := map[string]string{"PATH": "/usr/bin"}
	if err := b.LaunchEnv(Binding{RunKey: "run-1", TaskID: "task-1", WorkspaceID: "ws-1"}, env); err != nil {
		t.Fatal(err)
	}
	sep := string(os.PathListSeparator)
	if env[contextcli.EnvDaemonSocket] != "/tmp/agentd.sock" || env[contextcli.EnvTaskID] != "task-1" ||
		env[contextcli.EnvWorkspaceID] != "ws-1" || env["PATH"] != "/opt/riido/bin"+sep+"/usr/bin" {
		t.Fatalf("launch env = %v", env)
	}
	again := map[string]string{"Path": "/opt/riido/bin" + sep + "/usr/bin"}
	prependPATH(again, "/opt/riido/bin")
	if again["Path"] != "/opt/riido/bin"+sep+"/usr/bin" || len(again) != 1 {
		t.Fatalf("PATH must not duplicate and must keep Windows spelling: %v", again)
	}
}

func TestBrokerRejectsMissingUnknownAndEndedCapabilities(t *testing.T) {
	b := New("", "")
	token := issueForTest(t, b, Binding{RunKey: "run-1"})
	cases := map[string]contextcli.Envelope{
		"missing": b.Handle(request("", "whoami"), activeSet("run-1")),
		"unknown": b.Handle(request("rrc_forged", "whoami"), activeSet("run-1")),
		"ended":   b.Handle(request(token, "whoami"), activeSet()),
	}
	for name, env := range cases {
		if errorCode(t, env) != contextcli.CodePermissionDenied {
			t.Fatalf("%s: want PERMISSION_DENIED, got %+v", name, env.Error)
		}
	}
	if b.Registry.Len() != 0 {
		t.Fatal("an ended run must revoke its capability")
	}
	if got := b.Handle(request(token, "whoami"), activeSet("run-1")); errorCode(t, got) != contextcli.CodePermissionDenied {
		t.Fatal("a revoked capability must stay rejected even if the run key reappears")
	}
}

func TestBrokerEnforcesRunScope(t *testing.T) {
	b := New("", "")
	token := issueForTest(t, b, Binding{RunKey: "run-1", TaskID: "task-1", WorkspaceID: "ws-1"})
	req := request(token, "whoami")
	req.WorkspaceID = "ws-other"
	if got := b.Handle(req, activeSet("run-1")); errorCode(t, got) != contextcli.CodeOutOfRunScope {
		t.Fatalf("workspace mismatch: %+v", got.Error)
	}
	req = request(token, "whoami")
	req.TaskID = "task-other"
	if got := b.Handle(req, activeSet("run-1")); errorCode(t, got) != contextcli.CodeOutOfRunScope {
		t.Fatalf("task mismatch: %+v", got.Error)
	}
}

func TestBrokerAnswersWithinRunAndHoldsDataReads(t *testing.T) {
	b := New("", "")
	token := issueForTest(t, b, Binding{RunKey: "run-1", TaskID: "task-1", WorkspaceID: "ws-1", AgentID: "agent-1"})
	active := activeSet("run-1")
	who := b.Handle(request(token, "whoami"), active)
	if who.Error != nil || len(who.Warnings) != 1 || who.Warnings[0].Code != "REQUESTER_UNRESOLVED" {
		t.Fatalf("whoami: %+v", who)
	}
	body, err := json.Marshal(who)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), token) || !strings.Contains(string(body), `"agent_id":"agent-1"`) {
		t.Fatalf("whoami body leaked the token or missed run facts: %s", body)
	}
	if caps := b.Handle(request(token, "capabilities"), active); caps.Error != nil {
		t.Fatalf("capabilities probe: %+v", caps.Error)
	}
	read := b.Handle(request(token, "work get"), active)
	if errorCode(t, read) != contextcli.CodeControlPlaneUnsupported || read.Error.NextAction == "" {
		t.Fatalf("data read must wait for the control-plane grant: %+v", read.Error)
	}
}

func TestReissueReplacesEarlierCapability(t *testing.T) {
	b := New("", "")
	first := issueForTest(t, b, Binding{RunKey: "run-1"})
	second := issueForTest(t, b, Binding{RunKey: "run-1"})
	if first == second || b.Registry.Len() != 1 {
		t.Fatalf("reissue must rotate the token (len=%d)", b.Registry.Len())
	}
	if got := b.Handle(request(first, "whoami"), activeSet("run-1")); errorCode(t, got) != contextcli.CodePermissionDenied {
		t.Fatal("the replaced capability must be rejected")
	}
}

func TestBrokerRejectsUnknownSchemaAndNilActive(t *testing.T) {
	b := New("", "")
	token := issueForTest(t, b, Binding{RunKey: "run-1"})
	req := request(token, "whoami")
	req.SchemaVersion = "2"
	if got := b.Handle(req, activeSet("run-1")); errorCode(t, got) != contextcli.CodeUnsupported {
		t.Fatalf("schema: %+v", got.Error)
	}
	if got := b.Handle(request(token, "whoami"), nil); errorCode(t, got) != contextcli.CodePermissionDenied {
		t.Fatal("without an activity check the broker must fail closed")
	}
}

func TestRegistryEvictsOldestWhenFull(t *testing.T) {
	r := NewRegistry()
	for i := range maxBindings + 1 {
		if _, err := r.Issue(Binding{RunKey: "run-" + strconv.Itoa(i)}); err != nil {
			t.Fatal(err)
		}
	}
	if r.Len() != maxBindings {
		t.Fatalf("registry size = %d, want %d", r.Len(), maxBindings)
	}
}

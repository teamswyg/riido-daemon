package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/logging"
)

func TestParseDaemonContextCLI(t *testing.T) {
	cases := map[string]bool{"": false, "on": true, "OFF": false, "1": true, "true": true, "0": false}
	for raw, want := range cases {
		got, err := parseDaemonContextCLI(raw)
		if err != nil || got != want {
			t.Fatalf("%q: got %v err=%v", raw, got, err)
		}
	}
	if _, err := parseDaemonContextCLI("maybe"); err == nil {
		t.Fatal("invalid value must fail daemon config")
	}
}

func TestContextLaunchEnvOffLeavesLaunchUnchanged(t *testing.T) {
	if daemonContextLaunchEnv(nil) != nil {
		t.Fatal("feature off must not install a launch hook")
	}
}

func TestContextLaunchEnvUsesRunMetadata(t *testing.T) {
	broker := contextbroker.New("/tmp/agentd.sock", "/opt/riido/bin")
	hook := daemonContextLaunchEnv(broker)
	env := map[string]string{"PATH": "/usr/bin"}
	hook(bridge.TaskRequest{ID: "exec-1", Metadata: map[string]string{
		controlplane.MetadataTaskID:    "ArugNoqPo88Zb4Qj2_Tq9",
		"workspace_id":                 "ws-1",
		saasplane.MetadataAgentID:      "agent-1",
		saasplane.MetadataAssignmentID: "asg-1",
	}}, "device-1:codex", env)
	binding, ok := broker.Registry.Lookup(env[contextcli.EnvRunToken])
	if !ok || binding.RunKey != "exec-1" || binding.TaskID != "ArugNoqPo88Zb4Qj2_Tq9" ||
		binding.WorkspaceID != "ws-1" || binding.RuntimeID != "device-1:codex" || binding.AssignmentID != "asg-1" {
		t.Fatalf("binding=%+v ok=%v env=%v", binding, ok, env)
	}
	if env[contextcli.EnvTaskID] != "ArugNoqPo88Zb4Qj2_Tq9" || env[contextcli.EnvDaemonSocket] != "/tmp/agentd.sock" {
		t.Fatalf("env=%v", env)
	}
}

// brokerDial serves each CLI connection with the real IPC decode and
// writeContextRead path.
func brokerDial(settings daemonSettings) func(string, time.Duration) (net.Conn, error) {
	return func(string, time.Duration) (net.Conn, error) {
		client, server := net.Pipe()
		go func() {
			defer server.Close()
			req, ok, err := readDaemonRequest(server)
			if err != nil || !ok {
				return
			}
			writeContextRead(server, req, settings, nil, logging.NewWriterLogger(io.Discard))
		}()
		return client, nil
	}
}

func TestCLIAgainstDaemonBrokerEndToEnd(t *testing.T) {
	broker := contextbroker.New("/tmp/agentd.sock", "")
	env := map[string]string{}
	hook := daemonContextLaunchEnv(broker)
	hook(bridge.TaskRequest{ID: "exec-1"}, "rt-1", env)

	cases := []struct {
		name     string
		settings daemonSettings
		code     contextcli.ErrorCode
		exit     int
	}{
		{"feature-off", daemonSettings{}, contextcli.CodeContextReadUnsupported, 8},
		// No runtime reports exec-1 as running, so the broker fails closed.
		{"run-not-active", daemonSettings{contextBroker: broker}, contextcli.CodePermissionDenied, 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
			c := contextIO{
				stdout: stdout, stderr: stderr,
				getenv:        func(k string) string { return env[k] },
				dial:          brokerDial(tc.settings),
				defaultSocket: func() (string, error) { return "", errors.New("unused") },
				executable:    func() (string, error) { return "/opt/riido/bin/riido", nil },
				lookPath:      func(string) (string, error) { return "", errors.New("unused") },
				cliVersion:    func() string { return "test" },
			}
			if got := exitCodeOf(t, runWork(c, []string{"get", "RIID-5900", "--json"})); got != tc.exit {
				t.Fatalf("want exit %d, got %d (%s)", tc.exit, got, stdout.String())
			}
			if reply := decodeEnvelope(t, stdout.Bytes()); reply.Error == nil || reply.Error.Code != tc.code {
				t.Fatalf("want %s, got %+v", tc.code, reply.Error)
			}
			if strings.Contains(stdout.String()+stderr.String(), env[contextcli.EnvRunToken]) {
				t.Fatal("run capability leaked to CLI output")
			}
		})
	}
}

func TestContextReadAuditLineHasNoSecrets(t *testing.T) {
	req := daemonRequest{Command: "work get", Args: map[string]string{"work": "RIID-5900\ninjected"}, RunCapability: "rrc_secret"}
	reply := contextcli.FailureEnvelope("cli_1", contextcli.Fail(contextcli.CodeControlPlaneUnsupported, "x", ""))
	line := contextReadAuditLine(req, reply)
	if strings.Contains(line, "rrc_secret") {
		t.Fatalf("audit line leaked the run capability: %q", line)
	}
	for _, want := range []string{`command="work get"`, "result=CONTROL_PLANE_UNSUPPORTED", `request_id="cli_1"`} {
		if !strings.Contains(line, want) {
			t.Fatalf("audit line missing %s: %q", want, line)
		}
	}
	if strings.ContainsRune(line, '\n') {
		t.Fatalf("audit line must stay on one line: %q", line)
	}
}

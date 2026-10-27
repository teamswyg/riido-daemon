package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestContextCLIUserModeUnimplementedReadsAreExplicitlyUnavailable(t *testing.T) {
	h := newContextHarness(t, nil)
	err := runWork(h.io, []string{"tree", "RIID-5839", "--json"})
	if got := exitCodeOf(t, err); got != 8 {
		t.Fatalf("want exit 8, got %d", got)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	if env.Error == nil || env.Error.Code != contextcli.CodeUserModeUnavailable || env.Error.NextAction == "" {
		t.Fatalf("want USER_MODE_UNAVAILABLE with next action, got %+v", env.Error)
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("user mode must not call the daemon broker")
	}
}

func TestContextCLIAgentModeOldDaemonMapsToUnsupported(t *testing.T) {
	h := newContextHarness(t, agentVars())
	err := runWork(h.io, []string{"get", "riid-5839", "--json"})
	if got := exitCodeOf(t, err); got != 8 {
		t.Fatalf("want exit 8, got %d", got)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	if env.Error == nil || env.Error.Code != contextcli.CodeContextReadUnsupported {
		t.Fatalf("want CONTEXT_READ_UNSUPPORTED, got %+v", env.Error)
	}
	req := h.daemon.lastRequest(t)
	args, _ := req["args"].(map[string]any)
	if req["method"] != contextcli.IPCMethodContextRead || req["command"] != "work get" || args["work"] != "RIID-5839" {
		t.Fatalf("unexpected IPC request: %v", req)
	}
	if req["run_capability"] != testRunCapability || req["workspace_id"] != "ws-1" {
		t.Fatalf("run capability and scope must reach the broker: %v", req)
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), testRunCapability) {
		t.Fatal("run capability leaked to output")
	}
}

func TestContextCLIAgentModeV1ReplyIsPassedThrough(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"daemon-1","data":{"snapshot_id":"snap-1"},"sources":[{"id":"RIID-5839","kind":"work","status":"available","truncated":false}],"warnings":[],"page":null,"error":null}`
	if err := runContext(h.io, []string{"get", "--work", "RIID-5839", "--budget", "4000", "--json"}); err != nil {
		t.Fatalf("want success, got %v", err)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	if env.Error != nil || len(env.Sources) != 1 || env.Sources[0].Status != contextcli.SourceAvailable {
		t.Fatalf("unexpected envelope: %+v", env)
	}
	args, _ := h.daemon.lastRequest(t)["args"].(map[string]any)
	if args["budget_chars"] != "4000" || args["limit"] != nil || args["team"] != nil {
		t.Fatalf("payload not forwarded: %v", args)
	}
}

func TestContextCLIDaemonErrorReplyKeepsExitCode(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"daemon-1","data":null,"error":{"code":"SNAPSHOT_EXPIRED","message":"expired","next_action":"riido context get","retryable":false,"request_id":"daemon-1"}}`
	err := runContext(h.io, []string{"get"})
	if got := exitCodeOf(t, err); got != 5 {
		t.Fatalf("want exit 5, got %d", got)
	}
	if !strings.Contains(h.stderr.String(), "riido context get") {
		t.Fatalf("text mode must show next action, got %q", h.stderr.String())
	}
}

func TestContextCLIDaemonUnreachable(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.down = true
	err := runWhoami(h.io, []string{"--json"})
	if got := exitCodeOf(t, err); got != 6 {
		t.Fatalf("want exit 6, got %d", got)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	if env.Error == nil || env.Error.Code != contextcli.CodeDaemonUnreachable || !env.Error.Retryable {
		t.Fatalf("want retryable DAEMON_UNREACHABLE, got %+v", env.Error)
	}
}

func TestContextCLIUsageAndSecretErrors(t *testing.T) {
	cases := []struct {
		name string
		run  func(contextIO) error
		code contextcli.ErrorCode
	}{
		{"secret", func(c contextIO) error { return runWork(c, []string{"get", "RIID-1", "--token", "x", "--json"}) }, contextcli.CodeSecretInArgv},
		{"secret-inline", func(c contextIO) error { return runAuth(c, []string{"login", "--token=x", "--json"}) }, contextcli.CodeSecretInArgv},
		{"bad-ref", func(c contextIO) error { return runWork(c, []string{"get", "../x", "--json"}) }, contextcli.CodeUsage},
		{"missing-query", func(c contextIO) error { return runWork(c, []string{"search", "--json"}) }, contextcli.CodeUsage},
		{"unknown-sub", func(c contextIO) error { return runWork(c, []string{"delete", "--json"}) }, contextcli.CodeUsage},
		{"bad-budget", func(c contextIO) error {
			return runContext(c, []string{"get", "--work", "RIID-1", "--budget", "-1", "--json"})
		}, contextcli.CodeUsage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, agentVars())
			if got := exitCodeOf(t, tc.run(h.io)); got != 2 {
				t.Fatalf("want exit 2, got %d", got)
			}
			if env := decodeEnvelope(t, h.stdout.Bytes()); env.Error == nil || env.Error.Code != tc.code {
				t.Fatalf("want %s, got %+v", tc.code, env.Error)
			}
			if len(h.daemon.requests) != 0 {
				t.Fatal("invalid input must not reach the daemon")
			}
		})
	}
}

func TestContextCLIFlagsCannotLeaveRunWorkspace(t *testing.T) {
	h := newContextHarness(t, agentVars())
	err := runWork(h.io, []string{"get", "RIID-1", "--workspace", "ws-other", "--json"})
	if got := exitCodeOf(t, err); got != 4 {
		t.Fatalf("want exit 4, got %d", got)
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("out-of-scope request must not reach the daemon")
	}
}

func TestAuthLoginNeverAcceptsSecretsAndIsGatedByMode(t *testing.T) {
	user := newContextHarness(t, nil)
	if got := exitCodeOf(t, runAuth(user.io, []string{"login", "--with-token", "--json"})); got != 8 {
		t.Fatalf("user mode login: want exit 8, got %d", got)
	}
	agent := newContextHarness(t, agentVars())
	if got := exitCodeOf(t, runAuth(agent.io, []string{"login", "--browser", "--json"})); got != 4 {
		t.Fatalf("agent mode login: want exit 4, got %d", got)
	}
	both := newContextHarness(t, nil)
	if got := exitCodeOf(t, runAuth(both.io, []string{"login", "--browser", "--desktop"})); got != 2 {
		t.Fatalf("two login methods: want exit 2, got %d", got)
	}
}

func TestAuthStatusNeverPrintsCredentials(t *testing.T) {
	h := newContextHarness(t, agentVars())
	if err := runAuth(h.io, []string{"status", "--json"}); err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if strings.Contains(h.stdout.String(), testRunCapability) {
		t.Fatal("auth status printed the run capability")
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	data, _ := env.Data.(map[string]any)
	if data["mode"] != "agent" || data["run_capability_present"] != true || data["user_credential"] != "unsupported" {
		t.Fatalf("unexpected auth status: %v", data)
	}
}

func TestCapabilitiesListsEveryWorkContextCommand(t *testing.T) {
	h := newContextHarness(t, nil)
	if err := runCapabilities(h.io, []string{"--json"}); err != nil {
		t.Fatalf("capabilities: %v", err)
	}
	out := h.stdout.String()
	for _, name := range []string{"work get", "work search", "context get", "context changes", "agents list", "agents show", "auth login", "whoami", "doctor"} {
		if !strings.Contains(out, `"name": "`+name+`"`) {
			t.Fatalf("capabilities missing %q:\n%s", name, out)
		}
	}
}

func TestContextCommandHelpHasNoSideEffects(t *testing.T) {
	for _, args := range [][]string{{"work", "--help"}, {"work", "get", "--help"}, {"context", "-h"}, {"auth", "--help"}, {"doctor", "--help"}} {
		h := newContextHarness(t, agentVars())
		var err error
		switch args[0] {
		case "work":
			err = runWork(h.io, args[1:])
		case "context":
			err = runContext(h.io, args[1:])
		case "auth":
			err = runAuth(h.io, args[1:])
		default:
			err = runDoctor(h.io, args[1:])
		}
		if err != nil || h.stdout.Len() != 0 || len(h.daemon.requests) != 0 {
			t.Fatalf("%v: help must only print usage to stderr (err=%v stdout=%q)", args, err, h.stdout.String())
		}
	}
}

func TestRootHelpKeepsLegacyAndWorkContextGroups(t *testing.T) {
	var buf bytes.Buffer
	printContextUsage(&buf)
	for _, token := range []string{"riido work <get|tree|relations|comments|replies|activity|search>", "riido context get", "riido context changes", "riido auth <login|logout|status>", "riido whoami", "riido doctor", "riido capabilities"} {
		if !strings.Contains(buf.String(), token) {
			t.Fatalf("usage missing %q", token)
		}
	}
}

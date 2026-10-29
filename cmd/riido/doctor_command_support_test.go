package main

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestDoctorRequestedRemoteCheckCannotReportSuccess(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[string(daemonMethodStatus)] = `{"daemon_version":"test"}`
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","data":{"broker":"daemon","control_plane":"assignment_snapshot"}}`
	if got := exitCodeOf(t, runDoctor(h.io, []string{"--check-api", "--json"})); got != 8 {
		t.Fatalf("unimplemented remote check must exit 8, got %d", got)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	data := env.Data.(map[string]any)
	if data["healthy"] != false || data["api_verified"] != false || doctorChecks(t, env)["api"] != "unsupported" {
		t.Fatalf("remote check claimed success: %+v", data)
	}
	for _, req := range h.daemon.requests {
		if command, _ := req["command"].(string); command != "" && command != "capabilities" {
			t.Fatalf("doctor must not consume context-read quota: %s", command)
		}
	}
}

func TestDoctorContextProbeDoesNotHideErrors(t *testing.T) {
	for _, code := range []contextcli.ErrorCode{contextcli.CodePermissionDenied, contextcli.CodeLeaseLost, contextcli.CodeUnavailable} {
		h := newContextHarness(t, agentVars())
		h.daemon.replies[string(daemonMethodStatus)] = `{"daemon_version":"test"}`
		body, err := json.Marshal(contextcli.FailureEnvelope("probe", contextcli.Fail(code, "probe rejected", "")))
		if err != nil {
			t.Fatal(err)
		}
		h.daemon.replies[contextcli.IPCMethodContextRead] = string(body)
		if got := exitCodeOf(t, runDoctor(h.io, []string{"--json"})); got != 6 {
			t.Fatalf("probe %s must fail local diagnosis, exit=%d", code, got)
		}
		if checks := doctorChecks(t, decodeEnvelope(t, h.stdout.Bytes())); checks["context_read"] != "fail" {
			t.Fatalf("probe %s reported success: %+v", code, checks)
		}
	}
}

func TestDoctorValidLocalProbeDoesNotVerifyRemoteAPI(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[string(daemonMethodStatus)] = `{"daemon_version":"test"}`
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","data":{"broker":"daemon","control_plane":"assignment_snapshot"}}`
	if err := runDoctor(h.io, []string{"--json"}); err != nil {
		t.Fatal(err)
	}
	env := decodeEnvelope(t, h.stdout.Bytes())
	data := env.Data.(map[string]any)
	if data["api_verified"] != false || data["healthy"] != true || doctorChecks(t, env)["api"] != "skip" {
		t.Fatalf("local probe must leave API unverified: %+v", data)
	}
}

func TestDoctorMalformedCapabilityReplyIsNotHealthy(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[string(daemonMethodStatus)] = `{"daemon_version":"test"}`
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","data":null}`
	if got := exitCodeOf(t, runDoctor(h.io, []string{"--json"})); got != 6 {
		t.Fatalf("empty capability response must fail diagnosis, exit=%d", got)
	}
}

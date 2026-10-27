package contextcli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRunTokenSelectsAgentMode(t *testing.T) {
	vars := map[string]string{EnvRunToken: "cap-secret", EnvTaskID: "RIID-1", EnvDaemonSocket: "/tmp/agentd.sock"}
	env := DetectEnvironment(func(k string) string { return vars[k] })
	if env.Mode != ModeAgent || !env.HasRunCapability() {
		t.Fatalf("want agent mode, got %+v", env)
	}
	out, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "cap-secret") {
		t.Fatalf("environment JSON leaked the run capability: %s", out)
	}
	user := DetectEnvironment(func(string) string { return "" })
	if user.Mode != ModeUser || user.HasRunCapability() {
		t.Fatalf("want user mode, got %+v", user)
	}
}

func TestErrorCodesMapToDocumentedExitCodes(t *testing.T) {
	want := map[ErrorCode]ExitCode{
		CodeUsage: 2, CodeSecretInArgv: 2, CodeAuthRequired: 3, CodePermissionDenied: 4,
		CodeOutOfRunScope: 4, CodeNotFound: 5, CodeSnapshotExpired: 5, CodeUnavailable: 6,
		CodeDaemonUnreachable: 6, CodeRateLimited: 7, CodeUnsupported: 8,
		CodeContextReadUnsupported: 8, CodeUserModeUnavailable: 8, CodeControlPlaneUnsupported: 8, CodeIncomplete: 9,
		CodeConflict: 10, CodeLeaseLost: 10, ErrorCode("SOMETHING_NEW"): 8,
	}
	for code, exit := range want {
		if got := code.ExitCode(); got != exit {
			t.Fatalf("%s: want exit %d, got %d", code, exit, got)
		}
	}
	if !Fail(CodeDaemonUnreachable, "x", "").Retryable() || Fail(CodeUsage, "x", "").Retryable() {
		t.Fatal("retryable mapping is wrong")
	}
}

func TestEnvelopeSlicesNeverNull(t *testing.T) {
	out, err := json.Marshal(FailureEnvelope("cli_1", Fail(CodeUsage, "bad", "fix")))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"schema_version":"1"`, `"sources":[]`, `"warnings":[]`, `"code":"USAGE"`, `"request_id":"cli_1"`} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("envelope missing %s: %s", want, out)
		}
	}
}

func TestParseWorkRef(t *testing.T) {
	cases := map[string]WorkRef{
		"RIID-5839":                            {Value: "RIID-5839", Kind: WorkRefKey},
		"riid-12":                              {Value: "RIID-12", Kind: WorkRefKey},
		"0f8b2c1e-6a1d-4b7e-9c3f-2d5e8a7b6c4d": {Value: "0f8b2c1e-6a1d-4b7e-9c3f-2d5e8a7b6c4d", Kind: WorkRefID},
	}
	for in, want := range cases {
		got, err := ParseWorkRef(in)
		if err != nil || got != want {
			t.Fatalf("%q: want %+v, got %+v err=%v", in, want, got, err)
		}
	}
	for _, bad := range []string{"", "RIID 1", "a/b", "../etc"} {
		if _, err := ParseWorkRef(bad); err == nil {
			t.Fatalf("%q: want usage error", bad)
		}
	}
}

func TestDecodeContextReadReply(t *testing.T) {
	_, err := DecodeContextReadReply([]byte(`{"schema_version":"1","error":"unknown method","method":"context_read"}`))
	if err == nil || err.Code != CodeContextReadUnsupported || err.ExitCode() != ExitUnsupported {
		t.Fatalf("old daemon: want CONTEXT_READ_UNSUPPORTED, got %v", err)
	}
	if _, err := DecodeContextReadReply([]byte(`not json`)); err == nil || err.Code != CodeUnsupported {
		t.Fatalf("garbage: want UNSUPPORTED, got %v", err)
	}
	env, err := DecodeContextReadReply([]byte(`{"schema_version":"1","request_id":"r","data":{"title":"x"},"error":null}`))
	if err != nil || env.Sources == nil || env.Warnings == nil {
		t.Fatalf("v1 reply: env=%+v err=%v", env, err)
	}
	failed, err := DecodeContextReadReply([]byte(`{"schema_version":"1","request_id":"r","error":{"code":"OUT_OF_RUN_SCOPE","message":"no"}}`))
	if err != nil {
		t.Fatalf("v1 failure reply: %v", err)
	}
	if replyErr := ReplyError(failed); replyErr == nil || replyErr.ExitCode() != ExitPermission {
		t.Fatalf("v1 failure: want exit 4, got %v", replyErr)
	}
}

func TestContextReadRequestCarriesCapabilityOnlyInIPCBody(t *testing.T) {
	env := DetectEnvironment(func(k string) string {
		if k == EnvRunToken {
			return "cap-secret"
		}
		return ""
	})
	req := NewContextReadRequest(env, "cli_1", "work get", map[string]string{"work": "RIID-1"})
	if req.Method != IPCMethodContextRead || req.RunCapability != "cap-secret" {
		t.Fatalf("request: %+v", req)
	}
}

func TestReplyErrorKeepsPermanentRunLimit(t *testing.T) {
	for _, code := range []ErrorCode{CodeRateLimited, CodeUnavailable} {
		for _, retryable := range []bool{true, false} {
			env := NewEnvelope("request", nil)
			env.Error = &Error{Code: code, Message: "limit", Retryable: retryable, NextAction: "new execution", RequestID: "request"}
			reply := ReplyError(env)
			if reply.Retryable() != retryable || FailureEnvelope("request", reply).Error.Retryable != retryable {
				t.Fatalf("reply lost retry policy for %s: want %v", code, retryable)
			}
		}
	}
	if !Fail(CodeRateLimited, "temporary limit", "later").Retryable() || !Fail(CodeUnavailable, "temporary", "later").Retryable() {
		t.Fatal("ordinary transient failures must keep their retry policy")
	}
}

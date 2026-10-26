package main

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestContextCLINativeIPCRoundTrip(t *testing.T) {
	f := newNativeContextFixture(t)
	env, code := f.run(t, nil, "whoami")
	if code != 0 || env.Error != nil {
		t.Fatal("active capability must authorize local whoami")
	}
	cases := []struct {
		view string
		args []string
	}{
		{"base", []string{"context", "get"}},
		{"task", []string{"work", "get", "RIID-5900"}},
		{"tree", []string{"work", "tree", "RIID-5900"}},
		{"resources", []string{"work", "relations", "RIID-5900"}},
		{"comments", []string{"work", "comments", "RIID-5900"}},
		{"agents", []string{"agents", "list"}},
		{"agent", []string{"agents", "show", "synthetic-agent"}},
	}
	for _, tc := range cases {
		t.Run(tc.view, func(t *testing.T) {
			env, code := f.run(t, nil, tc.args...)
			if code != 0 || env.Error != nil || len(env.Sources) != 1 || env.Sources[0].Version != "fixture-v1" {
				t.Fatal("native CLI source envelope was not preserved")
			}
			encoded, err := json.Marshal(env.Data)
			if err != nil {
				t.Fatal(err)
			}
			var body struct{ Marker, View string }
			if json.Unmarshal(encoded, &body) != nil || body.Marker != "NATIVE_CONTEXT_FIXTURE" || body.View != tc.view {
				t.Fatal("native CLI returned a different read view")
			}
		})
	}
	expectDenied := func(env contextcli.Envelope, code int, want contextcli.ErrorCode) {
		t.Helper()
		if code != 4 || env.Error == nil || env.Error.Code != want || env.Data != nil || len(env.Sources) != 0 {
			t.Fatalf("want exit4/%s without data or sources", want)
		}
	}
	env, code = f.run(t, nil, "context", "get", "--workspace", "other-ws")
	expectDenied(env, code, contextcli.CodeOutOfRunScope) // OUT_OF_RUN_SCOPE in the CLI.
	env, code = f.run(t, map[string]string{contextcli.EnvTaskID: "RIID-9999"}, "context", "get")
	expectDenied(env, code, contextcli.CodeOutOfRunScope) // Broker binding mismatch.
	env, code = f.run(t, map[string]string{contextcli.EnvRunToken: "synthetic-unknown-capability"}, "whoami")
	expectDenied(env, code, contextcli.CodePermissionDenied)
	f.certain.Store(false)
	env, code = f.run(t, nil, "whoami")
	if code != 6 || env.Error == nil || env.Error.Code != contextcli.CodeUnavailable {
		t.Fatal("uncertain activity must fail retryably")
	}
	if _, ok := f.broker.Registry.Lookup(f.token); !ok {
		t.Fatal("uncertain activity revoked capability")
	}
	f.certain.Store(true)
	f.active.Store(false)
	for _, args := range [][]string{{"whoami"}, {"context", "get"}} {
		env, code = f.run(t, nil, args...)
		expectDenied(env, code, contextcli.CodePermissionDenied) // Same capability: PERMISSION_DENIED.
	}
	if _, ok := f.broker.Registry.Lookup(f.token); ok {
		t.Fatal("ended capability was not revoked")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.queries) != len(cases) {
		t.Fatal("refused calls must not reach control plane")
	}
	for i, q := range f.queries {
		if q.View != cases[i].view || q.AssignmentID != "synthetic-assignment" || q.LeaseToken != "synthetic-lease-private" {
			t.Fatal("broker forwarded incorrect scope or view")
		}
	}
	for _, line := range f.audit {
		f.assertNoSecrets(t, line)
	}
}

package saasplane

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCoveredDeviceRequiresNegotiatedGeneration(t *testing.T) {
	p, err := New(Config{BaseURL: "http://127.0.0.1:1", DaemonID: "daemon", DeviceID: "dcov_fixture", DeviceSecret: "fixture-secret"})
	if p != nil {
		defer p.Close()
	}
	if err == nil {
		t.Fatal("covered device accepted a missing server generation")
	}
}

func TestCoverageConfigurationRejectsMixedAndMalformedMetadata(t *testing.T) {
	for _, tc := range []struct{ name, device, secret, generation string }{
		{"legacy", "legacy", "secret", "cov_fixture"},
		{"no_secret", "dcov_fixture", "", "cov_fixture"},
		{"space", "dcov_fixture", "secret", " cov_fixture"},
		{"newline", "dcov_fixture", "secret", "cov_fixture\n"},
		{"internal_space", "dcov_fixture", "secret", "cov fixture"},
		{"slash", "dcov_fixture", "secret", "cov/fixture"},
		{"oversize", "dcov_fixture", "secret", strings.Repeat("a", 513)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateExecutionCoverage(tc.device, tc.secret, tc.generation); err == nil {
				t.Fatal("invalid enrollment metadata accepted")
			}
		})
	}
	if err := ValidateExecutionCoverage("legacy", "secret", ""); err != nil {
		t.Fatal("legacy device was rejected")
	}
}

func TestCoverageHeaderUsesExactServerGenerationWithoutFallback(t *testing.T) {
	for _, covered := range []bool{false, true} {
		t.Run(map[bool]string{true: "covered", false: "legacy"}[covered], func(t *testing.T) {
			generation, device := "", "legacy"
			if covered {
				generation, device = "cov_fixture.Exact-Case_1", "dcov_fixture"
			}
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				values := r.Header.Values("X-Riido-Execution-Coverage")
				if (covered && (len(values) != 1 || values[0] != generation)) || (!covered && len(values) != 0) {
					t.Error("request changed or omitted enrollment generation")
				}
				if r.Header.Get("X-Riido-Device-Id") != device || r.Header.Get("X-Riido-Device-Secret") != "fixture-secret" {
					t.Error("device authentication missing")
				}
				w.WriteHeader(http.StatusForbidden)
			}))
			defer server.Close()
			p, err := New(Config{BaseURL: server.URL, DaemonID: "daemon", DeviceID: device, DeviceSecret: "fixture-secret", ExecutionCoverageGeneration: generation})
			if err != nil {
				t.Fatal(err)
			}
			defer p.Close()
			paths := []string{"/v1/agents/agent/poll", "/v1/agents/agent/events", "/v1/agents/agent/heartbeat", "/v1/daemon/runtime-snapshot"}
			for _, path := range paths {
				if err := p.doJSON(context.Background(), http.MethodPost, path, []byte(`{}`), nil); err == nil {
					t.Fatal("policy denial was ignored")
				}
			}
			if calls != len(paths) {
				t.Fatal("policy rejection retried or fell back to an untracked request")
			}
		})
	}
}

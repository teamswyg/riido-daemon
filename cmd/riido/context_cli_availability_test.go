package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestContextCLIUnsupportedReadsNeverReachDaemon(t *testing.T) {
	for _, test := range []struct {
		name string
		run  func(contextIO) error
	}{
		{"changes", func(c contextIO) error { return runContext(c, []string{"changes", "--since", "snapshot", "--json"}) }},
		{"search", func(c contextIO) error { return runWork(c, []string{"search", "--query", "planning", "--json"}) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, vars := range []map[string]string{nil, agentVars()} {
				h := newContextHarness(t, vars)
				if got := exitCodeOf(t, test.run(h.io)); got != 8 {
					t.Fatalf("unsupported read exit=%d", got)
				}
				env := decodeEnvelope(t, h.stdout.Bytes())
				if env.Error == nil || env.Error.Code != contextcli.CodeControlPlaneUnsupported || env.Data != nil || len(h.daemon.requests) != 0 {
					t.Fatalf("unsupported read must fail locally without data: %+v", env)
				}
			}
		})
	}
}

func TestContextCLIUnsupportedFiltersNeverReachDaemon(t *testing.T) {
	for _, flag := range []string{"--limit=20", "--team=RIID", "--profile=default"} {
		h := newContextHarness(t, agentVars())
		if got := exitCodeOf(t, runWork(h.io, []string{"comments", "RIID-1", flag, "--json"})); got != 8 {
			t.Fatalf("%s: want exit 8, got %d", flag, got)
		}
		if len(h.daemon.requests) != 0 || !strings.Contains(h.stdout.String(), strings.Split(flag, "=")[0]) {
			t.Fatal("unsupported filter must be explained before IPC")
		}
	}
}

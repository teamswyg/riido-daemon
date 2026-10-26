package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestContextReadAuditIdentifiesRequestedTarget(t *testing.T) {
	for _, tc := range []struct {
		command string
		args    map[string]string
		want    string
	}{
		{"agents show", map[string]string{"agent": "agent-5900"}, `target="agent-5900"`},
		{"meeting get", map[string]string{"meeting": "meeting-5900"}, `target="meeting-5900"`},
		{"meeting list", map[string]string{"work": "RIID-5900"}, `target="RIID-5900"`},
		{"work get", map[string]string{"work": "RIID-5900"}, `target="RIID-5900"`},
		{"context get", map[string]string{"work": "RIID-5900"}, `target="RIID-5900"`},
		{"agents list", nil, `target="-"`},
	} {
		t.Run(tc.command, func(t *testing.T) {
			line := contextReadAuditLine(daemonRequest{Command: tc.command, Args: tc.args}, contextcli.Envelope{RequestID: "cli_audit"})
			if !strings.Contains(line, tc.want) {
				t.Fatalf("audit did not identify requested target: %s", line)
			}
		})
	}
}

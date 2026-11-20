package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestAgentExplicitReadsUseOnlyRunIPC(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		run                    func(contextIO) error
		command, option, value string
	}{
		{"activity", func(c contextIO) error {
			return runWork(c, []string{"activity", "RAX-43", "--page", "2", "--limit", "10", "--json"})
		}, "work activity", "page", "2"},
		{"comments", func(c contextIO) error {
			return runWork(c, []string{"comments", "RAX-43", "--page", "2", "--version", "version-one", "--json"})
		}, "work comments", "version", "version-one"},
		{"body", func(c contextIO) error {
			return runWork(c, []string{"body", "RAX-43", "--part", "metadata", "--version", "version-one", "--offset", "16384", "--json"})
		}, "work body", "offset", "16384"},
		{"threads", func(c contextIO) error { return runThread(c, []string{"list", "--work", "RAX-43", "--json"}) }, "thread list", "work", "RAX-43"},
		{"thread", func(c contextIO) error {
			return runThread(c, []string{"show", "thread-one", "--work", "RAX-43", "--json"})
		}, "thread show", "thread", "thread-one"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, agentVars())
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Fatal("agent opened user auth")
				return nil, nil
			}
			h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"test","data":{"live":true},"sources":[],"warnings":[],"page":null,"error":null}`
			if err := tc.run(h.io); err != nil {
				t.Fatal(err)
			}
			req := h.daemon.lastRequest(t)
			args := req["args"].(map[string]any)
			if req["command"] != tc.command || args[tc.option] != tc.value || req["run_capability"] != testRunCapability {
				t.Fatalf("wrong IPC: %v", req)
			}
			if strings.Contains(h.stdout.String()+h.stderr.String(), testRunCapability) {
				t.Fatal("capability exposed")
			}
		})
	}
}

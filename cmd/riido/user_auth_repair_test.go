package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestAuthRepairRejectsAgentBeforeOpeningUserStorage(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("agent opened user credentials")
		return nil, nil
	}
	if got := exitCodeOf(t, runAuth(h.io, []string{"repair", "--json"})); got != 4 {
		t.Fatalf("exit=%d", got)
	}
	if strings.Contains(h.stdout.String(), testRunCapability) || len(h.daemon.requests) != 0 {
		t.Fatal("agent repair leaked or contacted daemon")
	}
}

func TestAuthRepairHelpAndInvalidArgsHaveNoSideEffects(t *testing.T) {
	for _, args := range [][]string{{"repair", "--help"}, {"repair", "--with-token", "synthetic-secret", "--json"}} {
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("help/invalid opened user credentials")
			return nil, nil
		}
		_ = runAuth(h.io, args)
		if len(h.daemon.requests) != 0 {
			t.Fatal("help/invalid contacted daemon")
		}
	}
}

package main

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestLoginConfirmationFailurePreservesCLIExitAndStopsOAuth(t *testing.T) {
	for _, scenario := range []struct {
		name string
		err  error
		exit int
	}{
		{"user-cancel", userauth.ErrStoreLocked, 3},
		{"deadline", context.DeadlineExceeded, 6},
		{"unavailable-helper", securecredential.ErrUnavailable, 8},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			store := &cliMemoryStore{confirmErr: scenario.err}
			service, calls, opened := testUserService(t, store)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
			if got := exitCodeOf(t, runAuth(h.io, []string{"login", "--json"})); got != scenario.exit {
				t.Fatalf("exit=%d want=%d", got, scenario.exit)
			}
			if store.confirmCalls != 1 || store.legacyConfirms != 0 || *opened != 0 || *calls != 0 || store.present {
				t.Fatal("failed confirmation reached OAuth or storage")
			}
		})
	}
}

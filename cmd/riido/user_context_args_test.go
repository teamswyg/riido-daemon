package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserContextRequiresExplicitWorkAndTeamBeforeOpeningCredentials(t *testing.T) {
	for _, flags := range [][]string{
		{},
		{"--work", "work-one"},
		{"--team-id", "team-one"},
		{"--work", "", "--team-id", "team-one"},
		{"--work", "work-one", "--team-id", " "},
	} {
		t.Run(joinContextTestFlags(flags), func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Fatal("invalid input opened protected credentials")
				return nil, nil
			}
			err := runContext(h.io, append([]string{"get", "--json"}, flags...))
			if exit := exitCodeOf(t, err); exit != 2 {
				t.Fatalf("missing work/team exit = %d, want usage exit 2", exit)
			}
			if len(h.daemon.requests) != 0 {
				t.Fatal("user scope request reached daemon")
			}
		})
	}
}

func joinContextTestFlags(flags []string) string {
	if len(flags) == 0 {
		return "no flags"
	}
	return flags[0] + "=" + flags[1]
}

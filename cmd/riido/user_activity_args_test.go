package main

import (
	"slices"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserActivityRejectsArgumentsBeforeStorage(t *testing.T) {
	cases := []struct {
		name string
		args []string
		exit int
	}{
		{"missing-work", []string{"activity", "--team-id", "team-one"}, 2},
		{"missing-team", []string{"activity", "work-one"}, 2},
		{"key", []string{"activity", "RIID-5900", "--team-id", "team-one"}, 8},
		{"team-key", []string{"activity", "work-one", "--team-id", "RIID-5900"}, 8},
		{"empty-team", []string{"activity", "work-one", "--team-id="}, 2},
	}
	for _, flag := range []string{"--cursor=", "--workspace=", "--profile=default", "--team=team-one"} {
		cases = append(cases, struct {
			name string
			args []string
			exit int
		}{flag, append(slices.Clone(activityArgs()), flag), 8})
	}
	for _, flag := range []string{"--page=", "--page=0", "--page=-1", "--page=+1", "--page=01", "--page= 1", "--page=1.0", "--page=2147483648", "--page=2147483647", "--limit=", "--limit=0", "--limit=-1", "--limit=+1", "--limit=01", "--limit=101", "--limit=1.5", "--limit=2147483648"} {
		cases = append(cases, struct {
			name string
			args []string
			exit int
		}{flag, append(slices.Clone(activityArgs()), flag), 2})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Fatal("invalid activity opened storage")
				return nil, nil
			}
			if got := exitCodeOf(t, runWork(h.io, append(tc.args, "--json"))); got != tc.exit {
				t.Fatalf("exit=%d want=%d", got, tc.exit)
			}
			if len(h.daemon.requests) != 0 {
				t.Fatal("invalid user activity used daemon")
			}
		})
	}
}

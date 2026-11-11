package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestDesktopBootstrapRejectsArgvAndUnboundedInput(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		input string
	}{
		{[]string{"request", "--json", "not-a-request"}, "{}"},
		{[]string{"request", "--json"}, ""},
		{[]string{"request", "--json"}, strings.Repeat("x", 4097)},
		{[]string{"restart", "--json", "1800000000"}, "{}"},
		{[]string{"restart", "--json"}, ""},
		{[]string{"identity"}, ""},
		{[]string{"other", "--json"}, ""},
	} {
		h := newContextHarness(t, nil)
		h.io.stdin = strings.NewReader(tc.input)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
			t.Fatal("invalid Desktop request opened credentials")
			return nil, nil
		}
		if err := h.io.runDesktopBootstrap(tc.args); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
}

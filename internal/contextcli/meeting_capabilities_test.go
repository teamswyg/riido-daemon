package contextcli

import (
	"slices"
	"testing"
)

func TestMeetingCapabilitiesRequireAgentBroker(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status CommandStatus
		modes  []Mode
	}{
		{"meeting list", CommandNeedsDaemon, []Mode{ModeAgent}},
		{"meeting get", CommandAvailable, []Mode{ModeAgent, ModeUser}},
		{"meeting body", CommandAvailable, []Mode{ModeUser}},
	} {
		found := false
		for _, command := range Capabilities("test").Commands {
			if command.Name != tc.name {
				continue
			}
			found = true
			if command.Status != tc.status || !slices.Equal(command.Modes, tc.modes) {
				t.Fatalf("meeting read advertised unavailable mode: %+v", command)
			}
		}
		if !found {
			t.Fatalf("missing capability %s", tc.name)
		}
	}
}

package contextcli

import "testing"

func TestMeetingCapabilitiesRequireAgentBroker(t *testing.T) {
	for _, name := range []string{"meeting list", "meeting get"} {
		found := false
		for _, command := range Capabilities("test").Commands {
			if command.Name != name {
				continue
			}
			found = true
			if command.Status != CommandNeedsDaemon || len(command.Modes) != 1 || command.Modes[0] != ModeAgent {
				t.Fatalf("meeting read advertised unavailable mode: %+v", command)
			}
		}
		if !found {
			t.Fatalf("missing capability %s", name)
		}
	}
}

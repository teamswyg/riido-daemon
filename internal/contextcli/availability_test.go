package contextcli

import "testing"

func TestUnsupportedCommandsAreNotAdvertisedAsBrokerReads(t *testing.T) {
	for _, name := range []string{"context changes", "work activity", "work search"} {
		found := false
		for _, command := range Capabilities("test").Commands {
			if command.Name == name {
				found = true
				if command.Status != CommandUnavailable || command.Note == "" {
					t.Fatalf("unsupported command advertised as supported: %+v", command)
				}
			}
		}
		if !found {
			t.Fatalf("missing explicit unavailable capability %s", name)
		}
	}
}

func TestIgnoredOptionsAreExplicitlyUnsupported(t *testing.T) {
	for _, args := range [][]string{{"--limit", "20"}, {"--limit=50"}, {"--team", "RIID"}, {"--profile=default"}} {
		_, err := ParseArgs(args, ArgSpec{})
		if err == nil || err.ExitCode() != ExitUnsupported || err.NextAction == "" {
			t.Fatalf("ignored options %v must return unsupported, got %v", args, err)
		}
	}
}

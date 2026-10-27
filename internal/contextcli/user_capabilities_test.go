package contextcli

import (
	"slices"
	"strings"
	"testing"
)

func TestStandaloneCapabilitiesMatchImplementedScope(t *testing.T) {
	commands := Capabilities("test").Commands
	for _, name := range []string{"auth login", "auth logout", "whoami", "work get"} {
		index := slices.IndexFunc(commands, func(command CommandCapability) bool { return command.Name == name })
		if index < 0 {
			t.Fatalf("missing %s", name)
		}
		command := commands[index]
		if command.Status != CommandAvailable || !slices.Contains(command.Modes, ModeUser) || !strings.Contains(command.Note, "서명") {
			t.Fatalf("missing guarded user support: %s", name)
		}
		if (name == "auth login" || name == "auth logout") && slices.Contains(command.Modes, ModeAgent) {
			t.Fatalf("agent must not alter user auth: %s", name)
		}
	}
}

package claude

import (
	"encoding/json"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"reflect"
	"testing"
)

func TestReferenceFoldersUseSessionSettingsAndKeepCwd(t *testing.T) {
	paths := []string{"/자료/project one", "/tmp/a $(echo nope)"}
	cmd, err := BuildStart(agentbridge.StartRequest{Cwd: "/task/workdir", ReferenceDirectories: paths}, StartOptions{PermissionMode: PermissionModeApproval})
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != "/task/workdir" {
		t.Fatal("cwd changed")
	}
	for i, arg := range cmd.Args {
		if arg != "--settings" {
			continue
		}
		var settings struct {
			Permissions struct {
				AdditionalDirectories []string `json:"additionalDirectories"`
			} `json:"permissions"`
		}
		if err := json.Unmarshal([]byte(cmd.Args[i+1]), &settings); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(settings.Permissions.AdditionalDirectories, paths) {
			t.Fatalf("settings: %#v", settings)
		}
		return
	}
	t.Fatal("missing session settings")
}

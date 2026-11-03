package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestWorkspaceLabelKeepsID(t *testing.T) {
	names := map[string]string{"ws-1": "스위그"}
	for _, c := range []struct{ id, want string }{
		{"ws-1", "스위그 (ws-1)"},
		{"ws-unknown", "ws-unknown"},
		{"", ""},
	} {
		if got := workspaceLabel(c.id, names, 80); got != c.want {
			t.Errorf("workspaceLabel(%q) = %q, want %q", c.id, got, c.want)
		}
	}
}

// Off-terminal whoami prints the unchanged JSON, so it must not spend extra
// requests resolving names nobody sees.
func TestWhoamiSkipsNameLookupOffTerminal(t *testing.T) {
	for _, args := range [][]string{nil, {"--json"}} {
		queries := []string{}
		service := commentService(t, func(w http.ResponseWriter, r *http.Request) {
			var body struct{ Query string }
			_ = json.NewDecoder(r.Body).Decode(&body)
			queries = append(queries, body.Query)
			fmt.Fprint(w, `{"data":{"getMyAccountInfo":{"id":"user-one","lastViewedWorkspaceId":"workspace-one"}}}`)
		})
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return service, nil }
		if err := runWhoami(h.io, args); err != nil {
			t.Fatal(err)
		}
		if len(queries) != 1 || !strings.Contains(queries[0], "getMyAccountInfo") {
			t.Fatalf("args=%v made extra requests: %d", args, len(queries))
		}
	}
}

// Review on #822: a long name on a narrow terminal is shortened, never the ID.
func TestWorkspaceLabelShortensOnlyTheName(t *testing.T) {
	id := "1IFsD2cWorkspaceID"
	names := map[string]string{id: strings.Repeat("아주긴워크스페이스", 10)}
	for _, room := range []int{12, 30, 60} {
		got := workspaceLabel(id, names, room)
		if !strings.HasSuffix(got, id+")") && got != id {
			t.Fatalf("room=%d lost the ID: %q", room, got)
		}
		if got != id && cliui.DisplayWidth(got) > room {
			t.Fatalf("room=%d label wider than room: %q", room, got)
		}
	}
	if got := workspaceLabel(id, names, 12); got != id {
		t.Fatalf("no room for a name must show the bare ID: %q", got)
	}
}

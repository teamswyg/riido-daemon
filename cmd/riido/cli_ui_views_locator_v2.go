package main

import (
	"fmt"
	"io"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func workBodyV2Command(id, workspace, kind string, team *string) string {
	command := fmt.Sprintf("riido work body %q --workspace %q --locator-version 2", id, workspace)
	if team != nil {
		command += fmt.Sprintf(" --team-id %q", *team)
	}
	if kind != "" && kind != "component" {
		command += fmt.Sprintf(" --kind %q", kind)
	}
	return command
}

func contextV2View(w io.Writer, scope userapi.ReadScopeV2) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) {
		byID := make(map[string]*userapi.ReadScopeWorkV2, len(scope.Works))
		for _, work := range scope.Works {
			byID[work.Kind+":"+work.ID] = work
		}
		renderContextView(out, term, scope.ContentView(), func(common *userapi.ReadScopeWork) string {
			work := byID[common.Kind+":"+common.ID]
			return workBodyV2Command(work.ID, work.WorkspaceID, work.Kind, work.TeamID)
		})
	})
}

func scopeCommentsV2View(w io.Writer, page userapi.ReadScopeCommentPageV2) textRenderer {
	return terminalView(w, func(out io.Writer, term cliui.Terminal) {
		renderScopeCommentsView(out, term, page.ContentView(), workBodyV2Command(page.ResourceID, page.WorkspaceID, page.ResourceKind, page.TeamID))
	})
}

package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

// whoamiWorkspacePages bounds the best-effort name lookup to a few pages.
const whoamiWorkspacePages = 3

// whoamiWorkspaceNames looks up display names for the terminal whoami view
// only. A failed or partial lookup leaves the ID as-is; the identity payload
// and JSON output never depend on it.
func whoamiWorkspaceNames(ctx context.Context, session *standalonecli.Service, ids ...string) map[string]string {
	want := map[string]bool{}
	for _, id := range ids {
		if id != "" {
			want[id] = true
		}
	}
	names := map[string]string{}
	q := userapi.WorkspaceListRequest{Limit: 100}
	for page := 0; page < whoamiWorkspacePages && len(names) < len(want); page++ {
		list, err := session.Workspaces(ctx, q)
		if err != nil || list == nil {
			return names
		}
		for _, item := range list.Items {
			if want[item.ID] && item.Name != "" {
				names[item.ID] = item.Name
			}
		}
		if list.Complete || list.NextCursor == nil {
			return names
		}
		next := *list.NextCursor
		q.Cursor = &next
	}
	return names
}

// workspaceLabel shows "name (id)" so the ID stays visible for copying.
func workspaceLabel(id string, names map[string]string) string {
	if name := names[id]; id != "" && name != "" {
		return name + " (" + id + ")"
	}
	return id
}

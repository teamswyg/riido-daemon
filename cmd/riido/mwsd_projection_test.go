package main

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/project"
)

func TestMwsdProjectionPrintsWorkspaceProjection(t *testing.T) {
	socketPath, stop := serveTestMwsd(t)
	defer stop()

	out := captureStdout(t, func() {
		if err := run([]string{"mwsd", "projection", "--socket", socketPath}); err != nil {
			t.Fatalf("run mwsd projection: %v", err)
		}
	})
	var projection project.WorkspaceProjection
	if err := json.Unmarshal([]byte(out), &projection); err != nil {
		t.Fatalf("parse projection output: %v\n%s", err, out)
	}
	if projection.SchemaVersion != "riido-workspace-projection.v1" {
		t.Fatalf("unexpected projection schema: %s", projection.SchemaVersion)
	}
	if len(projection.DocumentTaskLinks) != 1 || projection.DocumentTaskLinks[0].TaskID != "task:mws.cli" {
		t.Fatalf("unexpected projection task links: %#v", projection.DocumentTaskLinks)
	}
}
